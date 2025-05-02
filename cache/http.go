package cache

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

func CreateHandler(cache OrasCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			if r.Method == "GET" || r.Method == "HEAD" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			return
		} else if len(r.URL.Path) > 128 {
			if r.Method == "GET" || r.Method == "HEAD" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			return
		}

		key := strings.ReplaceAll(r.URL.Path[1:], "/", "_")
		if r.Method == "PUT" {
			// We intentionally ignore any error returned, as the
			// cache response should not return an error and just
			// silently fail.
			cache.Store(key, r.Body)
			defer r.Body.Close()
			w.WriteHeader(http.StatusNoContent)
		} else if r.Method == "GET" || r.Method == "HEAD" {
			var buffer bytes.Buffer
			err := cache.Restore(key, &buffer)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusOK)
				if r.Method == "GET" {
					io.Copy(w, &buffer)
				}
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
