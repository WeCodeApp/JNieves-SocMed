package posts

import (
	"encoding/json"
	"go-rest-api/database"
	"go-rest-api/internal/models"
	"log"
	"net/http"
	"time"
)

func AddPostHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	var newPost models.Post
	err := json.NewDecoder(r.Body).Decode(&newPost)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	if newPost.UserID <= 0 {
		http.Error(w, "Valid user_id is required", http.StatusBadRequest)
		return
	}

	if newPost.Content == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	var userExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", newPost.UserID).Scan(&userExists)
	if err != nil {
		log.Println("Error checking if user exists:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !userExists {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare(`
        INSERT INTO posts (user_id, content, image_url, created_at, updated_at, likes)
        VALUES (?, ?, ?, ?, NULL, 0)
    `)
	if err != nil {
		log.Println("Error preparing statement:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	now := time.Now()
	result, err := stmt.Exec(
		newPost.UserID,
		newPost.Content,
		newPost.ImageURL,
		now,
	)
	if err != nil {
		log.Printf("Error inserting post: %v (UserID=%d, Content=%s)",
			err, newPost.UserID, newPost.Content)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting last insert ID:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	newPost.ID = int(lastInsertID)
	newPost.CreatedAt = now
	newPost.Likes = 0

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		Status string      `json:"status"`
		Data   models.Post `json:"data"`
	}{
		Status: "success",
		Data:   newPost,
	}

	json.NewEncoder(w).Encode(response)
}
