package misc

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"io/ioutil"
)

// UnmarshalHCL decodes the hcl file read from path
// then unmarshal it into the given interface
func UnmarshalHCL(path string, v interface{}) error {
	source, err := NewReadCloser(path)
	if err != nil {
		return err
	}
	defer source.Close()

	data, err := ioutil.ReadAll(source)
	if err != nil {
		return err
	}

	err = hclsimple.Decode("source.hcl", data, nil, v)
	if err != nil {
		return fmt.Errorf("failed to decode hcl: %s: %w", path, err)
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
		return fmt.Errorf("failed to decode toml: %s: %w", path, err)
	}
	return nil
}
