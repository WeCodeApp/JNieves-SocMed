package posts

import (
	"encoding/json"
	"go-rest-api/database"
	"log"
	"net/http"
)

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("DELETE request received: %s", r.URL.String())

	if IsCollectionRequest(r) {
		http.Error(w, "Cannot delete entire collection", http.StatusMethodNotAllowed)
		return
	}

	id, ok := ExtractPostID(w, r)
	if !ok {
		return
	}

	// Use the existing database connection
	db := database.DB // Changed to use global DB

	_, err := db.Exec("DELETE FROM likes WHERE post_id = ?", id)
	if err != nil {
		log.Println("Error deleting related likes:", err)
	}

	result, err := db.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		log.Println("Database delete error:", err)
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting affected rows:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "success",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}
