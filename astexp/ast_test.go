package astexp

import (
	"fmt"
	"os"
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
		// {
		// 	filename:     "/Users/huikang/go/src/github.com/hashicorp/consul-enterprise/agent/consul/state/acl_test.go",
		// 	functionName: "makeClientWithConfig",
		// },
	}

	for _, tc := range testCases {
		// searchFunctionCallerInFile(tc.functionName, tc.filename)
		astutilFindCallNode(tc.filename, tc.functionName)
	}
}

func TestAstFindCallerInDir(t *testing.T) {
	testCases := []struct {
		dir          string
		functionName string
	}{
		// {
		// 	filename:     path.Join("testdata", "src.go"),
		// 	functionName: "pred",
		// },
		{
			dir:          "/Users/huikang/go/src/github.com/hashicorp/consul-enterprise",
			functionName: "NewTestServerConfigT",
		},
	}
	for _, tc := range testCases {
		// searchFunctionCallerInFile(tc.functionName, tc.filename)
		path := tc.dir
		fileInfo, err := os.Stat(path)
		if err != nil {
			t.Log("can't open dir", tc.dir)
		}

		if !fileInfo.IsDir() {
			t.Logf("not a dir: %s", tc.dir)
		}
		// astutilFindCallNode(tc., tc.functionName)
		testFiles := listTestFiles(tc.dir)
		t.Logf("found total test files: %d", len(testFiles))
		for _, fileName := range testFiles {
			t.Logf("search in file %s", fileName)
			err = astutilFindCallNode(fileName, tc.functionName)
			if err != nil {
				t.Logf("error find caller %s: %s", fileName, err)
			}
		}
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
