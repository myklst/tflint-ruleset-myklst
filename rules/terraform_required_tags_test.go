package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformRequiredTags(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Config   string
		Expected helper.Issues
	}{
		{
			Name: "resource without any tags block.",
			Content: `
resource "my_resource" "my_resource_name" {
  name = "test"
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "resource with the correct required tags.",
			Content: `
resource "my_resource" "my_resource_name" {
  name = "test"

  tags = {
    my_required_tag = "my_tag"
  }
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "resource with the missing required tags.",
			Content: `

resource "my_resource" "my_resource_name" {
  name = "test"

  tags = {
    my_incorrect_tag = "my_tag"
  }
}
`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'my_resource.my_resource_name' is missing required tags: ['my_required_tag']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 6, Column: 10},
						End:      hcl.Pos{Line: 8, Column: 4},
					},
				},
			},
		},
		{
			Name: "resource that is excluded from the rule with resource type",
			Content: `
resource "my_excluded_resource" "my_excluded_resource_name" {
  name = "test"

  tags = {
    my_incorrect_tag = "my_tag"
  }
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "resource that is excluded from the rule with resource type and label.",
			Content: `
resource "my_excluded_resource_v2" "my_resource" {
  name = "test"

  tags = {
    my_incorrect_tag = "my_tag"
  }
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "resource that is excluded from the rule with resource type and label, but mismatched label.",
			Content: `
resource "my_excluded_resource_v2" "not_my_resource" {
  name = "test"

  tags = {
    my_incorrect_tag = "my_tag"
  }
}
`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'my_excluded_resource_v2.not_my_resource' is missing required tags: ['my_required_tag']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 10},
						End:      hcl.Pos{Line: 7, Column: 4},
					},
				},
			},
		},
		{
			Name: "aws resource with the correct required tags.",
			Content: `
resource "aws_resource" "my_resource_name" {
  name = "test"

  tags = {
    my_required_tag = "my_tag"
    Name            = "test"
  }
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "aws resource with correct required tags but missing 'Name' tag.",
			Content: `
resource "aws_resource" "my_resource_name" {
  name = "test"

  tags = {
   my_required_tag = "my_tag"
  }
}
		`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "aws resources must have 'Name' tag: 'aws_resource.my_resource_name'",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 10},
						End:      hcl.Pos{Line: 7, Column: 4},
					},
				},
			},
		},
		{
			Name: "aws resource with missing required tags and no 'Name' tags.",
			Content: `
resource "aws_s3_bucket" "my_resource_name" {
  name = "test"

  tags = {
    not_required_tag = "my_tag"
  }
}
`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'aws_s3_bucket.my_resource_name' is missing required tags: ['my_required_tag']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 10},
						End:      hcl.Pos{Line: 7, Column: 4},
					},
				},
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "aws resources must have 'Name' tag: 'aws_s3_bucket.my_resource_name'",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 10},
						End:      hcl.Pos{Line: 7, Column: 4},
					},
				},
			},
		},
		{
			Name: "aws resource with 'Name' tag, merged with local variable `tags` with the correct required tags.",
			Content: `
locals {
  tags = {
    brand           = "foo"
    project         = "bar"
    env             = "dev"
    my_required_tag = "test"
  }
}

resource "aws_instance" "aws_resource_name" {
  tags = merge(local.tags, {
    Name = "aws_instance-test"
  })
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "aws resource with 'Name' tag and the correct required tags, merged with nested local variable `tags` with function call.",
			Content: `
variable "my_tags" {
  type = map(string)
}

locals {
  tags = merge(var.my_tags, {
    brand           = "foo"
    project         = "bar"
    env             = "dev"
    my_required_tag = "test"
  })
}

resource "aws_instance" "aws_resource_name" {
  tags = merge(local.tags, {
    Name = "aws_instance-test"
  })
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "aws resource with 'Name' tag and the correct required tags, merged with nested local variable `tags` with single nested local variable.",
			Content: `
locals {
  my_tags = {
    brand           = "foo"
    project         = "bar"
    env             = "dev"
    my_required_tag = "test"
  }
  tags = local.my_tags
}

resource "aws_instance" "aws_resource_name" {
  tags = merge(local.tags, {
    Name = "aws_instance-test"
  })
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "aws resource with 'Name' tag and the correct required tags, merged with nested local variable `tags` with function call and single nested local variable.",
			Content: `
locals {
  my_tags = {
    brand           = "foo"
    project         = "bar"
    env             = "dev"
  }
  tags = merge(local.my_tags, {
    my_required_tag = "test"
  })
}

resource "aws_instance" "aws_resource_name" {
  tags = merge(local.tags, {
    Name = "aws_instance-test"
  })
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "aws resource with 'Name' tag and the correct required tags, merged with nested local variable `tags` with multiple function calls and nested local variables.",
			Content: `
locals {
  our_tags = {
    brand           = "foo"
    project         = "bar"
  }
  my_tags = merge(local.our_tags , {
    my_required_tag = "test"
  })
  tags = merge(local.my_tags, {
    env             = "dev"
  })
}

resource "aws_instance" "aws_resource_name" {
  tags = merge(local.tags, {
    Name = "aws_instance-test"
  })
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "aws resource with no 'Name' tag, merged with local variable `tags` with the missing required keys.",
			Content: `
locals {
  tags = {
    foo = "bar"
    bar = "foo"
  }
}

resource "aws_resource" "my_resource_name" {
  tags = merge(local.tags, {
    testTag = "test"
  })
}
`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'aws_resource.my_resource_name' is missing required tags: ['my_required_tag']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 10, Column: 10},
						End:      hcl.Pos{Line: 12, Column: 5},
					},
				},
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "aws resources must have 'Name' tag: 'aws_resource.my_resource_name'",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 10, Column: 10},
						End:      hcl.Pos{Line: 12, Column: 5},
					},
				},
			},
		},
		{
			Name: "resource using local variable as tags, and local variable `tags` with the correct default required tags.",
			Content: `
locals {
  tags = {
    brand                = "foo"
    env                  = "bar"
    project              = "dev"
    devops_project_kind  = "foo"
    devops_project_group = "bar"
    devops_project_name  = "dev"
  }
}

resource "my_resource" "my_resource_name" {
  tags = local.tags
}
`,
			Config: `
rule "terraform_required_tags" {
  enabled = true
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "resource using local variable as tags, and local variable `tags` with the missing default required tags.",
			Content: `
locals {
  tags = {
    brand                = "foo"
    env                  = "bar"
    project              = "dev"
  }
}

resource "my_resource" "my_resource_name" {
  tags = local.tags
}
`,
			Config: `
rule "terraform_required_tags" {
  enabled = true
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'my_resource.my_resource_name' is missing required tags: ['devops_project_kind', 'devops_project_group', 'devops_project_name']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 11, Column: 10},
						End:      hcl.Pos{Line: 11, Column: 20},
					},
				},
			},
		},
		{
			Name: "resource with correct required tags (list of string).",
			Content: `
resource "foo" "my_resource" {
  tags = ["my_required_tag:dev"]
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "resource with correct required tags (list of string) when referring to local variable.",
			Content: `
locals {
  tags = ["my_required_tag:dev"]
}

resource "foo" "my_resource" {
  tags =  local.tags
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "resource with correct required tags (list of string) after concating local variable.",
			Content: `
locals {
  tags = ["my_required_tag:dev"]
}

resource "foo" "my_resource" {
  tags = concat(local.tags, ["my_tag_2:prod"])
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "resource with correct required tags (list of string) when concating other local variable.",
			Content: `
locals {
  tags = ["my_tag_2:dev"]
}

resource "foo" "my_resource" {
  tags = concat(local.tags, ["my_required_tag:prod"])
}
`,
			Config:   testTerraformRequiredTagsConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "resource with missing required tags (list of string).",
			Content: `
resource "foo" "my_resource" {
  tags = ["my_incorrect_tag:dev"]
}
`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'foo.my_resource' is missing required tags: ['my_required_tag']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 10},
						End:      hcl.Pos{Line: 3, Column: 34},
					},
				},
			},
		},
		{
			Name: "resource with missing required tags (list of string) when referring to local variable.",
			Content: `
locals {
  tags = ["my_incorrect_tag:dev"]
}

resource "foo" "my_resource" {
  tags =  local.tags
}
`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'foo.my_resource' is missing required tags: ['my_required_tag']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 7, Column: 11},
						End:      hcl.Pos{Line: 7, Column: 21},
					},
				},
			},
		},
		{
			Name: "resource with missing required tags (list of string) after concating local variable.",
			Content: `
locals {
  tags = ["my_incorrect_tag:dev"]
}

resource "foo" "my_resource" {
  tags = concat(local.tags, ["my_tag_2:prod"])
}
`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'foo.my_resource' is missing required tags: ['my_required_tag']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 7, Column: 10},
						End:      hcl.Pos{Line: 7, Column: 47},
					},
				},
			},
		},
		{
			Name: "resource with missing required tags (list of string) when concating other local variable.",
			Content: `
locals {
  tags = ["my_tag_2:dev"]
}

resource "foo" "my_resource" {
  tags = concat(local.tags, ["my_incorrect_tag:prod"])
}
`,
			Config: testTerraformRequiredTagsConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredTags(),
					Message: "resource 'foo.my_resource' is missing required tags: ['my_required_tag']",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 7, Column: 10},
						End:      hcl.Pos{Line: 7, Column: 55},
					},
				},
			},
		},
	}

	rule := NewTerraformRequiredTags()
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

const testTerraformRequiredTagsConfig = `
rule "terraform_required_tags" {
  enabled            = true

  tags               = ["my_required_tag"]
  excluded_resources = ["my_excluded_resource", "my_excluded_resource_v2.my_resource"]
}
`
