package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ryanzola/dreampicai/db"
	"github.com/ryanzola/dreampicai/types"
)

const succeeded = "success"

type ReplicateResp struct {
	Status string   `json:"status"`
	Output []string `json:"output"`
}

func HandleReplicateCallback(w http.ResponseWriter, r *http.Request) error {
	var resp ReplicateResp
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return err
	}

	if resp.Status != succeeded {
		return fmt.Errorf("replicate callback responded with a non-ok status: %s", resp.Output)
	}

	fmt.Println(resp)
	batchID, err := uuid.Parse(chi.URLParam(r, "batchID"))
	if err != nil {
		return fmt.Errorf("replicate callback invalid batchID: %s", err)
	}

	images, err := db.GetImagesByBatchID(batchID)
	if err != nil {
		return fmt.Errorf("failed to find image with batchID %s: %s", batchID, err)
	}

	if len(images) != len(resp.Output) {
		return fmt.Errorf("replicate callback unequal images compared to replicate output")
	}

	for i, imageURL := range resp.Output {
		images[i].Status = types.ImageStatusCompleted
		images[i].ImageLocation = imageURL
	}

	return nil
}
