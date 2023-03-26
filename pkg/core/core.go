package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bschaatsbergen/tftag/pkg/helpers"
	"github.com/bschaatsbergen/tftag/pkg/mapper"
	"github.com/bschaatsbergen/tftag/pkg/model"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/sirupsen/logrus"
)

const filterCommentPrefix = "#tftag:"

func Main(dir string) {
	config, err := mapper.ParseTfTagFile(dir)
	if err != nil {
		panic(err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".tf" {
			if err := processFile(dir, config, file); err != nil {
				panic(err)
			}
			logrus.Debugf("Tags added to %s\n", file.Name())
		}
	}
}

// getResourceInfo returns the resource type and identifier for the given block.
func getResourceInfo(b *hclwrite.Block) (string, string) {
	return b.Labels()[0], b.Labels()[1]
}

// getFilterComment retrieves the filter string from the tftag comment in the given HCL block.
// The tftag comment must be in the form "// tftag:<filter>".
func getFilterComment(b *hclwrite.Block) string {
	var filter string

	// Loop through every line in the Terraform resource
	for _, attr := range b.Body().BuildTokens(nil) {
		// If a line contains the tftag comment, extract anything after `:`
		if bytes.HasPrefix(attr.Bytes, []byte(filterCommentPrefix)) {
			filterComment := string(attr.Bytes)
			value := filterComment[len(filterCommentPrefix):]
			filter = strings.TrimSpace(value)
			break
		}
	}
	return filter
}

// processFile reads a Terraform file from the given directory and processes its resource blocks,
// setting tags according to the provided configuration. The updated file is then written back to
// disk. Returns an error if there was an issue reading or writing the file, or if there was an
// issue parsing the file or its blocks.
func processFile(dir string, config model.Config, file os.DirEntry) error {
	// Read the contents of the file
	tfBytes, err := os.ReadFile(filepath.Join(dir, file.Name()))
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", file.Name(), err)
	}

	// Parse the file using HCL2
	f, diags := hclwrite.ParseConfig(tfBytes, "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("failed to parse file %s: %w", file.Name(), diags)
	}

	// Process each block in the file
	for _, b := range f.Body().Blocks() {
		// Get the resource type and identifier for this block
		resourceType, resourceIdentifier := getResourceInfo(b)

		// Get either `tags` or `labels` depending on the provider
		resourceTagAttribute, err := helpers.GetResourceTagType(resourceType)
		if err != nil {
			panic(err)
		}

		// Get the "tags" or "labels" attribute, if it exists
		tagsAttr := b.Body().GetAttribute(resourceTagAttribute)

		// Determine the tokens to write for the "tags" or "labels" attribute (or lack thereof)
		var writeTokens []*hclwrite.Token
		if tagsAttr != nil {
			// If the "tags" or "labels" attribute exists, extract the tokens between the braces
			buildTokens := tagsAttr.BuildTokens(nil)
			buildTokens = buildTokens[2:]
			depth := 0
			for _, t := range buildTokens {
				if t.Type == hclsyntax.TokenOBrace {
					depth++
				} else if t.Type == hclsyntax.TokenCBrace {
					depth--
					if depth == 0 {
						break
					}
				}
				writeTokens = append(writeTokens, t)
			}
		} else {
			// If the "tags" or "labels" attribute doesn't exist, create an empty block
			writeTokens = []*hclwrite.Token{
				{Type: hclsyntax.TokenOBrace, Bytes: []byte{'{'}, SpacesBefore: 1},
				{Type: hclsyntax.TokenNewline, Bytes: []byte{'\n'}, SpacesBefore: 1},
			}
		}

		// Get the filter comment for this block (if it exists)
		filter := getFilterComment(b)

		// Set tags on this block if it's taggable
		if helpers.IsTaggableResource(resourceType) {
			setTags(resourceTagAttribute, config, b, filter, writeTokens)
			logrus.Infof("Tagged `%s.%s` in %s\n", resourceType, resourceIdentifier, file.Name())
		} else {
			logrus.Warnf("Resource `%s.%s` in %s isn't taggable\n", resourceType, resourceIdentifier, file.Name())
		}
	}

	// Write the updated file back to disk
	if err := os.WriteFile(filepath.Join(dir, file.Name()), []byte(f.Bytes()), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", file.Name(), err)
	}

	return nil
}

// isTokenInArray checks if a given token type is present in an array of token types.
func isTokenInArray(tokenType hclsyntax.TokenType, tokenArray []hclsyntax.TokenType) bool {
	for _, t := range tokenArray {
		if tokenType == t {
			return true
		}
	}
	return false
}

// removeTagsFromExistingTokens removes tags from a list of HCL tokens by identifying the positions of
// identifier tokens followed by equal tokens and checking if they match any of the provided tags. If a match
// is found, the tokens between the previous identifier position and the current identifier position are skipped.
// The resulting list of tokens is then returned.
func removeTagsFromExistingTokens(tags map[string]string, tokens []*hclwrite.Token) []*hclwrite.Token {
	// Identify token pairs that indicate the start and end of a nested expression
	openPairs := []hclsyntax.TokenType{
		hclsyntax.TokenOBrace,
		hclsyntax.TokenOQuote,
		hclsyntax.TokenOParen,
		hclsyntax.TokenOBrack,
		hclsyntax.TokenOHeredoc,
	}

	closingPairs := []hclsyntax.TokenType{
		hclsyntax.TokenCBrace,
		hclsyntax.TokenCQuote,
		hclsyntax.TokenCParen,
		hclsyntax.TokenCBrack,
		hclsyntax.TokenCHeredoc,
	}

	// Keep track of the depth in the token hierarchy and the positions of identifier tokens
	depth := 0
	identifierPositions := []int{}
	for i, token := range tokens {
		if isTokenInArray(token.Type, openPairs) {
			depth++
		}
		if isTokenInArray(token.Type, closingPairs) {
			depth--
		}
		if depth == 1 && token.Type == hclsyntax.TokenIdent && i+1 < len(tokens) && tokens[i+1].Type == hclsyntax.TokenEqual {
			identifierPositions = append(identifierPositions, i)
		}
	}

	// Remove tokens between identifier positions that match the provided tags
	previousPosition := 0
	result := []*hclwrite.Token{}
	for _, identifierPosition := range identifierPositions {
		name := string(tokens[identifierPosition].Bytes)
		if _, ok := tags[name]; ok {
			result = append(result, tokens[previousPosition:identifierPosition]...)
			if i := sort.SearchInts(identifierPositions, identifierPosition+1); i < len(identifierPositions) {
				previousPosition = identifierPositions[i]
			} else {
				previousPosition = len(tokens)
			}
		}
	}
	result = append(result, tokens[previousPosition:]...)
	return result
}

// appendTagsAsTokens appends HCL tokens representing the given tags to the provided slice of tokens.
// It removes any existing tokens with the same key before appending new tokens.
// The resulting tokens represent a series of HCL attribute assignments, one per tag,
// in the format "key = "value"\n".
// Returns the modified slice of tokens.
func appendTagsAsTokens(tags map[string]string, tokens []*hclwrite.Token) []*hclwrite.Token {
	// Remove existing tokens with the same key, if any
	tokens = removeTagsFromExistingTokens(tags, tokens)

	for key, val := range tags {
		// Append identifier token for the tag key
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(key),
		})

		// Append equal sign token with space on either side
		tokens = append(tokens, &hclwrite.Token{
			Type:         hclsyntax.TokenEqual,
			Bytes:        []byte(" = "),
			SpacesBefore: 1,
		})

		// Append quoted literal token for the tag value
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenQuotedLit,
			Bytes: []byte(fmt.Sprintf(`"%s"`, val)),
		})

		// Append newline token to separate tags
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte("\n"),
		})
	}

	return tokens
}

// setTags sets the "tags" or "labels" attribute in the specified HCL block using the tags from the given tftag configuration.
func setTags(resourceTagAttribute string, config model.Config, b *hclwrite.Block, filter string, tokens []*hclwrite.Token) {
	matched := false
	for _, tfTagConfig := range config.Config {
		// Check if the filter matches the tftag config item
		if strings.TrimSpace(tfTagConfig.Type) == strings.TrimSpace(filter) {
			matched = true
			tokens = appendTagsAsTokens(tfTagConfig.Tags, tokens)
			break // Exit loop after first matching config item
		}
	}
	if !matched {
		for _, tfTagConfig := range config.Config {
			tokens = appendTagsAsTokens(tfTagConfig.Tags, tokens)
		}
	}
	tokens = append(tokens, &hclwrite.Token{
		Type:         hclsyntax.TokenCBrace,
		Bytes:        []byte("}"),
		SpacesBefore: 1,
	})

	b.Body().SetAttributeRaw(resourceTagAttribute, tokens)
}
