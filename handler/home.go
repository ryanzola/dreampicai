package handler

import (
	"net/http"

	"github.com/ryanzola/dreampicai/view/home"
)

func HandleHomeIndex(w http.ResponseWriter, r *http.Request) error {
	// user := getAuthenticatedUser(r)
	return home.Index().Render(r.Context(), w)
}
