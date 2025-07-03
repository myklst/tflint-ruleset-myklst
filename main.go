package main

import (
	"github.com/myklst/tflint-ruleset-myklst/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "myklst",
			Version: "0.0.1",
			Rules: []tflint.Rule{
				rules.NewTerraformMetaArguments(),
				rules.NewTerraformAnyTypeVariables(),
				rules.NewTerraformRequiredTags(),
				rules.NewTerraformModuleSourceVersion(),
				rules.NewTerraformVarsObjectKeysNamingConventions(),
				rules.NewTerraformRequiredVariables(),
			},
		},
	})
}
