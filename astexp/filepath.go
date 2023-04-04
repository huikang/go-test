package astexp

import (
	"fmt"
	"go/ast"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-memdb"
)

func walkDir(dir string) {
	count := 0
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fmt.Println(path, info.Size())
			if strings.Contains(info.Name(), "_test.go") {
				count++
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	log.Println("total number of file", count)
}

func listTestFiles(dir string) []string {
	results := []string{}
	count := 0
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fmt.Println(path, info.Size())
			if strings.Contains(info.Name(), "_test.go") {
				count++
				results = append(results, path)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	log.Println("total number of file", count)
	return results
}

// UnitTestKind is the type of unit test; currently the supported
// type is “TestServer”, “TestAgent”
type UnitTestKind int

const (
	UnitTestKindTestServer = iota
	UnitTestKindTestAgent
)

func (u UnitTestKind) String() string {
	return [...]string{"TestServer", "TestAgent"}[u]
}

// UnitTestCase represents a Test case, e.g., Test_HealthAPI
type UnitTestCase struct {
	Name     string
	FileName string
	// NameFileName is unique
	NameFileName string
	Kind         UnitTestKind
	RootAstNode  ast.Node
	// CallExprs is the list of functions called by the test function
	CallExprs []string
}

const unitTestTableName = "test_cases"

func newTestCasesDB() (*memdb.MemDB, error) {
	// Create an in-mem DB schema for all test functions
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			unitTestTableName: &memdb.TableSchema{
				Name: unitTestTableName,
				Indexes: map[string]*memdb.IndexSchema{
					// "name": {
					// 	Unique:  false,
					// 	Indexer: &memdb.StringFieldIndex{Field: "Name"},
					// },
					// "filename": {
					// 	Unique: false,
					// },
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "NameFileName"},
					},
					"test_kind": {
						Name:    "test_kind",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Kind"},
					},
					"call_exprs": {
						Name:   "call_exprs",
						Unique: false,
						Indexer: &memdb.StringSliceFieldIndex{
							Field: "CallExprs",
						},
					},
				},
			},
		},
	}
	err := schema.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating schema: %v", err)
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, fmt.Errorf("error creating memdb: %v", err)
	}
	return db, nil
}
