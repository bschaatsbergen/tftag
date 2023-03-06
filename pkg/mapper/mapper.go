package mapper

import (
	"fmt"
	"io/ioutil"

	"github.com/bschaatsbergen/tftag/pkg/model"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

const (
	tfTagFileName = ".tftag.hcl"
)

// ParseTfTagFile reads and parses a Terraform tag configuration file in the specified directory, returning a model.Config struct.
func ParseTfTagFile(directory string) (model.Config, error) {
	// Read the contents of the .hcl file
	b, err := ioutil.ReadFile(tfTagFileName)
	if err != nil {
		return model.Config{}, fmt.Errorf("error reading file: %s", err)
	}

	// Parse the HCL configuration into a struct
	var config model.Config
	err = hclsimple.Decode(tfTagFileName, b, nil, &config)
	if err != nil {
		return model.Config{}, fmt.Errorf("error parsing HCL: %s", err)
	}

	return config, nil
}
