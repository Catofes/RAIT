package misc

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// WriteCloserFromPath returns a WriteCloser from the given path
// path can be a file system path, or "-" for stdout
func WriteCloserFromPath(path string) (io.WriteCloser, error) {
	parsed, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %s: %w", path, err)
	}
	switch parsed.Scheme {
	case "":
		if path == "-" {
			return os.Stdout, nil
		}
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem path: %s: %w", path, err)
		}
		return file, nil
	default:
		return nil, fmt.Errorf("unsupported url scheme to write: %s", parsed.Scheme)
	}
}

// ReadCloserFromPath returns a ReadCloser from the given path
// path can be a file system path, a http url, or "-" for stdin
func ReadCloserFromPath(path string) (io.ReadCloser, error) {
	parsed, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %s: %w", path, err)
	}
	switch parsed.Scheme {
	case "http", "https":
		resp, err := http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to make http request: %s: %w", path, err)
		}
		return resp.Body, nil
	case "":
		if path == "-" {
			return os.Stdin, nil
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open filesystem path: %s: %w", path, err)
		}
		return file, nil
	default:
		return nil, fmt.Errorf("unsupported url scheme to read: %s", parsed.Scheme)
	}
}
