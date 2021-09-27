package lib

import (
	_ "embed"
	"os"
)

//go:embed res/.mk.yaml
var sample []byte

//go:embed res/mk.json
var validator []byte

func InitFile() error {
	return os.WriteFile(CLI.File, sample, 0600)
}

func DumpValidator() error {
	return os.WriteFile(CLI.File, validator, 0600)
}
