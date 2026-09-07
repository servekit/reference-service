// fetch.go — upstream fetcher with an on-disk cache so regeneration is
// offline-friendly and reproducible within a data release. Cache lives in
// tools/gen/.cache (gitignored), keyed by URL hash.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const httpTimeout = 90 * time.Second

// statusError marks a deterministic upstream answer (e.g. the 404s the
// languages endonym probe hits for ~half the tag set — CLDR ships no
// locale for them). Retrying cannot change it, so fetch() fails fast
// instead of burning the 3-attempt backoff on every regeneration.
type statusError struct{ url, status string }

func (e *statusError) Error() string { return "GET " + e.url + ": " + e.status }

// fetch returns the URL body, hitting the network with retries (upstream
// hosts are intermittently slow) and caching every success on disk.
// Non-200 answers are terminal — only transport failures are retried.
func fetch(url, cacheDir string) ([]byte, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(url))
	p := filepath.Join(cacheDir, hex.EncodeToString(sum[:8])+".json")
	if b, err := os.ReadFile(p); err == nil {
		return b, nil
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 5 * time.Second)
		}
		b, err := fetchOnce(url, p)
		if err == nil {
			return b, nil
		}
		if _, ok := err.(*statusError); ok {
			return nil, err // deterministic — retrying is wasted time
		}
		lastErr = err
	}
	return nil, lastErr
}

func fetchOnce(url, cachePath string) ([]byte, error) {
	cl := &http.Client{Timeout: httpTimeout}
	resp, err := cl.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, &statusError{url: url, status: resp.Status}
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(cachePath, b, 0o644); err != nil {
		return nil, err
	}
	return b, nil
}
