package astexp

import (
	"fmt"
	"path"
	"testing"
)

func TestAstFindCaller(t *testing.T) {
	testCases := []struct {
		filename     string
		functionName string
	}{
		// {
		// 	filename:     path.Join("testdata", "src.go"),
		// 	functionName: "pred",
		// },
		{
			filename:     path.Join("/Users/huikang/go/src/github.com/hashicorp/consul-enterprise/api", "health_test.go"),
			functionName: "makeClientWithConfig",
		},
	}

	for _, tc := range testCases {
		searchFunctionCallerInFile(tc.functionName, tc.filename)
	}
}

func TestParseTestfiles(t *testing.T) {
	dir := "/Users/huikang/go/src/github.com/hashicorp/consul-enterprise"
	ret := listTestFiles(dir)
	fmt.Println("All test files:", len(ret))

	functionName := "makeClientWithConfig"
	for _, filename := range ret {
		searchFunctionCallerInFile(functionName, filename)
	}
}

func TestAstInspectRewrite(t *testing.T) {
	astInspectRewrite()
}

func TestAstUtilInspectRewrite(t *testing.T) {
	testCases := []struct {
		filename     string
		functionName string
	}{
		// {
		// 	filename:     path.Join("testdata", "src.go"),
		// 	functionName: "pred",
		// },
		// {
		// 	filename:     path.Join("/Users/huikang/go/src/github.com/hashicorp/consul-enterprise/api", "health_test.go"),
		// 	functionName: "makeClientWithConfig",
		// },
		{
			filename:     path.Join("testdata", "health_test.go"),
			functionName: "makeClientWithConfig",
		},
	}

	for _, tc := range testCases {
		err := astutilRewriteNode(tc.filename, tc.functionName)
		if err != nil {
			t.Logf("error rewrite %s: %s", tc.filename, err)
		}
	}
}

func TestLookup(t *testing.T) {

}
