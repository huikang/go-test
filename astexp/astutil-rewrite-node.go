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
		"QueryOptions":             0,
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
	astutil.AddImport(fset, file, "context")
	astutil.AddImport(fset, file, "github.com/hashicorp/consul/api")
	astutil.AddNamedImport(fset, file, "libcluster", "github.com/hashicorp/consul/test/integration/consul-container/libs/cluster")
	astutil.AddImport(fset, file, "github.com/hashicorp/consul/test/integration/consul-container/libs/topology")
	astutil.AddImport(fset, file, "github.com/hashicorp/consul/test/integration/consul-container/libs/utils")

	astutil.DeleteImport(fset, file, "github.com/hashicorp/consul/sdk/testutil")

	foundFuncCallers := map[string]*ast.FuncDecl{}
	var currFunction *ast.FuncDecl
	var clusterName string
	rewriteNodeName := false
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

				useNodeName := hasNodeName(c, x, "nodename")
				if useNodeName {
					rewriteNodeName = true
					c.Delete()
					fmt.Println("rewriteNodeName set to true", rewriteNodeName)
					return true
				}

				rewritten = getAgentStatement(c, x, rewriteNodeName)
				if rewritten {
					return true
				}

			case *ast.CompositeLit:
				rewritten := rewriteConsulApi(c, x)
				if rewritten {
					return true
				}
			case *ast.ExprStmt:
				rewritten := removeClusterSelectorExpr(c, x, clusterName)
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

func getAgentStatement(c *astutil.Cursor, stmt *ast.AssignStmt, rewriteNodeName bool) bool {
	callExpr, ok := stmt.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}

	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if selectorExpr.Sel.Name != "Agent" {
		return false
	}

	if rewriteNodeName {
		getNodeName := &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "nodename",
				},
				&ast.Ident{
					Name: "_",
				},
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "agent",
						},
						Sel: &ast.Ident{
							Name: "NodeName",
						},
					},
					Lparen:   56,
					Ellipsis: 0,
				},
			},
		}

		c.InsertAfter(getNodeName)
	}

	return true
}

func hasNodeName(c *astutil.Cursor, stmt *ast.AssignStmt, nodename string) bool {
	id, ok := stmt.Lhs[0].(*ast.Ident)
	if !ok {
		return false
	}

	if id.Name != nodename {
		return false
	}

	return true
}

func removeClusterSelectorExpr(c *astutil.Cursor, expr *ast.ExprStmt, clusterName string) bool {
	if clusterName == "" {
		return false
	}
	callExpr, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	id, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	if id.Name != clusterName {
		return false
	}

	c.Delete()
	return true
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

	createClusterNode, upgradeClusterNode := clusterUpgradeNode2()
	c.Replace(createClusterNode)
	c.InsertAfter(upgradeClusterNode)

	serverIdent, ok := x.Lhs[1].(*ast.Ident)
	if !ok {
		return "", false
	}
	return serverIdent.Name, true
}

func clusterUpgradeNode2() (ast.Node, ast.Node) {
	creatClusterNode := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "cluster",
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

	upgradeFuncNode := &ast.DeferStmt{
		Defer: 27,
		Call: &ast.CallExpr{
			Fun: &ast.FuncLit{
				Type: &ast.FuncType{
					Func:   33,
					Params: &ast.FieldList{},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "require",
									},
									Sel: &ast.Ident{
										Name: "NoError",
									},
								},
								Lparen: 57,
								Args: []ast.Expr{
									&ast.Ident{
										Name: "t",
									},
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "cluster",
											},
											Sel: &ast.Ident{
												Name: "StandardUpgrade",
											},
										},
										Lparen: 93,
										Args: []ast.Expr{
											&ast.Ident{
												Name: "t",
											},
											&ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X: &ast.Ident{
														Name: "context",
													},
													Sel: &ast.Ident{
														Name: "Background",
													},
												},
												Lparen:   115,
												Ellipsis: 0,
											},
											&ast.SelectorExpr{
												X: &ast.Ident{
													Name: "utils",
												},
												Sel: &ast.Ident{
													Name: "LatestVersion",
												},
											},
										},
										Ellipsis: 0,
									},
								},
								Ellipsis: 0,
							},
						},
					},
				},
			},
			Lparen:   141,
			Ellipsis: 0,
		},
	}

	return creatClusterNode, upgradeFuncNode
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
