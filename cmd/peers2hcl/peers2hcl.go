package main

import (
	"github.com/BurntSushi/toml"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"gitlab.com/NickCao/RAIT/v2/pkg/rait"
	"log"
	"os"
)

func main() {
	var values rait.Peers
	_, err := toml.DecodeReader(os.Stdin, &values)
	if err != nil {
		log.Fatal(err)
	}
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&values, f.Body())
	_, err = f.WriteTo(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
