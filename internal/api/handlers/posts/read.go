package posts

import (
	"database/sql"
	"encoding/json"
	"go-rest-api/database" // Changed import
	"go-rest-api/internal/models"
	"log"
	"net/http"
	"strconv"
)

// GetPostHandler handles GET requests for posts
func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	// Log request for debugging
	log.Printf("GET request received: %s", r.URL.String())

	// Use the existing database connection
	db := database.DB // Changed to use global DB

	// Check if it's a collection request
	if IsCollectionRequest(r) {
		getPostCollection(w, r, db)
		return
	}

	// Otherwise, get a specific post by ID
	id, ok := ExtractPostID(w, r)
	if !ok {
		// Error response already sent by ExtractPostID
		return
	}

	getSinglePost(w, r, db, id)
}

func getPostCollection(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	if pageStr == "" {
		pageStr = "1"
	}
	if limitStr == "" {
		limitStr = "10"
	}

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	maxLimit := 100
	if limit > maxLimit {
		http.Error(w, "Limit cannot be greater than 100", http.StatusBadRequest)
		return
	}

	query := "SELECT id, user_id, content, image_url, created_at, updated_at, likes FROM posts ORDER BY created_at DESC LIMIT ? OFFSET ?"
	queryCount := "SELECT COUNT(id) FROM posts"

	offset := (page - 1) * limit

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	postList := make([]models.Post, 0)
	for rows.Next() {
		var post models.Post
		var updatedAt sql.NullTime
		var imageURL sql.NullString

		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &imageURL, &post.CreatedAt, &updatedAt, &post.Likes)
		if err != nil {
			log.Println("Database scan error:", err)
			http.Error(w, "Database scan error", http.StatusInternalServerError)
			return
		}

		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}

		if updatedAt.Valid {
			post.UpdatedAt = updatedAt.Time
		}

		postList = append(postList, post)
	}

	var totalPosts int
	err = db.QueryRow(queryCount).Scan(&totalPosts)
	if err != nil {
		log.Println("Count query error:", err)
		totalPosts = 0
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Post `json:"data"`
	}{
		Status: "success",
		Count:  totalPosts,
		Data:   postList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getSinglePost(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) {
	var post models.Post
	var updatedAt sql.NullTime
	var imageURL sql.NullString

	query := "SELECT id, user_id, content, image_url, created_at, updated_at, likes FROM posts WHERE id = ?"
	err := db.QueryRow(query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&imageURL,
		&post.CreatedAt,
		&updatedAt,
		&post.Likes,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}

	if imageURL.Valid {
		post.ImageURL = imageURL.String
	}

	if updatedAt.Valid {
		post.UpdatedAt = updatedAt.Time
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}
