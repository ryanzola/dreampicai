package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ryanzola/dreampicai/types"
	"github.com/ryanzola/dreampicai/view/generate"
)

func HandleGenerateIndex(w http.ResponseWriter, r *http.Request) error {

	data := generate.ViewData{
		Images: []types.Image{},
	}

	return render(r, w, generate.Index(data))
}

func HandleGenerateCreate(w http.ResponseWriter, r *http.Request) error {
	return render(r, w, generate.GalleryImage(types.Image{Status: types.ImageStatusPending}))
}

func HandleGenerateImageStatus(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	// fetch from db
	image := types.Image{
		Status: types.ImageStatusPending,
	}
	slog.Info("Checking image status", "id", id)

	return render(r, w, generate.GalleryImage(image))
}
