package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/myklst/tflint-ruleset-myklst/project"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformRequiredVariables checks whether variables have a type checked
type TerraformRequiredVariables struct {
	tflint.DefaultRule
}

type terraformRequiredVariablesConfig struct {
	RequiredVars []string `hclext:"required_vars,optional"`
}

// NewTerraformRequiredVariables returns a new rule
func NewTerraformRequiredVariables() *TerraformRequiredVariables {
	return &TerraformRequiredVariables{}
}

// Name returns the rule name
func (r *TerraformRequiredVariables) Name() string {
	return "terraform_required_variables"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformRequiredVariables) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformRequiredVariables) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformRequiredVariables) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether required_vars have been declared as variables within the module
func (r *TerraformRequiredVariables) Check(runner tflint.Runner) error {
	config := &terraformRequiredVariablesConfig{}

	if err := runner.DecodeRuleConfig(r.Name(), config); err != nil {
		return err
	}

	// Set default required variables if none are specified.
	if len(config.RequiredVars) == 0 {
		config.RequiredVars = []string{
			"cloud_creds",
			"module_info",
			"module_tmpl",
		}
	}

	variables, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{
							Name: "sensitive",
						},
					},
				},
			},
		},
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	declaredVars := make(map[string]bool)
	for _, block := range variables.Blocks {
		if len(block.Labels) > 0 {
			declaredVars[block.Labels[0]] = true
		}
	}

	var missingVars []string
	for _, requiredVar := range config.RequiredVars {
		if _, exists := declaredVars[requiredVar]; !exists {
			missingVars = append(missingVars, requiredVar)
		}
	}

	if len(missingVars) > 0 {
		err := runner.EmitIssue(
			r,
			fmt.Sprintf("required variable(s) not declared: %s", strings.Join(missingVars, ", ")),
			hcl.Range{
				Start: hcl.Pos{Line: 1, Column: 1},
				End:   hcl.Pos{Line: 1, Column: 1},
			},
		)
		if err != nil {
			return err
		}
	}

	// Check for "cloud_creds" variable and its "sensitive" attribute
	for _, variable := range variables.Blocks {
		if variable.Labels[0] == "cloud_creds" {
			sensitiveAttr, sensitiveExist := variable.Body.Attributes["sensitive"]
			// Check if "sensitive" attribute exist.
			if sensitiveExist {
				sensitiveValue, _ := sensitiveAttr.Expr.(*hclsyntax.LiteralValueExpr)
				// Check if "sensitive" attribute is placed under variable definition.
				if sensitiveAttr.Range.Start.Line != variable.DefRange.End.Line+1 {
					err := runner.EmitIssue(
						r,
						fmt.Sprintf("variable `%s` must place `sensitive = true` as first parameter after variable definition", variable.Labels[0]),
						sensitiveAttr.Range,
					)
					if err != nil {
						return err
					}
				}

				// Check if "sensitive" attribute value is `true`.
				if !sensitiveValue.Val.True() {
					err := runner.EmitIssue(
						r,
						fmt.Sprintf("variable `%s` must have `sensitive = true` attribute defined", variable.Labels[0]),
						sensitiveAttr.Range,
					)
					if err != nil {
						return err
					}
				}
			} else {
				err := runner.EmitIssue(
					r,
					fmt.Sprintf("variable `%s` is missing the `sensitive` attribute", variable.Labels[0]),
					variable.DefRange,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
