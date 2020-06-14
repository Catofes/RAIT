package misc

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"os"
)

// WriteCloserFromPath returns a WriteCloser from the given path
// path can be a file system path, or "-" for stdout
func WriteCloserFromPath(path string) (io.WriteCloser, error) {
	logger := zap.S().Named("misc.WriteCloserFromPath")
	parsed, err := url.Parse(path)
	if err != nil {
		logger.Errorf("failed to parse path: %s, error %s", path, err)
		return nil, err
	}
	switch parsed.Scheme {
	case "":
		if path == "-" {
			return os.Stdout, nil
		}
		file, err := os.Create(path)
		if err != nil {
			logger.Errorf("failed to create filesystem path: %s, error %s", path, err)
			return nil, err
		}
		return file, nil
	default:
		return nil, fmt.Errorf("unsupported url scheme: %s", parsed.Scheme)
	}
}

// ReadCloserFromPath returns a ReadCloser from the given path
// path can be a file system path, a http url, or "-" for stdin
func ReadCloserFromPath(path string) (io.ReadCloser, error) {
	logger := zap.S().Named("misc.ReadCloserFromPath")
	parsed, err := url.Parse(path)
	if err != nil {
		logger.Errorf("failed to parse path: %s, error %s", path, err)
		return nil, err
	}
	switch parsed.Scheme {
	case "http", "https":
		resp, err := http.Get(path)
		if err != nil {
			logger.Errorf("failed to make http request: %s, error %s", path, err)
			return nil, err
		}
		return resp.Body, nil
	case "":
		if path == "-" {
			return os.Stdin, nil
		}
		file, err := os.Open(path)
		if err != nil {
			logger.Errorf("failed to open filesystem path: %s, error %s", path, err)
			return nil, err
		}
		return file, nil
	default:
		return nil, fmt.Errorf("unsupported url scheme: %s", parsed.Scheme)
	}
}
