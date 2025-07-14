package project

import "fmt"

func ReferenceLink(name string) string {
	return fmt.Sprintf("https://github.com/myklst/tflint-ruleset-myklst/blob/main/docs/rules/%s.md", name)
}
