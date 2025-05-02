package cache

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadPut(t *testing.T) {
	server, _ := createCacheServer(t)
	url := server.URL + "/cas/15e2b0d3c33891ebb0f1ef609ec419420c20e320ce94c65fbc8c3312448eb225"
	const expected = "my test"
	request, err := http.NewRequest("PUT", url, strings.NewReader(expected))
	if err != nil {
		t.Fatalf("failed to create PUT request: %s", err.Error())
	}
	request.Header.Set("Content-Type", "text/plain")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("failed to execute PUT request: %s", err.Error())
	} else if response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected %d, but received %d", http.StatusNoContent, response.StatusCode)
	}

	assertCacheEntry(t, url, expected)
}

func TestNonExistentHead(t *testing.T) {
	server, _ := createCacheServer(t)
	response, err := http.Head(server.URL + "/cas/15e2b0d3c33891ebb0f1ef609ec419420c20e320ce94c65fbc8c3312448eb225")
	if err != nil {
		t.Fatalf("failed to execute head request: %s", err.Error())
	} else if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %d, but received %d", http.StatusNotFound, response.StatusCode)
	}
}

func TestNonExistentGet(t *testing.T) {
	server, _ := createCacheServer(t)
	response, err := http.Get(server.URL + "/cas/15e2b0d3c33891ebb0f1ef609ec419420c20e320ce94c65fbc8c3312448eb225")
	if err != nil {
		t.Fatalf("failed to execute head request: %s", err.Error())
	} else if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected %d, but received %d", http.StatusNotFound, response.StatusCode)
	}
}

func TestRootGet(t *testing.T) {
	server, _ := createCacheServer(t)
	for _, path := range []string{"/", ""} {
		response, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatalf("failed to execute head request: %s", err.Error())
		} else if response.StatusCode != http.StatusOK {
			t.Fatalf("expected %d, but received %d", http.StatusOK, response.StatusCode)
		}
	}
}

func TestRootHead(t *testing.T) {
	server, _ := createCacheServer(t)
	for _, path := range []string{"/", ""} {
		response, err := http.Head(server.URL + path)
		if err != nil {
			t.Fatalf("failed to execute head request: %s", err.Error())
		} else if response.StatusCode != http.StatusOK {
			t.Fatalf("expected %d, but received %d", http.StatusOK, response.StatusCode)
		}
	}
}

func TestRootPost(t *testing.T) {
	server, _ := createCacheServer(t)
	for _, path := range []string{"/", ""} {
		response, err := http.Post(server.URL+path, "text/plain", strings.NewReader(""))
		if err != nil {
			t.Fatalf("failed to execute head request: %s", err.Error())
		} else if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected %d, but received %d", http.StatusBadRequest, response.StatusCode)
		}
	}
}

func createCacheServer(t *testing.T) (*httptest.Server, Cache) {
	cacheInstance := NewOras(t.Context(), setupInMemoryRepository(t))
	server := httptest.NewServer(CreateHandler(cacheInstance))
	t.Cleanup(server.Close)
	return server, cacheInstance
}

func assertCacheEntry(t *testing.T, url string, expected string) {
	response, err := http.Head(url)
	if err != nil {
		t.Fatalf("failed to execute HEAD request: %s", err.Error())
	} else if response.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, but received %d", http.StatusOK, response.StatusCode)
	}

	response, err = http.Get(url)
	if err != nil {
		t.Fatalf("failed to execute GET request: %s", err.Error())
	} else if response.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, but received %d", http.StatusOK, response.StatusCode)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("failed to read GET response body: %s", err.Error())
	} else if string(responseBytes) != expected {
		t.Fatalf("expected %s, but received %s", expected, string(responseBytes))
	}
}
