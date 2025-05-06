package handlers

import (
	"go-rest-api/internal/api/handlers/posts"
	"go-rest-api/internal/models"
	"net/http"
)

var (
	postsMap = make(map[int]models.Post)
	nextID   = 1
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		posts.GetPostHandler(w, r)
	case http.MethodPost:
		posts.AddPostHandler(w, r)
	case http.MethodPut:
		posts.UpdatePostHandler(w, r)
	case http.MethodPatch:
		posts.PatchPostHandler(w, r)
	case http.MethodDelete:
		posts.DeletePostHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
