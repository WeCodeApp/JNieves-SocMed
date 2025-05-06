package posts

import (
	"database/sql"
	"encoding/json"
	"go-rest-api/database"
	"go-rest-api/internal/models"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("PUT request received: %s", r.URL.String())

	if IsCollectionRequest(r) {
		http.Error(w, "Cannot update entire collection", http.StatusMethodNotAllowed)
		return
	}

	id, ok := ExtractPostID(w, r)
	if !ok {
		return
	}

	var updatedPost models.Post
	err := json.NewDecoder(r.Body).Decode(&updatedPost)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	db := database.DB

	existingPost, err := getExistingPost(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			log.Println("Database query error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if updatedPost.Content == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	updatedPost.ID = existingPost.ID
	updatedPost.UserID = existingPost.UserID
	updatedPost.CreatedAt = existingPost.CreatedAt
	updatedPost.Likes = existingPost.Likes
	updatedPost.UpdatedAt = time.Now()

	updateQuery := `
        UPDATE posts
        SET content = ?, image_url = ?, updated_at = ?
        WHERE id = ?
    `
	_, err = db.Exec(
		updateQuery,
		updatedPost.Content,
		updatedPost.ImageURL,
		updatedPost.UpdatedAt,
		id,
	)

	if err != nil {
		log.Println("Database update error:", err)
		http.Error(w, "Error updating post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPost)
}

func PatchPostHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("PATCH request received: %s", r.URL.String())

	if IsCollectionRequest(r) {
		http.Error(w, "Cannot patch entire collection", http.StatusMethodNotAllowed)
		return
	}

	id, ok := ExtractPostID(w, r)
	if !ok {
		return
	}

	var updates map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	db := database.DB

	existingPost, err := getExistingPost(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			log.Println("Database query error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	postVal := reflect.ValueOf(&existingPost).Elem()
	postType := postVal.Type()

	protectedFields := map[string]bool{
		"id":         true,
		"user_id":    true,
		"created_at": true,
	}

	for k, v := range updates {
		if protectedFields[k] {
			continue
		}

		for i := 0; i < postVal.NumField(); i++ {
			field := postType.Field(i)
			jsonTag := field.Tag.Get("json")
			jsonName := strings.Split(jsonTag, ",")[0]

			if jsonName == k && postVal.Field(i).CanSet() {
				fieldVal := postVal.Field(i)
				switch k {
				case "content":
					if strVal, ok := v.(string); ok {
						fieldVal.SetString(strVal)
					}
				case "image_url":
					if strVal, ok := v.(string); ok {
						fieldVal.SetString(strVal)
					}
				case "likes":
					if numVal, ok := v.(float64); ok {
						fieldVal.SetInt(int64(numVal))
					}
				}
			}
		}
	}

	existingPost.UpdatedAt = time.Now()

	query := `
        UPDATE posts
        SET content = ?, image_url = ?, updated_at = ?, likes = ?
        WHERE id = ?
    `

	_, err = db.Exec(
		query,
		existingPost.Content,
		existingPost.ImageURL,
		existingPost.UpdatedAt,
		existingPost.Likes,
		id,
	)

	if err != nil {
		log.Println("Database update error:", err)
		http.Error(w, "Error updating post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingPost)
}

func getExistingPost(db *sql.DB, id int) (models.Post, error) {
	var post models.Post
	var updatedAt sql.NullTime
	var imageURL sql.NullString

	query := `
        SELECT id, user_id, content, image_url, created_at, updated_at, likes 
        FROM posts 
        WHERE id = ?
    `

	err := db.QueryRow(query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&imageURL,
		&post.CreatedAt,
		&updatedAt,
		&post.Likes,
	)

	if err != nil {
		return post, err
	}

	if imageURL.Valid {
		post.ImageURL = imageURL.String
	}

	if updatedAt.Valid {
		post.UpdatedAt = updatedAt.Time
	}

	return post, nil
}
