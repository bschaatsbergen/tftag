package core

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bschaatsbergen/tftag/pkg/helpers"
	"github.com/bschaatsbergen/tftag/pkg/mapper"
	"github.com/bschaatsbergen/tftag/pkg/model"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const filterCommentPrefix = "#tftag:"

func Main(dir string) {
	config, err := mapper.ParseTfTagFile(dir)
	if err != nil {
		panic(err)
	}

	// Loop through all .tf files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".tf" {
			// Read the contents of the .tf file
			tfBytes, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				panic(err)
			}

			// Parse the .tf file into an HCL AST
			f, diags := hclwrite.ParseConfig(tfBytes, "", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				panic(diags.Error())
			}

			// Loop through all blocks in the AST
			for _, b := range f.Body().Blocks() {
				// Extract the resource name, e.g. `aws_s3_bucket`
				resource := b.Labels()[0]
				// Extract the resource identifier, e.g. `default`
				resourceIdentifier := b.Labels()[1]

				tagsAttr := b.Body().GetAttribute("tags")
				toks := tagsAttr.BuildTokens(nil)
				depth := 0
				tokens := make([]*hclwrite.Token, 0)
				var buffer bytes.Buffer
				bufferWriter := io.Writer(&buffer)
				for _, t := range toks {
					logrus.Info(t)

					if t.Type == hclsyntax.TokenOBrace {
						depth = depth + 1
					}
					if t.Type == hclsyntax.TokenCBrace {
						depth = depth - 1
						if depth == 0 {
							break
						}
					}
					tokens = append(tokens, t)
					bufferWriter.Write(t.Bytes)
				}

				logrus.Info(string(buffer.String()))

				// Check if the Terraform resource contains a `#tftag:` filter comment
				filter := getFilterComment(b)

				// Determine whether the resource is supported by tftag
				if helpers.IsTaggableResource(resource) {
					setTags(config, b, filter, tokens)
					logrus.Infof("Tagged `%s.%s` in %s\n", resource, resourceIdentifier, file.Name())
				} else {
					logrus.Warnf("Resource `%s.%s` in %s isn't taggable\n", resource, resourceIdentifier, file.Name())
				}
			}

			// Write the modified AST back to the .tf file
			if err := ioutil.WriteFile(filepath.Join(dir, file.Name()), []byte(f.Bytes()), os.ModePerm); err != nil {
				panic(err)
			}

			logrus.Debugf("Tags added to %s\n", file.Name())
		}
	}
}

func getFilterComment(b *hclwrite.Block) string {
	var filter string

	// Loop through every line in the Terraform resource
	for _, attr := range b.Body().BuildTokens(nil) {
		// If a line contains the tftag comment, extract anything after `:`
		if bytes.Contains(attr.Bytes, []byte(filterCommentPrefix)) {

			filterComment := string(attr.Bytes)

			tagPos := strings.Index(filterComment, filterCommentPrefix)

			if tagPos != -1 {
				value := filterComment[tagPos+len(filterCommentPrefix):]
				filter = value
			}
		}
	}
	return filter
}

// setTags sets the 'tags' attribute in the specified HCL block using the tags from the given tftag configuration.
func setTags(config model.Config, b *hclwrite.Block, filter string, tokens []*hclwrite.Token) {
	matched := false
	for _, tfTagConfig := range config.Config {
		// Check if the filter matches the tftag config item
		if strings.TrimSpace(tfTagConfig.Type) == strings.TrimSpace(filter) {
			matched = true
			for key, val := range tfTagConfig.Tags {
				identToken := hclwrite.Token{
					Type:         hclsyntax.TokenIdent,
					Bytes:        []byte(key),
					SpacesBefore: 0,
				}
				tokens = append(tokens, &identToken)

				equalToken := hclwrite.Token{
					Type:         hclsyntax.TokenEqual,
					Bytes:        []byte("="),
					SpacesBefore: 1,
				}
				tokens = append(tokens, &equalToken)

				oQuoteToken := hclwrite.Token{
					Type:         hclsyntax.TokenOQuote,
					Bytes:        []byte("\""),
					SpacesBefore: 1,
				}
				tokens = append(tokens, &oQuoteToken)

				valToken := hclwrite.Token{
					Type:         hclsyntax.TokenQuotedLit,
					Bytes:        []byte(val),
					SpacesBefore: 0,
				}
				tokens = append(tokens, &valToken)

				cQuoteToken := hclwrite.Token{
					Type:         hclsyntax.TokenCQuote,
					Bytes:        []byte("\""),
					SpacesBefore: 1,
				}
				tokens = append(tokens, &cQuoteToken)

				newLineToken := hclwrite.Token{
					Type:         hclsyntax.TokenNewline,
					Bytes:        []byte("\n"),
					SpacesBefore: 0,
				}
				tokens = append(tokens, &newLineToken)
			}

			b.Body().SetAttributeRaw("tags", tokens)
			break // Exit loop after first matching config item
		}
	}
	if !matched {
		ctyTags := make(map[string]cty.Value)
		for _, tfTagConfig := range config.Config {
			for key, val := range tfTagConfig.Tags {
				ctyTags[key] = cty.StringVal(val)
			}
		}
		b.Body().SetAttributeValue("tags", cty.ObjectVal(ctyTags))
	}
}
