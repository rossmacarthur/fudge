package main

import (
	"crypto/rand"
	"fmt"
	"go/token"
	"math"
	"math/big"
	"os"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/rossmacarthur/fudge"
	"github.com/rossmacarthur/fudge/errors"
)

const imp = "github.com/rossmacarthur/fudge/errors"

func main() {
	fset := token.NewFileSet()

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		err = rewriteSentinelErrors(fset, path)
		if err != nil {
			return errors.Wrap(err, "", fudge.KV("path", path))
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %#v\n", err)
		os.Exit(1)
	}
}

func rewriteSentinelErrors(fset *token.FileSet, path string) error {
	// Parse the file
	f, err := decorator.ParseFile(fset, path, nil, 0)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// Find and rewrite the sentinel errors
	rewriteSentinelErrorsInFile(f)

	// Write the file back
	of, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "")
	}
	defer of.Close()

	err = decorator.Fprint(of, f)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return err
}

func rewriteSentinelErrorsInFile(f *dst.File) {
	var alias string

	dst.Inspect(f, func(n dst.Node) bool {
		if spec, ok := n.(*dst.ImportSpec); ok {
			if spec.Path.Value == `"`+imp+`"` {
				if spec.Name != nil {
					alias = spec.Name.Name
				} else {
					alias = filepath.Base(imp)
				}
			}
		}

		if spec, ok := n.(*dst.ValueSpec); ok {
			// Check if the value is a call to errors.Sentinel
			if len(spec.Values) != 1 {
				return false
			}
			call, ok := spec.Values[0].(*dst.CallExpr)
			if !ok {
				return false
			}
			sel, ok := call.Fun.(*dst.SelectorExpr)
			if !ok {
				return false
			}
			if sel.Sel.Name != "Sentinel" {
				return false
			}
			if sel.X.(*dst.Ident).Name != alias {
				return false
			}

			if len(call.Args) == 2 {
				if lit, ok := call.Args[1].(*dst.BasicLit); ok && lit.Kind == token.STRING {
					if checkCode(lit.Value[1 : len(lit.Value)-1]) {
						return false
					}
				}
			}

			var args []dst.Expr
			if len(call.Args) > 0 {
				args = append(args, call.Args[0])
			} else {
				args = append(args, &dst.BasicLit{
					Kind:  token.STRING,
					Value: `""`,
				})
			}
			call.Args = append(args, &dst.BasicLit{
				Kind:  token.STRING,
				Value: `"` + randomCode() + `"`,
			})

		}

		return true
	})
}

var randReader = rand.Reader

func randomCode() string {
	n, err := rand.Int(randReader, big.NewInt(int64(math.MaxInt64)))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("ERR_%016x", n.Int64())
}

func checkCode(code string) bool {
	if len(code) != 20 {
		return false
	}
	if code[:4] != "ERR_" {
		return false
	}
	for _, c := range code[4:] {
		if c < '0' || c > 'f' {
			return false
		}
	}
	return true
}
