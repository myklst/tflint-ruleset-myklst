package rules

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/myklst/tflint-ruleset-myklst/project"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

type TerraformRequiredTags struct {
	tflint.DefaultRule
}

// NewTerraformRequiredTags returns a new rule
func NewTerraformRequiredTags() *TerraformRequiredTags {
	return &TerraformRequiredTags{}
}

type terraformRequiredTagsConfig struct {
	Tags              []string `hclext:"tags,optional"`
	ExcludedResources []string `hclext:"excluded_resources,optional"`
}

// Name returns the rule name
func (r *TerraformRequiredTags) Name() string {
	return "terraform_required_tags"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformRequiredTags) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformRequiredTags) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformRequiredTags) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether resources have the required tags if applicable
func (r *TerraformRequiredTags) Check(runner tflint.Runner) error {
	config := &terraformRequiredTagsConfig{}

	if err := runner.DecodeRuleConfig(r.Name(), config); err != nil {
		return err
	}

	// Set default required tags if none are specified
	if len(config.Tags) == 0 {
		config.Tags = []string{
			"brand",
			"env",
			"project",
			"devops_project_kind",
			"devops_project_group",
			"devops_project_name",
		}
	}

	// Parse resources and check their `tags` blocks
	resources, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "tags"},
					},
				},
			},
		},
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		// If the resource is stated in excluded_resources, then ignore checking.
		if slices.Contains(config.ExcludedResources, resource.Labels[0]) || slices.Contains(config.ExcludedResources, fmt.Sprintf("%s.%s", resource.Labels[0], resource.Labels[1])) {
			continue
		}

		// If the resource do not have attribute "tags", then ignore checking.
		tagsAttr, tagsExist := resource.Body.Attributes["tags"]
		if !tagsExist {
			continue
		}

		// tagKeys is used to compare with required_tags to check any missing tags.
		var tagKeys []string
		tagKeys, err = r.traverseSearchExpr(runner, tagsAttr.Expr)
		if err != nil {
			return err
		}

		// Remove any duplicated keys if any
		tagKeys = slices.Compact(tagKeys)
		var missing []string
		for _, requiredTags := range config.Tags {
			if !slices.Contains(tagKeys, requiredTags) {
				missing = append(missing, requiredTags)
			}
		}

		// Output linting error if any missing tags are present
		if len(missing) > 0 {
			err := runner.EmitIssue(
				r,
				fmt.Sprintf("resource '%s.%s' is missing required tags: ['%s']", resource.Labels[0], resource.Labels[1], strings.Join(missing, "', '")),
				tagsAttr.Expr.Range(),
			)
			if err != nil {
				return err
			}
		}

		// If resource is AWS Cloud resource, check if `Name` tag key exists
		if r.isAwsResource(resource.Labels[0]) && !slices.Contains(tagKeys, "Name") {
			err := runner.EmitIssue(
				r,
				fmt.Sprintf("aws resources must have 'Name' tag: '%s.%s'", resource.Labels[0], resource.Labels[1]),
				tagsAttr.Expr.Range(),
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Function to determine whether resource has `aws_` prefix
func (r *TerraformRequiredTags) isAwsResource(resource string) bool {
	return strings.HasPrefix(resource, "aws_")
}

// This function will perform a deep traverse into every nested local variables used.
func (r *TerraformRequiredTags) traverseSearchExpr(runner tflint.Runner, expr hcl.Expression) ([]string, error) {
	var tagKeys []string
	// Check the value of tags and invoke different logics to evaluate.
	switch expr := expr.(type) {
	// Usage of function calls like merge(local.tags, { ... })
	case *hclsyntax.FunctionCallExpr:
		switch expr.Name {
		case "merge":
			for _, arg := range expr.Args {
				if traversal, ok := arg.(*hclsyntax.ScopeTraversalExpr); ok {
					// If the argument is a valid local variable invocation, then
					// evaluate the value and get the tag key.
					if localVarName, ok := r.extractLocalVarName(traversal); ok {
						localVarTagsKey, err := r.evaluateLocalVarTagsKey(runner, localVarName)
						if err != nil {
							return nil, err
						}
						tagKeys = slices.Concat(tagKeys, localVarTagsKey)
					}
				} else {
					// Otherwise, evaluate and extract keys.
					if err := runner.EvaluateExpr(arg, func(val cty.Value) error {
						tagKeys = slices.Concat(tagKeys, r.getTagsKey(val))
						return nil
					}, nil); err != nil {
						return nil, err
					}
				}
			}
		case "concat":
			for _, arg := range expr.Args {
				if traversal, ok := arg.(*hclsyntax.ScopeTraversalExpr); ok {
					// If the argument is a valid local variable invocation, then
					// evaluate the value and get the tag key.
					if localVarName, ok := r.extractLocalVarName(traversal); ok {
						localVarTagsKey, err := r.evaluateLocalVarTagsKey(runner, localVarName)
						if err != nil {
							return nil, err
						}
						tagKeys = slices.Concat(tagKeys, localVarTagsKey)
					}
				} else {
					// Otherwise, evaluate and extract keys.
					if err := runner.EvaluateExpr(arg, func(val cty.Value) error {
						tagKeys = slices.Concat(tagKeys, r.getTagsKey(val))
						return nil
					}, nil); err != nil {
						keys, ok := r.handleEvaluateTupleError(err, arg.(*hclsyntax.TupleConsExpr))
						if !ok {
							return nil, err
						} else {
							tagKeys = slices.Concat(tagKeys, keys)
						}
					}
				}
			}
		}

	// Direct use of local variable on tags
	// E.g. tags = local.tags
	case *hclsyntax.ScopeTraversalExpr:
		if localVarName, ok := r.extractLocalVarName(expr); ok {
			localVarTagsKey, err := r.evaluateLocalVarTagsKey(runner, localVarName)
			if err != nil {
				return nil, err
			}
			tagKeys = slices.Concat(tagKeys, localVarTagsKey)
		}

	// When it's actual list values in tags field.
	case *hclsyntax.TupleConsExpr:
		if err := runner.EvaluateExpr(expr, func(val cty.Value) error {
			tagKeys = slices.Concat(tagKeys, r.getTagsKey(val))
			return nil
		}, nil); err != nil {
			keys, ok := r.handleEvaluateTupleError(err, expr)
			if !ok {
				return nil, err
			} else {
				tagKeys = slices.Concat(tagKeys, keys)
			}
		}

	// When it's actual object values in tags field.
	case *hclsyntax.ObjectConsExpr:
		if err := runner.EvaluateExpr(expr, func(val cty.Value) error {
			tagKeys = slices.Concat(tagKeys, r.getTagsKey(val))
			return nil
		}, nil); err != nil {
			return nil, err
		}

	// Do nothing if unknown type.
	default:
	}
	return tagKeys, nil
}

// Extract the traversal expression to get the variable name, return false if it
// is not an valid local variable invocation.
// For example, a valid traversal expression to invoke local variable would be
// 'local.my_tags'.
func (r *TerraformRequiredTags) extractLocalVarName(traversal *hclsyntax.ScopeTraversalExpr) (string, bool) {
	if len(traversal.Traversal) >= 2 {
		rootName := traversal.Traversal[0].(hcl.TraverseRoot).Name
		if rootName == "local" {
			return traversal.Traversal[1].(hcl.TraverseAttr).Name, true
		}
	}
	return "", false
}

func (r *TerraformRequiredTags) evaluateLocalVarTagsKey(runner tflint.Runner, localVarName string) ([]string, error) {
	locals, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "locals",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: localVarName},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	var localTagKeys []string
	for _, block := range locals.Blocks {
		if localVarAttr, ok := block.Body.Attributes[localVarName]; ok {
			// Because there might be function call like merge() and concat() in
			// local variables, or even using another local variable, so it will
			// requires to perform a deep traverse into the nested local variable.
			tagKeys, err := r.traverseSearchExpr(runner, localVarAttr.Expr)
			if err != nil {
				return nil, err
			}
			localTagKeys = slices.Concat(localTagKeys, tagKeys)
		}
	}
	return localTagKeys, nil
}

func (r *TerraformRequiredTags) getTagsKey(val cty.Value) []string {
	if val.IsKnown() && !val.IsNull() && val.CanIterateElements() {
		var localTagKeys []string
		if val.Type().IsObjectType() {
			// If tags is object value
			for it := val.ElementIterator(); it.Next(); {
				k, _ := it.Element()
				localTagKeys = append(localTagKeys, k.AsString())
			}
		} else if val.Type().IsTupleType() {
			// If tags is list value, used in Openstack provider like compute_instance_v2.
			for it := val.ElementIterator(); it.Next(); {
				_, v := it.Element()
				localTagKeys = append(localTagKeys, r.splitTagKeyString(v))
			}
		}
		return localTagKeys
	}
	return []string{}
}

// Handle error when trying to evaluate tuple values by manually Manually parsing
// the list when the error is unknown variable or null value.
// This might happen because TFLint do not know the actual value when using locals,
// variable or output from other resources in the string.
func (r *TerraformRequiredTags) handleEvaluateTupleError(err error, expr *hclsyntax.TupleConsExpr) ([]string, bool) {
	var tagsKey []string
	if strings.Contains(err.Error(), "Unknown variable") || strings.Contains(err.Error(), "Attempt to get attribute from null value") || strings.Contains(err.Error(), "This object does not have an attribute named") {
		for _, e := range expr.Exprs {
			if tmplExpr, ok := e.(*hclsyntax.TemplateExpr); ok {
				for _, part := range tmplExpr.Parts {
					if partExpr, ok := part.(*hclsyntax.LiteralValueExpr); ok {
						tagsKey = append(tagsKey, r.splitTagKeyString(partExpr.Val))
					}
				}
			}
		}
		return tagsKey, true
	} else {
		return nil, false
	}
}

// Split a single tag key in string with delimiter ':'.
func (r *TerraformRequiredTags) splitTagKeyString(val cty.Value) string {
	// If the value is unknown, AsString() will throw panic errors.
	if !val.IsWhollyKnown() {
		return strings.Split(val.Range().StringPrefix(), ":")[0]
	} else {
		return strings.Split(val.AsString(), ":")[0]
	}
}
