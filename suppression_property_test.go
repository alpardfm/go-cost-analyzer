package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestProperty_SuppressionCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	validPatterns := map[string]bool{"CEG-016": true, "CEG-017": true, "CEG-018": true}
	patternNames := []string{"CEG-016", "CEG-017", "CEG-018"}

	properties.Property("suppressed lines are never reported, unsuppressed lines are always reported", prop.ForAll(
		func(patternIdx int, suppressLine bool) bool {
			pattern := patternNames[patternIdx%len(patternNames)]

			var src string
			if suppressLine {
				src = fmt.Sprintf("package main\n\n//noinspect:%s\nfunc main() {}\n", pattern)
			} else {
				src = "package main\n\nfunc main() {}\n"
			}

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
			if err != nil {
				return true
			}

			result := ParseSuppressions(file, fset, validPatterns)

			// Line 4 is where "func main() {}" is
			if suppressLine {
				// Comment on line 3, code on line 4 → line 4 should be suppressed
				return result.IsSuppressed(4, pattern)
			}
			// No suppression → line 4 should NOT be suppressed
			return !result.IsSuppressed(4, pattern)
		},
		gen.IntRange(0, 2),
		gen.Bool(),
	))

	properties.TestingRun(t)
}
