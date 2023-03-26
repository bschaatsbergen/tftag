package helpers

import (
	"fmt"
	"strings"

	resources "github.com/bschaatsbergen/tftag/pkg/resources"
	"golang.org/x/exp/slices"
)

const (
	AWSResourcePrefix    = "aws_"
	GoogleResourcePrefix = "google_"
	AzureResourcePrefix  = "azurerm_"
)

func isAWSResource(resource string) bool {
	return strings.Contains(resource, AWSResourcePrefix)
}

func isGoogleResource(resource string) bool {
	return strings.Contains(resource, GoogleResourcePrefix)
}

func isAzureResource(resource string) bool {
	return strings.Contains(resource, AzureResourcePrefix)
}

// IsTaggableResource returns true if the specified resource is taggable by the tftag supported providers.
func IsTaggableResource(resource string) bool {
	switch {
	case isAWSResource(resource):
		return slices.Contains(resources.AWS, resource)
	case isGoogleResource(resource):
		return slices.Contains(resources.Google, resource)
	case isAzureResource(resource):
		return slices.Contains(resources.Azure, resource)
	default:
		return false
	}
}

// GetResourceTagType returns the respective 'tags' attribute associated with a given Terraform provider resource, or an error if the resource is not recognized.
func GetResourceTagType(resource string) (string, error) {
	switch {
	case isAWSResource(resource):
		return "tags", nil
	case isGoogleResource(resource):
		return "labels", nil
	case isAzureResource(resource):
		return "tags", nil
	default:
		return "", fmt.Errorf("unkown Terraform provider resource")
	}
}
