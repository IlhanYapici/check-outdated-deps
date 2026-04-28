package cache

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

const DB_FILENAME = ".check-outdated-deps.db"

type CacheEntry struct {
	Version   string    `json:"version"`
	FetchedAt time.Time `json:"fetched_at"`
}

func GetDatabasePath() string {
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}

	return filepath.Join(home, DB_FILENAME)
}
