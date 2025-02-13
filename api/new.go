package api

import (
	"net/http"

	"github.com/google/uuid"
)

func New(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte(uuid.NewString()))
}
