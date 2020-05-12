package utils

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"net/http"
	"net/url"
	"os"
)

// WriteCloserFromPath returns a WriteCloser from the given path
// Path can be a file system path, or "-" for stdout
func WriteCloserFromPath(path string) (io.WriteCloser, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("WriteCloserFromPath: failed to parse path: %s", path)
	}
	switch u.Scheme {
	case "":
		if path == "-" {
			return os.Stdout, nil
		}
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("WriteCloserFromPath: failed to open file %s: %w", path, err)
		}
		return file, nil
	default:
		return nil, fmt.Errorf("WriteCloserFromPath: unsupported url scheme: %s", u.Scheme)
	}
}

// ReadCloserFromPath returns a ReadCloser from the given path
// Path can be a file system path, a http url, or "-" for stdin
func ReadCloserFromPath(path string) (io.ReadCloser, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("ReadCloserFromPath: failed to parse path: %s", path)
	}
	switch u.Scheme {
	case "http", "https":
		resp, err := http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("ReadCloserFromPath: failed to make http request to %s: %w", path, err)
		}
		return resp.Body, nil
	case "":
		if path == "-" {
			return os.Stdin, nil
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("ReadCloserFromPath: failed to open file %s: %w", path, err)
		}
		return file, nil
	default:
		return nil, fmt.Errorf("ReadCloserFromPath: unsupported url scheme: %s", u.Scheme)
	}
}

// DecodeTOMLFromPath decodes the toml file loaded from path
// Then unmarshal it into the given interface
func DecodeTOMLFromPath(path string, v interface{}) error {
	source, err := ReadCloserFromPath(path)
	if err != nil {
		return err
	}
	defer source.Close()
	_, err = toml.DecodeReader(source, v)
	if err != nil {
		return fmt.Errorf("DecodeTOMLFromPath: failed to decode toml: %w", err)
	}
	return nil
}
