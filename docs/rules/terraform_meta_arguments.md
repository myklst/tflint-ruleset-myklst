# terraform_meta_arguments

Check the sequences and format of `source`, `count`, `for_each`, `providers` and
`provider` meta-arguments in Terraform `module`, `resource` and `data source`.
Any comment lines before the meta-arguments are allowed, but an extra newline must
be placed under each meta-arguments except meta-argument `lifecycle`. If there are
no other attributes before/after the meta-arguments, then new line condition is
not checked.

## Terraform `module`

### Format

- Check the beginning arguments sequences in Terraform modules as the following:
  1. `module` definition _(end with newline)_
  2. `source` _(end with newline)_
  3. -- if `count` or `for_each` exist --
     1. `count`/`for_each` _(end with newline)_
     2. _(extra newline)_
  4. -- if `providers` exist --
     1. `providers` _(end with newline)_
     2. _(extra newline)_
  5. other attributes/blocks

### Valid example

```hcl
module "alicloud_ecs_instances" {
  source = "./alicloud-ecs-instance/"

  count = 3

  providers = {
    alicloud = alicloud.ecs
  }

  # ...
}
```

```hcl
module "aws_ec2_instance" {
  source = "./aws-ec2-instance/"

  providers = {
    aws = aws.ec2
  }

  # ...
}
```

```hcl
module "alicloud_ecs_instances" {
  # I'm first comment.
  source = "./alicloud-ecs-instance/"

  # I'm second comment.
  count = 3

  providers = {
    alicloud = alicloud.ecs
  }

  # ...
}
```

## Terraform `resource` and `data source`

### Format

- Check the beginning arguments sequences in Terraform resources and data sources
  as the following:
  1. `resource` definition _(end with newline)_
  2. -- if `count` or `for_each` exist --
     1. `count`/`for_each` _(end with newline)_
     2. _(extra newline)_
  3. -- if `provider` exist --
     1. `provider` _(end with newline)_
     2. _(extra newline)_
  4. other attributes/blocks
- `lifecycle{}` meta-argument
  - block must have extra newline in one line above.
  - block must be placed as last block at the end of the resource without extra new lines.

## Valid example

```hcl
resource "aws_ec2_instance" "my_instance" {
  count = 3

  provider = aws.ec2

  # ...

  lifecycle {
    create_before_destroy = true
  }
}
```

```hcl
resource "aws_ec2_instance" "my_instance" {
  # ...

  lifecycle {
    create_before_destroy = true
  }
}
```

```hcl
data "aws_ec2_instance" "my_instance" {
  count = 3

  provider = aws.ec2

  # ...

  lifecycle {
    create_before_destroy = true
  }
}
```

```hcl
resource "aws_ec2_instance" "my_instance" {
  # I'm first comment.
  count = 3

  # I'm second comment.
  provider = aws.ec2

  # ...

  # I'm third comment.
  lifecycle {
    create_before_destroy = true
  }
}
```
