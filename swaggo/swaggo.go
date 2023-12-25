package swaggo

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/swaggo/swag"
)

const (
	// CamelCase indicates using CamelCase strategy for struct field.
	CamelCase = "camelcase"

	// PascalCase indicates using PascalCase strategy for struct field.
	PascalCase = "pascalcase"

	// SnakeCase indicates using SnakeCase strategy for struct field.
	SnakeCase = "snakecase"

	idAttr                  = "@id"
	acceptAttr              = "@accept"
	produceAttr             = "@produce"
	paramAttr               = "@param"
	successAttr             = "@success"
	failureAttr             = "@failure"
	responseAttr            = "@response"
	headerAttr              = "@header"
	tagsAttr                = "@tags"
	routerAttr              = "@router"
	summaryAttr             = "@summary"
	deprecatedAttr          = "@deprecated"
	securityAttr            = "@security"
	titleAttr               = "@title"
	conNameAttr             = "@contact.name"
	conURLAttr              = "@contact.url"
	conEmailAttr            = "@contact.email"
	licNameAttr             = "@license.name"
	licURLAttr              = "@license.url"
	versionAttr             = "@version"
	descriptionAttr         = "@description"
	descriptionMarkdownAttr = "@description.markdown"
	secBasicAttr            = "@securitydefinitions.basic"
	secAPIKeyAttr           = "@securitydefinitions.apikey"
	secApplicationAttr      = "@securitydefinitions.oauth2.application"
	secImplicitAttr         = "@securitydefinitions.oauth2.implicit"
	secPasswordAttr         = "@securitydefinitions.oauth2.password"
	secAccessCodeAttr       = "@securitydefinitions.oauth2.accesscode"
	tosAttr                 = "@termsofservice"
	extDocsDescAttr         = "@externaldocs.description"
	extDocsURLAttr          = "@externaldocs.url"
	xCodeSamplesAttr        = "@x-codesamples"
	scopeAttrPrefix         = "@scope."
)

func getMarkdownForTag(tagName string, dirPath string) ([]byte, error) {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()

		if !strings.Contains(fileName, ".md") {
			continue
		}

		if strings.Contains(fileName, tagName) {
			fullPath := filepath.Join(dirPath, fileName)

			commentInfo, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("Failed to read markdown file %s error: %s ", fullPath, err)
			}

			return commentInfo, nil
		}
	}

	return nil, fmt.Errorf("Unable to find markdown file for tag %s in the given directory", tagName)
}

// ParseComment parses comment for given comment string and returns error if error occurs.
func ParseComment(operation *swag.Operation, comment string, astFile *ast.File, markDownFileDir string) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "/"))
	if len(commentLine) == 0 {
		return nil
	}

	fields := FieldsByAnySpace(commentLine, 2)
	attribute := fields[0]
	lowerAttribute := strings.ToLower(attribute)
	var lineRemainder string
	if len(fields) > 1 {
		lineRemainder = fields[1]
	}
	switch lowerAttribute {
	case descriptionAttr:
		operation.ParseDescriptionComment(lineRemainder)
	case descriptionMarkdownAttr:
		commentInfo, err := getMarkdownForTag(lineRemainder, markDownFileDir)
		if err != nil {
			return err
		}

		operation.ParseDescriptionComment(string(commentInfo))
	case summaryAttr:
		operation.Summary = lineRemainder
	case idAttr:
		operation.ID = lineRemainder
	case tagsAttr:
		operation.ParseTagsComment(lineRemainder)
	case acceptAttr:
		return operation.ParseAcceptComment(lineRemainder)
	case produceAttr:
		return operation.ParseProduceComment(lineRemainder)
	case paramAttr:
		return operation.ParseParamComment(lineRemainder, astFile)
	case successAttr, failureAttr, responseAttr:
		return operation.ParseResponseComment(lineRemainder, astFile)
	case headerAttr:
		return operation.ParseResponseHeaderComment(lineRemainder, astFile)
	case routerAttr:
		return operation.ParseRouterComment(lineRemainder)
	case securityAttr:
		return operation.ParseSecurityComment(lineRemainder)
	case deprecatedAttr:
		operation.Deprecate()
	case xCodeSamplesAttr:
		return operation.ParseCodeSample(attribute, commentLine, lineRemainder)
	default:
		return operation.ParseMetadata(attribute, lowerAttribute, lineRemainder)
	}

	return nil
}

// FieldsFunc split a string s by a func splitter into max n parts
func FieldsFunc(s string, f func(rune2 rune) bool, n int) []string {
	// A span is used to record a slice of s of the form s[start:end].
	// The start index is inclusive and the end index is exclusive.
	type span struct {
		start int
		end   int
	}
	spans := make([]span, 0, 32)

	// Find the field start and end indices.
	// Doing this in a separate pass (rather than slicing the string s
	// and collecting the result substrings right away) is significantly
	// more efficient, possibly due to cache effects.
	start := -1 // valid span start if >= 0
	for end, rune := range s {
		if f(rune) {
			if start >= 0 {
				spans = append(spans, span{start, end})
				// Set start to a negative value.
				// Note: using -1 here consistently and reproducibly
				// slows down this code by a several percent on amd64.
				start = ^start
			}
		} else {
			if start < 0 {
				start = end
				if n > 0 && len(spans)+1 >= n {
					break
				}
			}
		}
	}

	// Last field might end at EOF.
	if start >= 0 {
		spans = append(spans, span{start, len(s)})
	}

	// Create strings from recorded field indices.
	a := make([]string, len(spans))
	for i, span := range spans {
		a[i] = s[span.start:span.end]
	}
	return a
}

// FieldsByAnySpace split a string s by any space character into max n parts
func FieldsByAnySpace(s string, n int) []string {
	return FieldsFunc(s, unicode.IsSpace, n)
}
