package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/myklst/tflint-ruleset-myklst/project"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformMetaArguments checks the sequences of meta arguments
type TerraformMetaArguments struct {
	tflint.DefaultRule
}

// NewTerraformMetaArguments returns a new rule
func NewTerraformMetaArguments() *TerraformMetaArguments {
	return &TerraformMetaArguments{}
}

// Name returns the rule name
func (r *TerraformMetaArguments) Name() string {
	return "terraform_meta_arguments"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformMetaArguments) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformMetaArguments) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformMetaArguments) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether variables have type
func (r *TerraformMetaArguments) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		fileContents, _, diags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "module",
					LabelNames: []string{"name"},
				},
				{
					Type:       "resource",
					LabelNames: []string{"type", "name"},
				},
				{
					Type:       "data",
					LabelNames: []string{"type", "name"},
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		for _, fileContent := range fileContents.Blocks {
			var content *hcl.BodyContent
			switch fileContent.Type {
			case "module":
				content, _, diags = fileContent.Body.PartialContent(&hcl.BodySchema{
					Attributes: []hcl.AttributeSchema{
						{Name: "source"},
						{Name: "count"},
						{Name: "for_each"},
						{Name: "providers"},
					},
				})
			case "resource", "data":
				content, _, diags = fileContent.Body.PartialContent(&hcl.BodySchema{
					Attributes: []hcl.AttributeSchema{
						{Name: "count"},
						{Name: "for_each"},
						{Name: "provider"},
					},
					Blocks: []hcl.BlockHeaderSchema{
						{
							Type:       "lifecycle",
							LabelNames: []string{},
						},
					},
				})
			}
			if diags.HasErrors() {
				return diags
			}

			// currentLine is used as a pointer to perform checking.
			// filename is used in function countCommentLine().
			// fileContentBody is used to check resource block ending line.
			var currentLine int
			filename := fileContent.DefRange.Filename
			fileContentBody := fileContent.Body.(*hclsyntax.Body)

			// Move pointer 'currentLine' to next line after 'module' definition and ignore comment lines.
			commentLines := r.countCommentLinesForward(file, filename, fileContent.DefRange.End.Line+1)
			currentLine = fileContent.DefRange.End.Line + commentLines + 1

			// Check meta argument 'source'.
			source, sourceExist := content.Attributes["source"]
			if sourceExist {
				if currentLine != source.Range.Start.Line {
					if err := runner.EmitIssue(
						r,
						fmt.Sprintf("%s '%s' has invalid 'source' meta argument arrangement", fileContent.Type, strings.Join(fileContent.Labels, ".")),
						source.Range,
					); err != nil {
						return err
					}
					continue
				}
				// Move pointer 'currentLine' to next line after 'source' meta argument and ignore comment lines.
				commentLines := r.countCommentLinesForward(file, filename, source.Range.End.Line+1)
				currentLine = source.Range.End.Line + commentLines + 1

				// Check new line after meta argument 'source'.
				// Ignore if next line is end of resource (there is no other attributes).
				if currentLine != fileContentBody.EndRange.Start.Line && !r.isNewLine(file, filename, currentLine) {
					if err := runner.EmitIssue(
						r,
						fmt.Sprintf("%s '%s' has missing new line after 'source' meta argument", fileContent.Type, strings.Join(fileContent.Labels, ".")),
						source.Range,
					); err != nil {
						return err
					}
					continue
				}

				// Move pointer 'currentLine' to next line after the new line and ignore comment lines.
				commentLines = r.countCommentLinesForward(file, filename, currentLine+1)
				currentLine = currentLine + commentLines + 1
			}

			// Check meta arguments 'count' or 'for_each'.
			// Only one of 'count' and 'for_each' are allowed in Terraform meta argument.
			count, countExist := content.Attributes["count"]
			forEach, forEachExist := content.Attributes["for_each"]
			if countExist {
				if currentLine != count.Range.Start.Line {
					if err := runner.EmitIssue(
						r,
						fmt.Sprintf("%s '%s' has invalid 'count' meta argument arrangement", fileContent.Type, strings.Join(fileContent.Labels, ".")),
						count.Range,
					); err != nil {
						return err
					}
					continue
				}
				// Move pointer 'currentLine' to next line after 'count' meta argument and ignore comment lines.
				commentLines := r.countCommentLinesForward(file, filename, count.Range.End.Line+1)
				currentLine = count.Range.End.Line + commentLines + 1

				// Check new line after meta argument 'count'.
				// Ignore if next line is end of resource (there is no other attributes).
				if currentLine != fileContentBody.EndRange.Start.Line && !r.isNewLine(file, filename, currentLine) {
					if err := runner.EmitIssue(
						r,
						fmt.Sprintf("%s '%s' has missing new line after 'count' meta argument", fileContent.Type, strings.Join(fileContent.Labels, ".")),
						count.Range,
					); err != nil {
						return err
					}
					continue
				}

				// Move pointer 'currentLine' to next line after the new line and ignore comment lines.
				commentLines = r.countCommentLinesForward(file, filename, currentLine+1)
				currentLine = currentLine + commentLines + 1
			} else if forEachExist {
				if currentLine != forEach.Range.Start.Line {
					if err := runner.EmitIssue(
						r,
						fmt.Sprintf("%s '%s' has invalid 'for_each' meta argument arrangement", fileContent.Type, fileContent.Labels[0]),
						forEach.Range,
					); err != nil {
						return err
					}
					continue
				}
				// Move pointer 'currentLine' to next line after 'for_each' meta argument and ignore comment lines.
				commentLines := r.countCommentLinesForward(file, filename, forEach.Range.End.Line+1)
				currentLine = forEach.Range.End.Line + commentLines + 1

				// Check new line after meta argument 'for_each'.
				// Ignore if next line is end of resource (there is no other attributes).
				if currentLine != fileContentBody.EndRange.Start.Line && !r.isNewLine(file, filename, currentLine) {
					if err := runner.EmitIssue(
						r,
						fmt.Sprintf("%s '%s' has missing new line after 'for_each' meta argument", fileContent.Type, strings.Join(fileContent.Labels, ".")),
						forEach.Range,
					); err != nil {
						return err
					}
					continue
				}

				// Move pointer 'currentLine' to next line after the new line and ignore comment lines.
				commentLines = r.countCommentLinesForward(file, filename, currentLine+1)
				currentLine = currentLine + commentLines + 1
			}

			// Check meta arguments 'providers' or 'provider'.
			var providerExist bool
			var provider *hcl.Attribute
			var placementErrMsg, newLineErrMsg string

			switch fileContent.Type {
			case "module":
				provider, providerExist = content.Attributes["providers"]
				placementErrMsg = fmt.Sprintf("%s '%s' has invalid 'providers' meta argument arrangement",
					fileContent.Type, strings.Join(fileContent.Labels, "."))
				newLineErrMsg = fmt.Sprintf("%s '%s' has missing new line after 'providers' meta argument",
					fileContent.Type, strings.Join(fileContent.Labels, "."))
			case "resource", "data": // Combined case since they have identical handling
				provider, providerExist = content.Attributes["provider"]
				placementErrMsg = fmt.Sprintf("%s '%s' has invalid 'provider' meta argument arrangement",
					fileContent.Type, strings.Join(fileContent.Labels, "."))
				newLineErrMsg = fmt.Sprintf("%s '%s' has missing new line after 'provider' meta argument",
					fileContent.Type, strings.Join(fileContent.Labels, "."))
			}

			if providerExist {
				if currentLine != provider.Range.Start.Line {
					if err := runner.EmitIssue(
						r,
						placementErrMsg,
						provider.Range,
					); err != nil {
						return err
					}
					continue
				}
				// Move pointer 'currentLine' to next line after 'provider' meta argument and ignore comment lines.
				commentLines := r.countCommentLinesForward(file, filename, provider.Range.End.Line+1)
				currentLine = provider.Range.End.Line + commentLines + 1

				// Check new line after meta argument 'provider'.
				// Ignore if next line is end of resource (there is no other attributes).
				if currentLine != fileContentBody.EndRange.Start.Line && !r.isNewLine(file, filename, currentLine) {
					if err := runner.EmitIssue(
						r,
						newLineErrMsg,
						provider.Range,
					); err != nil {
						return err
					}
					continue
				}
			}

			// Check meta argument 'lifecycle'.
			for _, contentBlock := range content.Blocks {
				if contentBlock.Type == "lifecycle" {
					// Check if newline exist one line before 'lifecycle' meta argument and ignore comment lines.
					// Ignore if next line is end of resource (there is no other attributes).
					commentLines := r.countCommentLinesBackward(file, filename, contentBlock.DefRange.Start.Line-1)
					checkLine := contentBlock.DefRange.Start.Line - commentLines - 1
					if checkLine != fileContent.DefRange.Start.Line && !r.isNewLine(file, filename, checkLine) {
						if err := runner.EmitIssue(
							r,
							fmt.Sprintf("%s '%s' has missing new line before 'lifecycle' meta argument", fileContent.Type, strings.Join(fileContent.Labels, ".")),
							contentBlock.DefRange,
						); err != nil {
							return err
						}
						continue
					}

					// Check if lifecycle is placed at the end of the module/resource/data source and ignore comment lines.
					lifeCycleBlockBody := contentBlock.Body.(*hclsyntax.Body)
					commentLines = r.countCommentLinesForward(file, filename, lifeCycleBlockBody.SrcRange.End.Line+1)
					checkLine = lifeCycleBlockBody.SrcRange.End.Line + commentLines + 1
					if checkLine != fileContentBody.EndRange.End.Line {
						if err := runner.EmitIssue(
							r,
							fmt.Sprintf("%s '%s' has invalid 'lifecycle' meta argument arrangement", fileContent.Type, strings.Join(fileContent.Labels, ".")),
							contentBlock.DefRange,
						); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func (r *TerraformMetaArguments) isNewLine(file *hcl.File, filename string, checkLine int) bool {
	tokens, _ := hclsyntax.LexConfig(file.Bytes, filename, hcl.Pos{Line: 1, Column: 1})

	for i, token := range tokens {
		// Empty new lines are expected to have two TokenNewLine continuously.
		if token.Range.End.Line == checkLine && token.Range.End.Column == 1 {
			if i+1 <= len(tokens) && token.Type == hclsyntax.TokenNewline && tokens[i+1].Type == hclsyntax.TokenNewline {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

func (r *TerraformMetaArguments) countCommentLinesForward(file *hcl.File, filename string, startingLine int) int {
	var commentLinesCount int
	tokens, _ := hclsyntax.LexConfig(file.Bytes, filename, hcl.Pos{Line: 1, Column: 1})

checking:
	for {
		for _, token := range tokens {
			if token.Range.Start.Line == startingLine+commentLinesCount {
				if token.Type == hclsyntax.TokenComment {
					commentLinesCount++
					continue checking
				} else {
					return commentLinesCount
				}
			}
		}
		return commentLinesCount
	}
}

func (r *TerraformMetaArguments) countCommentLinesBackward(file *hcl.File, filename string, startingLine int) int {
	var commentLinesCount int
	tokens, _ := hclsyntax.LexConfig(file.Bytes, filename, hcl.Pos{Line: 1, Column: 1})

checking:
	for {
		for _, token := range tokens {
			if token.Range.Start.Line == startingLine-commentLinesCount {
				if token.Type == hclsyntax.TokenComment {
					commentLinesCount++
					continue checking
				} else {
					return commentLinesCount
				}
			}
		}
		return commentLinesCount
	}
}
