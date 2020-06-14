package misc

import (
	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

// DecodeTOMLFromPath decodes the toml file read from path
// then unmarshal it into the given interface
func DecodeTOMLFromPath(path string, v interface{}) error {
	logger := zap.S().Named("misc.DecodeTOMLFromPath")

	source, err := ReadCloserFromPath(path)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = toml.DecodeReader(source, v)
	if err != nil {
		logger.Errorf("failed to decode toml: %s, error %s", path, err)
		return err
	}
	return nil
}
