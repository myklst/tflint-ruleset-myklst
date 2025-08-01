package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformMetaArguments(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "source only in module",
			Content: `
module "my_module" {
  source = "./my-module/"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source and attributes in module",
			Content: `
module "my_module" {
  source = "./my-module/"

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source and attributes in module with comment",
			Content: `
module "my_module" {
  # I'm a comment.
  source = "./my-module/"

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source, count and attributes in module",
			Content: `
module "my_module" {
  source = "./my-module/"

  count = 3

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source, count and attributes in module with comments",
			Content: `
module "my_module" {
  # I'm first comment.
  source = "./my-module/"

  # I'm second comment.
  count = 3

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source, for_each and attributes in module",
			Content: `
module "my_module" {
  source = "./my-module/"

  for_each = {}

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source, for_each and attributes in module with comments",
			Content: `
module "my_module" {
  # I'm first comment.
  source = "./my-module/"

  # I'm second comment.
  for_each = {}

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source, count, providers and attributes in module",
			Content: `
module "my_module" {
  source = "./my-module/"

  count = {}

  providers = {}

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source, count, providers and attributes in module with comments",
			Content: `
module "my_module" {
  # I'm first comment.
  source = "./my-module/"

  # I'm second comment.
  count = {}

  # I'm third comment.
  providers = {}

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source, for_each, providers and attributes in module",
			Content: `
module "my_module" {
  source = "./my-module/"

  for_each = {}

  providers = {}

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source, for_each, providers and attributes in module with comments",
			Content: `
module "my_module" {
  # I'm first comment.
  source = "./my-module/"

  # I'm second comment.
  for_each = {}

  # I'm third comment.
  providers = {}

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "count only in resource",
			Content: `
resource "foo" "my_resource" {
  count = 3
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "for_each only in resource",
			Content: `
resource "foo" "my_resource" {
  for_each = {}
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "provider only in resource",
			Content: `
resource "foo" "my_resource" {
  provider = foo.default
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "count and attributes in resource with comment",
			Content: `
resource "foo" "my_resource" {
  # I'm a comment.
  count = 3

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "for_each and attributes in resource",
			Content: `
resource "foo" "my_resource" {
  for_each = {}

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "for_each and attributes in resource with comment",
			Content: `
resource "foo" "my_resource" {
  # I'm a comment.
  for_each = {}

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "provider and attributes in resource",
			Content: `
resource "foo" "my_resource" {
  provider = foo.default

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "provider and attributes in resource with comment",
			Content: `
resource "foo" "my_resource" {
  # I'm a comment.
  provider = foo.default

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "count, provider and attributes in resource",
			Content: `
resource "foo" "my_resource" {
  count = 3

  provider = foo.default

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "count, provider and attributes in resource with comments",
			Content: `
resource "foo" "my_resource" {
  # I'm a comment.
  count = 3

  provider = foo.default

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "count, provider and attributes in data source",
			Content: `
data "foo" "my_resource" {
  count = 3

  provider = foo.default

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "count, provider and attributes in data source with comments",
			Content: `
data "foo" "my_resource" {
  # I'm first comment.
  count = 3

  # I'm second comment.
  provider = foo.default

  name = "my name"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "lifecycle only in resource",
			Content: `
resource "foo" "my_resource" {
  lifecycle {}
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "lifecycle and attributes in resource",
			Content: `
resource "foo" "my_resource" {
  name = "my name"

  lifecycle {}
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "lifecycle and attributes in resource with comment",
			Content: `
resource "foo" "my_resource" {
  name = "my name"

  # I'm a comment.
  lifecycle {}
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "count, provider, lifecycle and attributes in resource",
			Content: `
resource "foo" "my_resource" {
  count = 3

  provider = foo.default

  name = "my name"

  lifecycle {}
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "count, provider, lifecycle and attributes in resource with comments",
			Content: `
resource "foo" "my_resource" {
  # I'm first comment.
  count = 3

  # I'm second comment.
  provider = foo.default

  name = "my name"

  # I'm third comment.
  lifecycle {}
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "source and attributes in module, invalid arrangement",
			Content: `
module "my_module" {

  source = "./my-module/"

  name = "my name"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "module 'my_module' has invalid 'source' meta argument arrangement",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 26},
					},
				},
			},
		},
		{
			Name: "source, count and attributes in module, invalid arrangement",
			Content: `
module "my_module" {
  count = 3

  source = "./my-module/"

  name = "my name"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "module 'my_module' has invalid 'source' meta argument arrangement",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 26},
					},
				},
			},
		},
		{
			Name: "count, provider and attributes in resource, invalid arrangement",
			Content: `
resource "foo" "my_resource" {
  provider = foo.default

  count = 3

  name = "my name"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "resource 'foo.my_resource' has invalid 'count' meta argument arrangement",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 12},
					},
				},
			},
		},
		{
			Name: "count, provider, lifecycle and attributes in resource, invalid arrangement",
			Content: `
resource "foo" "my_resource" {
  count = 3

  provider = foo.default

  lifecycle {}

  name = "my name"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "resource 'foo.my_resource' has invalid 'lifecycle' meta argument arrangement",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 12},
					},
				},
			},
		},
		{
			Name: "source, count, providers and attributes in module, invalid arrangement",
			Content: `
module "my_module" {
  source = "./my-module/"

  providers = {}

  count = 3

  name = "my name"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "module 'my_module' has invalid 'count' meta argument arrangement",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 12},
					},
				},
			},
		},
		{
			Name: "source and attributes in module, missing new line",
			Content: `
module "my_module" {
  source = "./my-module/"
  name = "my name"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "module 'my_module' has missing new line after 'source' meta argument",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 26},
					},
				},
			},
		},
		{
			Name: "source, count and attributes in module, missing new line",
			Content: `
module "my_module" {
  source = "./my-module/"

  count = 3
  name = "my name"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "module 'my_module' has missing new line after 'count' meta argument",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 12},
					},
				},
			},
		},
		{
			Name: "source, count, providers and attributes in module, missing new line",
			Content: `
module "my_module" {
  source = "./my-module/"

  count = 3

  providers = {}
  name = "my name"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "module 'my_module' has missing new line after 'providers' meta argument",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 17},
					},
				},
			},
		},
		{
			Name: "lifecycle and attributes in resource, missing new line",
			Content: `
resource "foo" "my_resource" {
  name = "my name"
  lifecycle {}
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "resource 'foo.my_resource' has missing new line before 'lifecycle' meta argument",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 12},
					},
				},
			},
		},
		{
			Name: "lifecycle and attributes in resource with comment, missing new line",
			Content: `
resource "foo" "my_resource" {
  name = "my name"
  # I'm a comment.
  lifecycle {}
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArguments(),
					Message: "resource 'foo.my_resource' has missing new line before 'lifecycle' meta argument",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 12},
					},
				},
			},
		},
	}

	rule := NewTerraformMetaArguments()
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, test.Expected, runner.Issues)
		})
	}
}
