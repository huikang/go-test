// Apply more intricate AST rewriting to Go code, using the astutil package.
//
// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.
package astexp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

	"golang.org/x/tools/go/ast/astutil"
)

func astutilFindCallNode(filename string, funcName string) error {
	fset := token.NewFileSet()
	src, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("error read file: ", err)
	}
	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		log.Fatal("error parse file", err)
	}

	foundFuncCallers := map[string]*ast.FuncDecl{}
	var currFunction *ast.FuncDecl
	astutil.Apply(file,
		func(c *astutil.Cursor) bool {
			n := c.Node()
			switch x := n.(type) {
			case *ast.CallExpr:
				id, ok := x.Fun.(*ast.Ident)
				if ok {
					if id.Name == funcName {
						fmt.Printf("Inspect found call to %s() at %s\n", funcName, fset.Position(n.Pos()))
						if currFunction != nil {
							foundFuncCallers[currFunction.Name.Name] = currFunction
						} else {
							panic("found a caller, but currFunction is nil")
						}
					}
				}
			case *ast.FuncDecl:
				currFunction = x
			}
			return true
		},
		nil,
	)

	for n, v := range foundFuncCallers {
		fmt.Println("Found caller", n, v.Name)
	}

	return nil
}
