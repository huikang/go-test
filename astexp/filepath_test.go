package astexp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilePath(t *testing.T) {
	dir := "/Users/huikang/go/src/github.com/hashicorp/consul-enterprise"
	walkDir(dir)
}

func TestMemDBTestCases(t *testing.T) {
	db, err := newTestCasesDB()
	require.NoError(t, err)
	t.Log("created db")

	// Insert some data
	consulTestCases := []*UnitTestCase{
		{
			Name:         "Test1",
			FileName:     "File1",
			NameFileName: "Test1_File1",
			Kind:         UnitTestKindTestServer,
			CallExprs:    []string{"NewTestServerConfigT"},
		},
		{
			Name:         "Test2",
			FileName:     "File1",
			NameFileName: "Test2_File1",
			Kind:         UnitTestKindTestServer,
			CallExprs:    []string{"NewTestServerConfigT", "makeClient"},
		},
		{
			Name:         "Test3",
			FileName:     "File3",
			NameFileName: "Test3_File3",
			Kind:         UnitTestKindTestServer,
			CallExprs:    []string{"not_calling"},
		},
		{
			Name:         "Test4",
			FileName:     "File3",
			NameFileName: "Test4_File3",
			Kind:         UnitTestKindTestServer,
			CallExprs:    []string{"makeClient"},
		},
	}

	// Create a write transaction
	txn := db.Txn(true)
	for _, p := range consulTestCases {
		if err := txn.Insert(unitTestTableName, p); err != nil {
			panic(err)
		}
	}

	// Commit the transaction
	txn.Commit()

	// Create read-only transaction
	txn = db.Txn(false)
	defer txn.Abort()

	// Lookup by call_exprs
	it, err := txn.Get(unitTestTableName, "call_exprs", "makeClient")
	if err != nil {
		panic(err)
	}
	count := 0
	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*UnitTestCase)
		t.Logf("  %s\n", p.Name)
		count++
	}
	require.Equal(t, 2, count)
}
