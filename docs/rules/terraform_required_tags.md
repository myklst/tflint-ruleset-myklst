# terraform_required_tags

This rule checks whether all Terraform resources with `tags` attribute had included the required tag keys as defined in
the rule configuration. It will perform the checking even when the value is exact value(object/list), using local variable,
using terraform function `merge()` or `concat()` together with the local variable. Additionally, for AWS resources, it
enforces the presence of a `Name` tag. Unsupported expressions or function calls in tags will be ignored.

## Configuration

| Name               | Default                                                                                           | Value          |
| ------------------ | ------------------------------------------------------------------------------------------------- | -------------- |
| enabled            | true                                                                                              | Bool           |
| tags               | ["brand", "env", "project", "devops_project_kind", "devops_project_group", "devops_project_name"] | List of string |
| excluded_resources | []                                                                                                | List of string |

#### `tags`

The `tags` option defines the list of tags that needs to be included for any resources that has the `tags` block.
Defaults to the following list:

```hcl
tags = [
  "brand",
  "env",
  "project",
  "devops_project_kind",
  "devops_project_group",
  "devops_project_name",
]
```

#### `excluded_resources`

The `excluded_resources` option defines the list of resources to be ignored in ths rule checking. There will be two ways
to define the excluding resources,
- with only resource type: '\${resource_type}'. This configuration will skip checking for all resources 'aws_s3_bucket'.
For example,
  ```hcl
  rule "terraform_required_tags" {
    enabled            = true
    excluded_resources = ["aws_s3_bucket"]
  }
  ```
- with resource type and label: '\${resource_type}.\${resource_label}'. This configuration will skip checking for resource
'aws_s3_bucket.my_bucket', regardless of it's count/for_each. For example,
  ```hcl
  rule "terraform_required_tags" {
    enabled            = true
    excluded_resources = ["aws_s3_bucket.my_bucket"]
  }
  ```

## Example

### Rule configuration

```hcl
rule "terraform_required_tags" {
  enabled            = true
  tags               = ["example_tag1", "example_tag2", "example_tag3"]
}
```

### Sample terraform source file

```hcl
resource "my_resource" "my_resource_name" {
  name = "test"

  tags = {
    example_tag1 = "value1"
    example_tag2 = "value2"
  }
}
```

```
$ tflint
1 issue(s) found:

Warning: my_resource 'my_resource_name' is missing required tags: [example_tag3] (terraform_required_tags)

  on main.tf line 4:
   4:   tags = {
   5:     example_tag1 = "value1"
   6:     example_tag2 = "value2"
   7:   }

Reference: https://github.com/myklst/tflint-ruleset-myklst/docs/rules/terraform_required_tags.md
```

## Disable for specificed resources

### Rule configuration

```hcl
rule "terraform_required_tags" {
  enabled            = true
  tags               = ["example_tag1", "example_tag2", "example_tag3"]
  excluded_resources = ["my_excluded_resource"]
}
```

### Sample terraform source file

```hcl
// resource "my_excluded_resource" will not be enforced
resource "my_excluded_resource" "my_resource_name" {
  name = "test"

  tags = {
    example_tag1 = "value1"
    example_tag2 = "value2"
  }
}
```

## Direct use of local `tags` variable

### Rule configuration

```hcl
rule "terraform_required_tags" {
  enabled = true
  tags               = ["example_tag1", "example_tag2", "example_tag3"]
}
```

### Sample terraform source file

```hcl
locals {
  tags = {
    example_tag1 = "value1"
    example_tag2 = "value2"
  }
}

resource "my_excluded_resource" "my_resource_name" {
  name = "test"

  tags = local.tags
}
```

```hcl
$ tflint
1 issue(s) found:

Warning: my_excluded_resource 'my_resource_name' is missing required tags: [example_tag3] (terraform_required_tags)

  on main.tf line 11:
  11:   tags = local.tags

Reference: https://github.com/myklst/tflint-ruleset-myklst/docs/rules/terraform_required_tags.md
```

## Usage of function call `merge` with locals `tags`

### Rule configuration

```hcl
rule "terraform_required_tags" {
  enabled = true
  tags    = ["example_tag1", "example_tag2", "example_tag3"]
}
```

### Sample terraform source file

```hcl
locals {
  tags = {
    example_tag1 = "value1"
    example_tag2 = "value2"
  }
}

resource "my_excluded_resource" "my_resource_name" {
  name = "test"

  tags = merge(local.tags, {
    example_tag3 = "value3"
  })
}
```
