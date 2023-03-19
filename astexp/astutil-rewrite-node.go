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

var (
	consulApiStructs = map[string]int{
		"AgentServiceRegistration": 0,
		"HealthChecks":             0,
		"AgentServiceCheck":        0,
		"HealthCheck":              0,
	}
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
	var clusterName string
	astutil.Apply(file,
		func(c *astutil.Cursor) bool {
			n := c.Node()
			switch x := n.(type) {
			case *ast.CallExpr:
				id, ok := x.Fun.(*ast.Ident)
				if ok {
					if id.Name == funcName {
						fmt.Printf("Inspect found call to %s() at %s\n", funcName, fset.Position(n.Pos()))
						foundFuncCallers[currFunction.Name.Name] = currFunction
					}
				}
			case *ast.FuncDecl:
				currFunction = x
			}
			return true
		},
		func(c *astutil.Cursor) bool {
			n := c.Node()
			switch x := n.(type) {
			case *ast.AssignStmt:

				name, rewritten := rewriteClusterCreation(c, x, funcName)
				if rewritten {
					clusterName = name
					fmt.Println("Cluster name is", clusterName)
					return true
				}

				// rewritten = rewriteConsulApi(c, x)
				// if rewritten {
				// 	return true
				// }

			case *ast.CompositeLit:
				rewritten := rewriteConsulApi(c, x)
				if rewritten {
					return true
				}
			case *ast.DeferStmt:
				if clusterName == "" {
					return true
				}
				selector, ok := x.Call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				id, ok := selector.X.(*ast.Ident)
				if !ok {
					return true
				}

				if id.Name == clusterName {
					c.Delete()
				}

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

func rewriteConsulApi(c *astutil.Cursor, compositeLit *ast.CompositeLit) bool {
	// compositeLit, ok := x.Rhs[0].(*ast.CompositeLit)
	// if !ok {
	// 	return false
	// }

	id, ok := compositeLit.Type.(*ast.Ident)
	if !ok {
		return false
	}

	if _, ok := consulApiStructs[id.Name]; !ok {
		return false
	}

	id.Name = "api." + id.Name
	return true
}

func rewriteClusterCreation(c *astutil.Cursor, x *ast.AssignStmt, funcName string) (string, bool) {
	callExp, ok := x.Rhs[0].(*ast.CallExpr)
	if !ok {
		return "", false
	}

	id, ok := callExp.Fun.(*ast.Ident)
	if !ok {
		return "", false
	}

	if id.Name != funcName {
		return "", false
	}

	upgradeNode := clusterUpgradeNode2()
	c.Replace(upgradeNode)

	serverIdent, ok := x.Lhs[1].(*ast.Ident)
	if !ok {
		return "", false
	}
	return serverIdent.Name, true
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

	outFilename := strings.Replace(baseFilename, "_test.go", "_generated_test.go", 1)

	outFilename = filepath.Join(dir, outFilename)
	return outFilename, nil
}
