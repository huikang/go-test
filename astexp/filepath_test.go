package astexp

import (
	"testing"
)

func TestFilePath(t *testing.T) {
	dir := "/Users/huikang/go/src/github.com/hashicorp/consul-enterprise"
	walkDir(dir)
}
