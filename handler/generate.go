package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/replicate/replicate-go"
	"github.com/ryanzola/dreampicai/db"
	"github.com/ryanzola/dreampicai/pkg/kit/validate"
	"github.com/ryanzola/dreampicai/types"
	"github.com/ryanzola/dreampicai/view/generate"
	"github.com/uptrace/bun"
)

const creditsPerImage = 2

func HandleGenerateIndex(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)
	images, err := db.GetImagesByUserID(user.ID)
	if err != nil {
		return err
	}

	data := generate.ViewData{
		Images: images,
	}

	return render(r, w, generate.Index(data))
}

func HandleGenerateCreate(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)
	amount, _ := strconv.Atoi(r.FormValue("amount"))

	params := generate.FormParams{
		Prompt: r.FormValue("prompt"),
		Amount: amount,
	}

	var errors generate.FormErrors

	if amount <= 0 || amount > 8 {
		errors.Amount = "Amount should be between 1 and 8"
	}

	ok := validate.New(params, validate.Fields{
		"Prompt": validate.Rules(validate.Min(10), validate.Max(500)),
	}).Validate(&errors)
	if !ok || len(errors.Amount) > 0 {
		return render(r, w, generate.Form(params, errors))
	}

	// check if user has enough credits
	creditsNeeded := params.Amount * creditsPerImage
	if user.Account.Credits < creditsNeeded {
		errors.CreditsNeeded = creditsNeeded
		errors.UserCredits = user.Account.Credits
		errors.Credits = true

		return render(r, w, generate.Form(params, errors))
	}

	user.Account.Credits -= creditsNeeded
	if err := db.UpdateAccount(&user.Account); err != nil {
		return err
	}

	batchID := uuid.New()
	genParams := GenerateImageParams{
		Prompt:  params.Prompt,
		Amount:  params.Amount,
		UserID:  user.ID,
		BatchID: batchID,
	}

	if err := generateImages(r.Context(), genParams); err != nil {
		log.Println("Error generating image", err)
		return err
	}

	// if there is an error, rollback the transaction
	// if there is no error, commit the transaction, create the image and return the response
	err := db.Bun.RunInTx(r.Context(), &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		for i := 0; i < params.Amount; i++ {
			img := types.Image{
				Prompt:  params.Prompt,
				UserID:  user.ID,
				Status:  types.ImageStatusPending,
				BatchID: batchID,
			}
			if err := db.CreateImage(tx, &img); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return hxRedirect(w, r, "/generate")
}

func HandleGenerateImageStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}

	// fetch from db
	image, err := db.GetImageByID(id)
	if err != nil {
		return err
	}

	slog.Info("Checking image status", "id", id)

	return render(r, w, generate.GalleryImage(image))
}

type GenerateImageParams struct {
	Prompt  string
	Amount  int
	UserID  uuid.UUID
	BatchID uuid.UUID
}

func generateImages(ctx context.Context, params GenerateImageParams) error {
	// You can also provide a token directly with
	// `replicate.NewClient(replicate.WithToken("r8_..."))`
	r8, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	// stability-ai/stable-diffusion:ac732df83cea7fff18b8472768c88ad041fa750ff7682a21affe81863cbe77e4
	// owner := "stability-ai/stable-diffusion"
	version := "ac732df83cea7fff18b8472768c88ad041fa750ff7682a21affe81863cbe77e4"

	input := replicate.PredictionInput{
		"prompt":      params.Prompt,
		"num_outputs": params.Amount,
	}

	baseURL := os.Getenv("REPLICATE_CALLBACK_URL")
	URL := fmt.Sprintf("%s/%s/%s", baseURL, params.UserID, params.BatchID)

	webhook := replicate.Webhook{
		URL:    URL,
		Events: []replicate.WebhookEventType{"completed"},
	}

	// Run a model and wait for its output
	_, err = r8.CreatePrediction(ctx, version, input, &webhook, false)

	return err
}
