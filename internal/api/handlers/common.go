package handlers

import (
	"database/sql"
	"go-rest-api/database"
	"strconv"
	"strings"
)

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"id":         true,
		"user_id":    true,
		"content":    true,
		"created_at": true,
		"likes":      true,
	}

	return validFields[field]
}

func getDbConnection() (*sql.DB, error) {
	// Return the existing DB connection instead of creating a new one
	return database.DB, nil
}

func getIDFromPath(path, prefix string) (int, error) {
	path = strings.TrimPrefix(path, prefix)
	idStr := strings.TrimSuffix(path, "/")

	if idStr == "" {
		return 0, nil
	}

	return strconv.Atoi(idStr)
}
