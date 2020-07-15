package misc

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

// DecodeTOMLFromPath decodes the toml file read from path
// then unmarshal it into the given interface
func DecodeTOMLFromPath(path string, v interface{}) error {
	source, err := ReadCloserFromPath(path)
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
