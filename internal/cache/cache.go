package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const TTL = 2 * time.Minute

type CacheEntry[T any] struct {
	StoredAt int64 `json:"stored_at"` // Unix timestamp (seconds)
	Data     T     `json:"data"`
}

func cachePath(name string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ttg", name+".cache.json")
}

// Get loads cached data. Returns (data, true) if cache is valid, or (zero, false) if missing/expired.
func Get[T any](name string) (T, bool) {
	var zero T
	data, err := os.ReadFile(cachePath(name))
	if err != nil {
		return zero, false
	}
	var entry CacheEntry[T]
	if err := json.Unmarshal(data, &entry); err != nil {
		return zero, false
	}
	age := time.Since(time.Unix(entry.StoredAt, 0))
	if age > TTL {
		return zero, false
	}
	return entry.Data, true
}

// Set saves data to the cache.
func Set[T any](name string, data T) error {
	entry := CacheEntry[T]{
		StoredAt: time.Now().Unix(),
		Data:     data,
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(name), raw, 0600)
}

// Invalidate removes the cache file for the given name.
func Invalidate(name string) {
	os.Remove(cachePath(name))
}
