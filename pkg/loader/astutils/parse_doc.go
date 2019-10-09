package astutils

import (
	"go/ast"
	"go/doc"
	"go/token"
	"strconv"
)

func PackageDoc(pkgPath string, files []*ast.File, fset *token.FileSet) *doc.Package {
	fileMap := make(map[string]*ast.File, len(files))
	for i, file := range files {
		astFile := file
		fileMap[strconv.Itoa(i)] = astFile
	}
	astPkg, _ := ast.NewPackage(fset, fileMap, nil, nil)
	return doc.New(astPkg, pkgPath, 0)
}

func TypeDoc(pkgDoc *doc.Package, typeName string) string {
	for _, t := range pkgDoc.Types {
		if t.Name == typeName {
			return t.Doc
		}
	}
	return ""
}
