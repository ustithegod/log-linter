package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewWithConfig(DefaultConfig()), "cases")
}

func TestAnalyzerSuggestedFixes(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.RunWithSuggestedFixes(t, testdata, NewWithConfig(DefaultConfig()), "fixes")
}
