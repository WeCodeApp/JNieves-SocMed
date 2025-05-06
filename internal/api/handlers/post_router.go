package handlers

import (
	"net/http"
)

func PostRouter(w http.ResponseWriter, r *http.Request) {
	PostHandler(w, r)
}
