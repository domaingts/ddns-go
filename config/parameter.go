package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

type Parameter struct {
	Listen string       `yaml:"listen"`
	TLS    *TLSParameter `yaml:"tls,omitempty"`
}

type TLSParameter struct {
	CertificatePath string `yaml:"certificate_path"`
	KeyPath         string `yaml:"key_path"`
}

func ReadParameter(in io.Reader) (*Parameter, error) {
	parameter := &Parameter{}
	err := yaml.NewDecoder(in).Decode(parameter)
	return parameter, err
}
