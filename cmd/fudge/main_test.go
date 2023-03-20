package main

import (
	"bytes"
	"go/token"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

func Test_rewriteSentinelErrors(t *testing.T) {
	r := randReader
	t.Cleanup(func() { randReader = r })
	randReader = rand.New(rand.NewSource(1))

	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "testdata")

	entries, err := os.ReadDir(dir)
	require.Nil(t, err)

	fset := token.NewFileSet()

	g := goldie.New(t)

	for _, entry := range entries {
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".go") {
			continue
		}

		name := strings.TrimSuffix(filename, ".go")

		t.Run(name, func(t *testing.T) {
			in := filepath.Join(dir, filename)

			f, err := decorator.ParseFile(fset, in, nil, 0)
			require.Nil(t, err)

			rewriteSentinelErrorsInFile(f)

			bs := bytes.NewBuffer(nil)
			err = decorator.Fprint(bs, f)
			require.Nil(t, err)

			g.Assert(t, name, bs.Bytes())
		})
	}
}
