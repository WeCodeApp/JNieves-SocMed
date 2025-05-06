package posts

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func ExtractPostID(w http.ResponseWriter, r *http.Request) (int, bool) {
	log.Printf("Processing URL: %s", r.URL.String())

	urlPath := r.URL.Path
	log.Printf("URL path: %s", urlPath)

	if urlPath == "/posts" || urlPath == "/posts/" {
		return 0, true
	}

	parts := strings.Split(urlPath, "/")
	var idPart string

	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "posts" && i+1 < len(parts) {
			idPart = parts[i+1]
			break
		}
	}

	if idPart == "" {
		return 0, true
	}

	id, err := strconv.Atoi(idPart)
	if err != nil {
		errMsg := fmt.Sprintf("Invalid post ID: %s", idPart)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return 0, false
	}

	return id, true
}

func IsCollectionRequest(r *http.Request) bool {
	path := r.URL.Path
	return path == "/posts" || path == "/posts/"
}
