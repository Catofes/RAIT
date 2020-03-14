package rait

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"text/template"
)

const BabeldConfigTemplate = `{{$prefix := .IFPrefix}}
random-id true
export-table 254

default type tunnel link-quality true split-horizon false
default rxcost 32 hello-interval 20 max-rtt-penalty 1024 rtt-max 1024

interface {{$prefix}}local rxcost 1 hello-interval 4

{{range .IFSuffix}}interface {{$prefix}}{{.}}
{{end}}
redistribute local deny
`

func GenerateBabeldConfig(IFPrefix string, n int, filepath string) error {
	var IFSuffix []string
	for i := 0; i < n; i++ {
		IFSuffix = append(IFSuffix, strconv.Itoa(i))
	}
	tmpl, err := template.New("babeld.conf").Parse(BabeldConfigTemplate)
	if err != nil {
		return err
	}
	var conf bytes.Buffer
	err = tmpl.Execute(&conf, struct {
		IFPrefix string
		IFSuffix []string
	}{
		IFPrefix: IFPrefix,
		IFSuffix: IFSuffix,
	})
	if err != nil {
		return err
	}
	err = CreateParentDirIfNotExist(filepath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, conf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}
