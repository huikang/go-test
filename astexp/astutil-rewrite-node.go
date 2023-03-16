// Apply more intricate AST rewriting to Go code, using the astutil package.
//
// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.
package astexp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func astutilRewriteNode(filename string, funcName string) error {
	fset := token.NewFileSet()
	src, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("error read file", err)
	}
	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		log.Fatal("error parse file", err)
	}

	outFilename, err := generatedFilename(filename)
	if err != nil {
		return err
	}
	fmt.Println("Generated filename", outFilename)
	// return nil

	// add imports
	astutil.AddImport(fset, file, "github.com/hashicorp/consul/api")
	astutil.AddNamedImport(fset, file, "libcluster", "github.com/hashicorp/consul/test/integration/consul-container/libs/cluster")
	astutil.AddImport(fset, file, "github.com/hashicorp/consul/test/integration/consul-container/libs/topology")
	astutil.AddImport(fset, file, "github.com/hashicorp/consul/test/integration/consul-container/libs/utils")

	astutil.DeleteImport(fset, file, "github.com/hashicorp/consul/sdk/testutil")

	foundFuncCallers := map[string]*ast.FuncDecl{}
	var currFunction *ast.FuncDecl
	astutil.Apply(file,
		func(c *astutil.Cursor) bool {
			n := c.Node()
			switch x := n.(type) {
			case *ast.CallExpr:
				// fmt.Println("DEBUG hui, node is CallExpr")
				id, ok := x.Fun.(*ast.Ident)
				if ok {
					if id.Name == funcName {
						fmt.Printf("Inspect found call to %s() at %s\n", funcName, fset.Position(n.Pos()))
						foundFuncCallers[currFunction.Name.Name] = currFunction
					}
				}
			case *ast.FuncDecl:
				fmt.Println("DEBUG hui, node is FuncDecl, func name", x.Name.Name)
				currFunction = x
			}
			return true
		},
		func(c *astutil.Cursor) bool {
			n := c.Node()
			switch x := n.(type) {
			case *ast.AssignStmt:
				// rhs := x.Rhs
				callExp, ok := x.Rhs[0].(*ast.CallExpr)
				if !ok {
					return true
				}

				id, ok := callExp.Fun.(*ast.Ident)
				if ok {
					if id.Name == funcName {
						upgradeNode := clusterUpgradeNode2()
						c.Replace(upgradeNode)
					}
				}
				fmt.Println("DEUBG hui, rewrite")
			case *ast.CallExpr:
				// id, ok := x.Fun.(*ast.Ident)
				// if ok {
				// 	if id.Name == funcName {
				// 		upgradeNode := clusterUpgradeNode()
				// 		// c.Replace(&ast.UnaryExpr{
				// 		// 	Op: token.NOT,
				// 		// 	X:  x,
				// 		// })
				// 		c.Replace(upgradeNode)
				// 	}
				// }
			}

			return true
		})

	for n, v := range foundFuncCallers {
		fmt.Println("Found callers", n, v.Name)
	}
	fmt.Println("Modified AST:")
	// printer.Fprint(os.Stdout, fset, file)

	outf, err := os.Create(outFilename)
	if err != nil {
		return err
	}
	err = printer.Fprint(outf, fset, file)
	if err != nil {
		return err
	}
	return nil
}

func clusterUpgradeNode2() ast.Node {
	node := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "_",
			},
			&ast.Ident{
				Name: "_",
			},
			&ast.Ident{
				Name: "c",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "topology",
					},
					Sel: &ast.Ident{
						Name: "NewCluster",
					},
				},
				Lparen: 57,
				Args: []ast.Expr{
					&ast.Ident{
						Name: "t",
					},
					&ast.UnaryExpr{
						OpPos: 61,
						Op:    token.AND,
						X: &ast.CompositeLit{
							Type: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "topology",
								},
								Sel: &ast.Ident{
									Name: "ClusterConfig",
								},
							},
							Elts: []ast.Expr{
								&ast.KeyValueExpr{
									Key: &ast.Ident{
										Name: "NumServers",
									},
									Colon: 96,
									Value: &ast.BasicLit{
										ValuePos: 98,
										Kind:     token.INT,
										Value:    "1",
									},
								},
								&ast.KeyValueExpr{
									Key: &ast.Ident{
										Name: "NumClients",
									},
									Colon: 111,
									Value: &ast.BasicLit{
										ValuePos: 113,
										Kind:     token.INT,
										Value:    "1",
									},
								},
								&ast.KeyValueExpr{
									Key: &ast.Ident{
										Name: "BuildOpts",
									},
									Colon: 125,
									Value: &ast.UnaryExpr{
										OpPos: 127,
										Op:    token.AND,
										X: &ast.CompositeLit{
											Type: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "libcluster",
												},
												Sel: &ast.Ident{
													Name: "BuildOptions",
												},
											},
											Elts: []ast.Expr{
												&ast.KeyValueExpr{
													Key: &ast.Ident{
														Name: "ConsulVersion",
													},
													Colon: 167,
													Value: &ast.Ident{
														Name: "utils.LatestVersion",
													},
												},
												&ast.KeyValueExpr{
													Key: &ast.Ident{
														Name: "Datacenter",
													},
													Colon: 199,
													Value: &ast.BasicLit{
														ValuePos: 211,
														Kind:     token.STRING,
														Value:    "\"dc1\"",
													},
												},
												&ast.KeyValueExpr{
													Key: &ast.Ident{
														Name: "InjectAutoEncryption",
													},
													Colon: 239,
													Value: &ast.Ident{
														Name: "true",
													},
												},
											},
											Incomplete: false,
										},
									},
								},
								&ast.KeyValueExpr{
									Key: &ast.Ident{
										Name: "ApplyDefaultProxySettings",
									},
									Colon: 275,
									Value: &ast.Ident{
										Name: "true",
									},
								},
							},
							Incomplete: false,
						},
					},
				},
				Ellipsis: 0,
			},
		},
	}

	return node
}

func generatedFilename(filename string) (string, error) {
	dir := filepath.Dir(filename)
	baseFilename := filepath.Base(filename)

	if !strings.Contains(baseFilename, "_test.go") {
		return "", fmt.Errorf("input isn't go test file: %s", baseFilename)
	}

	fmt.Printf("Dir %s, base file %s\n", dir, baseFilename)

	// ext := filepath.Ext(baseFilename)
	// fmt.Println("ext", ext)

	outFilename := strings.Replace(baseFilename, "_test.go", "_generated_test.go", 1)

	outFilename = filepath.Join(dir, outFilename)
	return outFilename, nil
}
