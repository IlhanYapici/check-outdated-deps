package cache

import (
	"check-outdated-deps/internal/npm"
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

// GetCachedVersion attempts to retrieve a fresh version from the local store.
func GetCachedVersion(db *bolt.DB, pkgName string) (string, bool) {
	var version string

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Versions"))
		if b == nil {
			return nil
		}

		v := b.Get([]byte(pkgName))
		if v == nil {
			return nil
		}

		var entry CacheEntry
		if err := json.Unmarshal(v, &entry); err != nil {
			return err
		}

		// Only return if it's less than 4 hours old
		if time.Since(entry.FetchedAt) < 4*time.Hour {
			version = entry.Version
		}
		return nil
	})

	if err != nil {
		return "", false
	}

	return version, version != ""
}

// SaveToCache remains for single-off updates if needed.
func SaveToCache(db *bolt.DB, pkgName string, version string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Versions"))
		if err != nil {
			return err
		}

		entry := CacheEntry{Version: version, FetchedAt: time.Now()}
		encoded, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		return b.Put([]byte(pkgName), encoded)
	})
}

// SaveFromChannel drains the results channel and saves everything in one transaction.
// This should be called after you are sure the workers are done and the channel is closed.
func SaveFromChannel(db *bolt.DB, results <-chan npm.Package) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Versions"))
		if err != nil {
			return err
		}

		now := time.Now()

		// This loop will run until the channel is closed
		for pkg := range results {
			// Skip empty versions
			if pkg.Version == "" {
				continue
			}

			entry := CacheEntry{
				Version:   pkg.Version,
				FetchedAt: now,
			}

			encoded, err := json.Marshal(entry)
			if err != nil {
				continue
			}

			if err := b.Put([]byte(pkg.Name), encoded); err != nil {
				return err
			}
		}
		return nil
	})
}
