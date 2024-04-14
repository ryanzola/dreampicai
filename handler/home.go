package handler

import (
	"net/http"
	"time"

	"github.com/ryanzola/dreampicai/view/home"
)

func HandleLongProcess(w http.ResponseWriter, r *http.Request) error {
	// This is a long process
	time.Sleep(5 * time.Second)

	return home.UserLikes(1000).Render(r.Context(), w)
}

func HandleHomeIndex(w http.ResponseWriter, r *http.Request) error {
	return home.Index().Render(r.Context(), w)
}
