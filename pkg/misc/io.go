package misc

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// NewWriteCloser returns a WriteCloser from the given path
// path can be a file system path, or "-" for stdout
func NewWriteCloser(path string) (io.WriteCloser, error) {
	parsed, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %s: %s", path, err)
	}
	switch parsed.Scheme {
	case "":
		if path == "-" {
			return os.Stdout, nil
		}
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem path: %s: %s", path, err)
		}
		return file, nil
	default:
		return nil, fmt.Errorf("unsupported url scheme to write: %s", parsed.Scheme)
	}
}

// NewReadCloser returns a ReadCloser from the given path
// path can be a file system path, a http url, or "-" for stdin
func NewReadCloser(path string) (io.ReadCloser, error) {
	parsed, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %s: %s", path, err)
	}
	switch parsed.Scheme {
	case "http", "https":
		resp, err := http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to make http request: %s: %s", path, err)
		}
		return resp.Body, nil
	case "":
		if path == "-" {
			return os.Stdin, nil
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open filesystem path: %s: %s", path, err)
		}
		return file, nil
	default:
		return nil, fmt.Errorf("unsupported url scheme to read: %s", parsed.Scheme)
	}
}
