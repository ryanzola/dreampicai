package main

import (
	"embed"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/ryanzola/dreampicai/db"
	"github.com/ryanzola/dreampicai/handler"
	"github.com/ryanzola/dreampicai/pkg/sb"
)

//go:embed public
var FS embed.FS

func main() {
	if err := initEverything(); err != nil {
		log.Fatal(err)
	}

	router := chi.NewMux()
	router.Use(handler.WithUser)

	router.Handle("/*", public())
	router.Get("/", handler.Make(handler.HandleHomeIndex))
	router.Get("/long-process", handler.Make(handler.HandleLongProcess))

	router.Get("/login", handler.Make(handler.HandleLoginIndex))
	router.Get("/login/provider/google", handler.Make(handler.HandleLoginWithGoogle))
	router.Post("/login", handler.Make(handler.HandleLoginCreate))

	router.Post("/logout", handler.Make(handler.HandleLogoutCreate))

	router.Get("/auth/callback", handler.Make(handler.HandleAuthCallback))

	router.Post("/replicate/callback/{userID}/{batchID}", handler.Make(handler.HandleReplicateCallback))

	router.Group(func(auth chi.Router) {
		auth.Use(handler.WithAuth)

		auth.Get("/account/setup", handler.Make(handler.HandleAccountSetupIndex))
		auth.Post("/account/setup", handler.Make(handler.HandleAccountSetupCreate))
	})

	router.Group(func(auth chi.Router) {
		auth.Use(handler.WithAuth, handler.WithAccountSetup)

		auth.Get("/settings", handler.Make(handler.HandleSettingsIndex))
		auth.Put("/settings/account/profile", handler.Make(handler.HandleSettingsUsernameUpdate))

		auth.Get("/generate", handler.Make(handler.HandleGenerateIndex))
		auth.Post("/generate", handler.Make(handler.HandleGenerateCreate))
		auth.Get("/generate/image/status/{id}", handler.Make(handler.HandleGenerateImageStatus))
		auth.Get("/checkout/create/{priceID}", handler.Make(handler.HandleStripeCheckoutCreate))
		auth.Get("/checkout/success/{sessionID}", handler.Make(handler.HandleStripeCheckoutSuccess))
		auth.Get("/checkout/cancel/{priceID}", handler.Make(handler.HandleStripeCheckoutCancel))

		auth.Get("/buy-credits", handler.Make(handler.HandleCreditsIndex))
	})

	port := os.Getenv("HTTP_LISTEN_ADDR")
	slog.Info("Application is running on", "port", port)
	log.Fatal(http.ListenAndServe(port, router))
}

func initEverything() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	if err := db.Init(); err != nil {
		return err
	}

	return sb.Init()
}
