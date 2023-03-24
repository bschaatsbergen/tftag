package helpers

import (
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
