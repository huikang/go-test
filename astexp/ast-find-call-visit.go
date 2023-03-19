// Example copied from https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/

package astexp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
)

func searchFunctionCallerInFile(fn string, filename string) {
	fset := token.NewFileSet()
	src, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("error read file", err)
	}

	fmt.Println("start parsing", filename)
	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		log.Fatal("error parsing file", err)
	}
	fmt.Println("finished parsing")

	// visitor := &Visitor{fset: fset}
	// ast.Walk(visitor, file)

	callers := map[string]*ast.FuncDecl{}
	// funcStart := 0
	funcEnd := -1
	currFuncName := ""
	var currFunction *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		var currPos int
		if n != nil {
			currPos = int(n.Pos())
		}
		if n != nil && currPos >= funcEnd && funcEnd > 0 {
			fmt.Println("DEBUG hui, function end", currFuncName)
			currFunction = nil
		}
		switch x := n.(type) {
		case *ast.CallExpr:
			id, ok := x.Fun.(*ast.Ident)
			if ok {
				if id.Name == fn {
					fmt.Printf("Inspect found call to %s() at %s\n", fn, fset.Position(n.Pos()))
					callers[currFuncName] = currFunction
				}
			}
		case *ast.FuncDecl:
			// fmt.Println("	Lbrace", x.Body.Lbrace)
			// fmt.Println("	Rbrace", x.Body.Rbrace)
			// fmt.Println("	len ", len(x.Body.List))
			currFuncName = x.Name.Name
			currFunction = x
			funcEnd = int(x.Body.Rbrace)
		}
		return true
	})

	for n, v := range callers {
		fmt.Println("Found callers", n, v.Name)
	}

	printer.Fprint(os.Stdout, fset, callers["TestAPI_HealthService_SingleTag"])
}

type Visitor struct {
	fset *token.FileSet
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	// fmt.Println("Visit node")
	if n == nil {
		return nil
	}

	switch x := n.(type) {
	case *ast.CallExpr:
		// fmt.Println("Visit type", x)
		id, ok := x.Fun.(*ast.Ident)
		if ok {
			if id.Name == "pred" {
				fmt.Printf("Visit found call to pred() at %s\n", v.fset.Position(n.Pos()))
			}
		}
	}
	return v
}
