package main

import (
	"go/ast"
	"go/token"

	"github.com/alpardfm/cost-efficient-go/types"
)

// FileDetector wraps a library Detector to operate at file level.
// Each instance wraps exactly one types.Detector.
// Implementations are stateless and safe for concurrent use.
type FileDetector interface {
	// DetectFile walks all AST nodes in the file and delegates detection
	// to the underlying library Detector. Returns aggregated findings.
	DetectFile(fset *token.FileSet, file *ast.File) []types.Finding

	// Rule returns the detection rule metadata from the underlying Detector.
	Rule() types.Rule
}

// fileDetectorAdapter wraps a single types.Detector into a FileDetector.
type fileDetectorAdapter struct {
	detector types.Detector
}

// NewFileDetector creates a FileDetector adapter for the given library Detector.
func NewFileDetector(d types.Detector) FileDetector {
	return &fileDetectorAdapter{detector: d}
}

// DetectFile walks all AST nodes in the file and delegates detection
// to the underlying library Detector. Returns aggregated findings.
func (a *fileDetectorAdapter) DetectFile(fset *token.FileSet, file *ast.File) []types.Finding {
	if file == nil {
		return nil
	}

	filePath := fset.Position(file.Pos()).Filename

	var findings []types.Finding
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		pos := fset.Position(n.Pos())
		ctx := types.ASTContext{
			FilePath:    filePath,
			Line:        pos.Line,
			Node:        n,
			CodeContext: "",
		}

		results := a.detector.Detect(ctx)
		findings = append(findings, results...)

		return true
	})

	return findings
}

// Rule returns the detection rule metadata from the underlying Detector.
func (a *fileDetectorAdapter) Rule() types.Rule {
	return a.detector.Rule()
}
