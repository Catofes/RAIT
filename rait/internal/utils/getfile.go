package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func FileFromURL(fileUrl string) ([]byte, error) {
	var u *url.URL
	var err error
	u, err = url.Parse(fileUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	var data []byte
	switch u.Scheme {
	case "http", "https":
		var resp *http.Response
		resp, err = http.Get(fileUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to get file from http: %w", err)
		}
		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body from http: %w", err)
		}
	case "":
		data, err = ioutil.ReadFile(u.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to read local file: %w", err)
		}
	case "stdin":
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported url scheme: %w", err)
	}
	return data, nil
}
