package loader

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/loader"

	"github.com/jimmidyson/prettyconf/pkg/loader/astutils"
)

type ASTLoader struct {
	requestedPackages []string
	logger            logr.Logger
	prog              *loader.Program
}

func New(packages []string, logger logr.Logger) *ASTLoader {
	return &ASTLoader{requestedPackages: packages, logger: logger}
}

type Package struct {
	Path  string
	Types []Type
	Doc   string
}

type Type struct {
	Name           string
	Package        string
	Fields         []Field
	Doc            string
}

type Field struct {
	Name         string
	Doc          string
	Anonymous    bool
	JSONRequired bool
	JSONProperty string
	Type         types.Type
	TypeName     string
}

func (l *ASTLoader) Load() ([]Package, error) {
	var conf loader.Config
	conf.ParserMode = parser.ParseComments

	for _, pkg := range l.requestedPackages {
		conf.Import(pkg)
	}

	prog, err := conf.Load()
	if err != nil {
		return nil, errors.Wrap(err, "cannot load requested packages")
	}
	l.prog = prog

	loadedPackages := make([]Package, 0, len(l.requestedPackages))
	for _, pkg := range prog.InitialPackages() {
		pkgPath := pkg.Pkg.Path()

		l.logger.V(5).Info("parsing package", "package", pkgPath)

		l.logger.V(5).Info("extracting package docs", "package", pkgPath)
		pkgDoc := astutils.PackageDoc(pkgPath, pkg.Files, prog.Fset)

		exportedTypes := []Type{}
		for _, file := range pkg.Files {
			filePos := prog.Fset.Position(file.Pos())
			l.logger.V(5).Info("parsing file", "package", pkgPath, "file", filePos.Filename)

			l.logger.V(5).Info("sorting objects", "package", pkgPath, "file", filePos.Filename)
			sortedObjects := astutils.SortObjectsByPos(file.Scope.Objects)

			for _, currentObj := range sortedObjects {
				t, ok := currentObj.Decl.(*ast.TypeSpec)
				if !ok || !t.Name.IsExported() {
					continue
				}
				astStructType, ok := t.Type.(*ast.StructType)
				if !ok {
					continue
				}

				typ, ok := pkg.Types[t.Type]
				if !ok {
					return nil, errors.Errorf("unable to load struct type: %s", t.Name.Name)
				}
				structType, ok := typ.Type.(*types.Struct)
				if !ok {
					continue
				}
				l.logger.V(5).Info("loaded struct type", "name", t.Name.Name)

				structFields := make([]Field, 0, structType.NumFields())

				for j := 0; j < structType.NumFields(); j++ {
					fld := structType.Field(j)
					if !fld.IsField() || !fld.Exported() {
						continue
					}

					jsonProperty := fld.Name()
					required := true
					fldTag := structType.Tag(j)
					tags, err := ParseStructTags(fldTag)
					if err != nil {
						return nil, errors.Wrapf(err, "failed to parse struct tag `%s`", fldTag)
					}

					for _, t := range tags {
						if t.Name == "json" {
							split := strings.Split(t.Value, ",")
							jsonProperty = split[0]
							for _, tagValue := range split[1:] {
								if tagValue == "omitempty" {
									required = false
									break
								}
							}
							break
						}
					}

					if jsonProperty == "-" {
						l.logger.V(5).Info("ignoring struct field as not serialized", "struct", t.Name.Name, "field", fld.Name())
						continue
					}

					l.logger.V(5).Info("adding struct field", "struct", t.Name.Name, "field", fld.Name(), "type", fld.Type().String())
					fldDoc := ""
					if j < astStructType.Fields.NumFields() {
						fldDoc = strings.TrimSpace(astStructType.Fields.List[j].Doc.Text())
						docLines := strings.Split(fldDoc, "\n")
						for i := len(docLines) - 1; i >= 0; i-- {
							if strings.HasPrefix(strings.TrimSpace(docLines[i]), "+optional") {
								required = false
							}
						}
					}

					typeName := fld.Type().String()
					if idx := strings.Index(typeName, "vendor/"); idx > -1 {
						typeName = typeName[idx+len("vendor/"):]
					}
					f := Field{
						Name:         fld.Name(),
						Doc:          fldDoc,
						Type:         fld.Type(),
						TypeName:     typeName,
						Anonymous:    fld.Anonymous(),
						JSONProperty: jsonProperty,
						JSONRequired: required,
					}
					structFields = append(structFields, f)
					l.logger.V(5).Info("added struct field definition", "struct", t.Name.Name, "field", f)
				}

				if len(structFields) == 0 {
					continue
				}

				apiType := Type{
					Name:           currentObj.Name,
					Package:        pkgPath,
					Doc:            strings.TrimSpace(astutils.TypeDoc(pkgDoc, currentObj.Name)),
					Fields:         structFields,
				}
				exportedTypes = append(exportedTypes, apiType)
			}
		}

		if len(exportedTypes) == 0 {
			l.logger.V(5).Info("skipping package - no exported types", "package", pkgPath)
			continue
		}

		loadedPackage := Package{
			Path:  pkg.Pkg.Path(),
			Types: exportedTypes,
			Doc:   pkgDoc.Doc,
		}
		loadedPackages = append(loadedPackages, loadedPackage)
	}

	return loadedPackages, nil
}

type StructTag struct {
	Name  string
	Value string
}

func (t StructTag) String() string {
	return fmt.Sprintf("%s:%q", t.Name, t.Value)
}

type StructTags []StructTag

func (tags StructTags) String() string {
	s := make([]string, 0, len(tags))
	for _, tag := range tags {
		s = append(s, tag.String())
	}
	return "`" + strings.Join(s, " ") + "`"
}

func (tags StructTags) Has(name string) bool {
	for i := range tags {
		if tags[i].Name == name {
			return true
		}
	}
	return false
}

// ParseStructTags returns the full set of fields in a struct tag in the order they appear in
// the struct tag.
func ParseStructTags(tag string) (StructTags, error) {
	tags := StructTags{}
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return nil, err
		}
		tags = append(tags, StructTag{Name: name, Value: value})
	}
	return tags, nil
}
