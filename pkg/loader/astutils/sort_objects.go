package astutils

import (
	"go/ast"
	"sort"
)

type objectsByPos []*ast.Object

func (p objectsByPos) Len() int {
	return len(p)
}

func (p objectsByPos) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p objectsByPos) Less(i, j int) bool {
	return p[i].Pos() < p[j].Pos()
}

// SortObjectsByPos takes a map of ast objects and sorts it by Pos to a slice.
// All objects must be from the same file.
func SortObjectsByPos(objectsMap map[string]*ast.Object) []*ast.Object {
	objects := make(objectsByPos, 0, len(objectsMap))
	for _, v := range objectsMap {
		objects = append(objects, v)
	}
	sort.Sort(objects)
	return objects
}
