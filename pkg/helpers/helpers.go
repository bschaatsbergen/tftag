package helpers

import (
	"strings"

	aws_resources "github.com/bschaatsbergen/tftag/pkg/resources/aws"
	azure_resources "github.com/bschaatsbergen/tftag/pkg/resources/azure"
	google_resources "github.com/bschaatsbergen/tftag/pkg/resources/google"
	"golang.org/x/exp/slices"
)

const (
	AWSProviderResourcePrefix    = "aws_"
	GoogleProviderResourcePrefix = "google_"
	AzureProviderResourcePrefix  = "azurerm_"
)

func isAWSProviderResource(resource string) bool {
	return strings.Contains(resource, AWSProviderResourcePrefix)
}

func isGoogleProviderResource(resource string) bool {
	return strings.Contains(resource, GoogleProviderResourcePrefix)
}

func isAzureProviderResource(resource string) bool {
	return strings.Contains(resource, AzureProviderResourcePrefix)
}

// IsTaggableResource returns true if the specified resource is taggable by the tftag supported providers.
func IsTaggableResource(resource string) bool {
	switch {
	case isAWSProviderResource(resource):
		return slices.Contains(aws_resources.AWS, resource)
	case isGoogleProviderResource(resource):
		return slices.Contains(google_resources.Google, resource)
	case isAzureProviderResource(resource):
		return slices.Contains(azure_resources.Azure, resource)
	default:
		return false
	}
}
