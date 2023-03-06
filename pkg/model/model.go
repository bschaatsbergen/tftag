package model

type Config struct {
	Config []TfTagConfig `hcl:"tftag,block"`
}

type TfTagConfig struct {
	Type string            `hcl:"type,label"`
	Tags map[string]string `hcl:"tags,attr"`
}
