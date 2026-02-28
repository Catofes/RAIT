package misc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

type peerCache struct {
	URL  string
	Etag string
	Data []byte
}

func loadPeerCache(path string) *peerCache {
	p := &peerCache{}
	p.Data = make([]byte, 0)
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, p); err == nil {
			return p
		}
	}
	p.Etag = ""
	return p
}

func (s *peerCache) save(path string) {
	data, _ := json.Marshal(s)
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("failed to save peer cache: %s\n", err)
	}
}

func loadURLwithCache(url, cachePath string) []byte {
	if c := loadPeerCache(cachePath); c.Etag != "" {
		fmt.Printf("load cache from %s\n", cachePath)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("If-None-Match", c.Etag)
		if resp, err := http.DefaultClient.Do(req); err == nil {
			defer resp.Body.Close()
			switch resp.StatusCode {
			case http.StatusNotModified:
				fmt.Printf("load url from %s success: %s\n", url, resp.Status)
				return c.Data
			case http.StatusOK:
				c.Etag = resp.Header.Get("ETag")
				c.URL = url
				c.Data, _ = io.ReadAll(resp.Body)
				if c.Etag != "" {
					c.save(cachePath)
				}
				fmt.Printf("load url from %s success: %s\n", url, resp.Status)
				return c.Data
			default:
				fmt.Printf("failed to load url from %s: %s, load from cache\n", url, resp.Status)
				return c.Data
			}
		} else {
			fmt.Printf("failed to load url from %s: %s, load from cache\n", url, err)
			return c.Data
		}
	} else {
		fmt.Printf("cache miss, load from %s\n", url)
		req, _ := http.NewRequest("GET", url, nil)
		if resp, err := http.DefaultClient.Do(req); err == nil {
			defer resp.Body.Close()
			switch resp.StatusCode {
			case http.StatusOK:
				c.Etag = resp.Header.Get("ETag")
				c.URL = url
				c.Data, _ = io.ReadAll(resp.Body)
				if c.Etag != "" {
					c.save(cachePath)
				}
				fmt.Errorf("load url from %s success: %s", url, resp.Status)
				return c.Data
			default:
				fmt.Errorf("failed to load url from %s: %s, cache empty", url, resp.Status)
				return nil
			}
		} else {
			fmt.Errorf("failed to load url from %s: %s, cache empty", url, err)
			return nil
		}
	}
}

func LoadPeers(path, cachePath string, v interface{}) error {
	data := loadURLwithCache(path, cachePath)
	if data == nil {
		return fmt.Errorf("failed to load hcl from %s and cache", path)
	}
	err := hclsimple.Decode("source.hcl", data, nil, v)
	if err != nil {
		return fmt.Errorf("failed to decode hcl: %s: %s", path, err)
	}
	return nil
}

// UnmarshalHCL decodes the hcl file read from path
// then unmarshal it into the given interface
func UnmarshalHCL(path string, v interface{}) error {
	var err error
	var data []byte
	if _, err := url.Parse(path); err != nil {
		resp, err := http.Get(path)
		if err != nil {
			return fmt.Errorf("failed to load hcl from %s: %s", path, err)
		}
		defer resp.Body.Close()
		if data, err = ioutil.ReadAll(resp.Body); err != nil {
			return fmt.Errorf("failed to load hcl from %s: %s", path, err)
		}
	} else {
		source, err := NewReadCloser(path)
		if err != nil {
			return err
		}
		defer source.Close()

		data, err = ioutil.ReadAll(source)
		if err != nil {
			return err
		}
	}
	err = hclsimple.Decode("source.hcl", data, nil, v)
	if err != nil {
		return fmt.Errorf("failed to decode hcl: %s: %s", path, err)
	}
	return nil
}

// UnmarshalTOML decodes the toml file read from path
// then unmarshal it into the given interface
func UnmarshalTOML(path string, v interface{}) error {
	source, err := NewReadCloser(path)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = toml.DecodeReader(source, v)
	if err != nil {
		return fmt.Errorf("failed to decode toml: %s: %s", path, err)
	}
	return nil
}
