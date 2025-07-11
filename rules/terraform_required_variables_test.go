package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformRequredVariables(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Config   string
		Expected helper.Issues
	}{
		{
			Name: "module with complete required variables and correct attribute.",
			Content: `
variable "cloud_creds" {
  sensitive = true
  type      = string
}

variable "module_tmpl" {
  type      = string
}

variable "module_info" {
  type      = string
}
`,
			Config:   testTerraformRequiredVariablesConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "module with complete required variables, correct attributes and complete additional required custom variables.",
			Content: `
variable "additional_required_var" {
  type = string
}
`,
			Config: `
rule "terraform_required_variables" {
  enabled       = true
  required_vars = ["additional_required_var"]
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "module with no required variables.",
			Content: `
variable "my_variable" {
  type = string
}
`,
			Config: testTerraformRequiredVariablesConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredVariables(),
					Message: "required variable(s) not declared: cloud_creds, module_info, module_tmpl",
					Range: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1},
						End:   hcl.Pos{Line: 1, Column: 1},
					},
				},
			},
		},
		{
			Name: "module with incomplete required variables. (no module_info), and no sensitive attribute on `cloud_creds`.",
			Content: `
variable "cloud_creds" {
  type = string
}

variable "module_tmpl" {
  type = string
}
`,
			Config: testTerraformRequiredVariablesConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredVariables(),
					Message: "required variable(s) not declared: module_info",
					Range: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1},
						End:   hcl.Pos{Line: 1, Column: 1},
					},
				},
				{
					Rule:    NewTerraformRequiredVariables(),
					Message: "variable `cloud_creds` is missing the `sensitive` attribute",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 23},
					},
				},
			},
		},
		{
			Name: "module with incomplete required variables. (no module_info & module_tmpl), but with sensitive attribute on `cloud_creds`.",
			Content: `
variable "cloud_creds" {
  sensitive = true
  type      = string
}
`,
			Config: testTerraformRequiredVariablesConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredVariables(),
					Message: "required variable(s) not declared: module_info, module_tmpl",
					Range: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1},
						End:   hcl.Pos{Line: 1, Column: 1},
					},
				},
			},
		},
		{
			Name: "module with complete required variables, but no senstitive attribute",
			Content: `
variable "cloud_creds" {
  type      = string
}

variable "module_tmpl" {
  type      = string
}

variable "module_info" {
  type      = string
}
`,
			Config: testTerraformRequiredVariablesConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredVariables(),
					Message: "variable `cloud_creds` is missing the `sensitive` attribute",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 23},
					},
				},
			},
		},
		{
			Name: "module with complete required variables, correct attributes and incomplete additional required custom variables.",
			Content: `
variable "cloud_creds" {
  sensitive = true
  type      = string
}
`,
			Config: `
rule "terraform_required_variables" {
  enabled       = true
  required_vars = ["additional_required_var"]
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredVariables(),
					Message: "required variable(s) not declared: additional_required_var",
					Range: hcl.Range{
						Filename: "",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 1},
					},
				},
			},
		},
		{
			Name: "variable cloud_creds have invalid sensitive placement and value.",
			Content: `
variable "cloud_creds" {
  type      = string
  sensitive = false
}
`,
			Config: `
rule "terraform_required_variables" {
  enabled       = true
  required_vars = ["cloud_creds"]
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredVariables(),
					Message: "variable `cloud_creds` must have `sensitive = true` attribute defined",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 20},
					},
				},
				{
					Rule:    NewTerraformRequiredVariables(),
					Message: "variable `cloud_creds` must place `sensitive = true` as first parameter after variable definition",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 20},
					},
				},
			},
		},
	}

	rule := NewTerraformRequiredVariables()
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"main.tf":     test.Content,
				".tflint.hcl": test.Config,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, test.Expected, runner.Issues)
		})
	}
}

const testTerraformRequiredVariablesConfig = `
rule "terraform_required_variables" {
  enabled = true
}
`
