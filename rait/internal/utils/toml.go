package utils

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"net/http"
	"net/url"
	"os"
)

func ReaderFromPath(path string) (io.ReadCloser, error) {
	var u *url.URL
	var err error
	u, err = url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %w", err)
	}
	switch u.Scheme {
	case "http", "https":
		var resp *http.Response
		resp, err = http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get file from http: %w", err)
		}
		return resp.Body, nil
	case "":
		return os.Open(path)
	case "stdin":
		return os.Stdin, nil
	default:
		return nil, fmt.Errorf("unsupported url scheme: %w", err)
	}
}

func DecodeTOMLFromPath(path string, v interface{}) error {
	var source io.ReadCloser
	var err error
	source, err = ReaderFromPath(path)
	if err != nil {
		return err
	}
	defer source.Close()
	_, err = toml.DecodeReader(source, v)
	if err != nil {
		return err
	}
	return nil
}
