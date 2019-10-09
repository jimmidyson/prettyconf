package printer

import (
	"encoding/json"
	"fmt"
	"go/types"
	"io"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/jimmidyson/prettyconf/pkg/loader"
)

// PrettyPrint prints the passed in conf to the writer w, including all fields and comments if
// parsed from the package.
func PrettyPrint(conf interface{}, w io.Writer, logger logr.Logger) error {
	confType := reflect.TypeOf(conf)

	confTypePkgPath := confType.PkgPath()

	astLoader := loader.New([]string{confTypePkgPath}, logger)
	packages, err := astLoader.Load()
	if err != nil {
		return errors.Wrapf(err, "failed to parse package %s", confTypePkgPath)
	}

	pkg, found := filterPackage(confTypePkgPath, packages)
	if !found {
		return errors.Errorf("package %s could not be found", confTypePkgPath)
	}
	pkgType, found := filterType(confType.Name(), pkg)
	if !found {
		return errors.Errorf("type %s.%s could not be found", confTypePkgPath, confType.Name())
	}

	marshalledConfig, err := json.Marshal(conf)
	if err != nil {
		return errors.Wrap(err, "failed to marshal initial config to json")
	}

	var unmarshaledConfigToMap map[string]interface{}
	if err := json.Unmarshal(marshalledConfig, &unmarshaledConfigToMap); err != nil {
		return errors.Wrap(err, "failed to unmarshal config to map")
	}

	if err := zeroUnsetFields(unmarshaledConfigToMap, pkgType.Fields, packages); err != nil {
		return errors.Wrapf(err, "failed to zero unset fields")
	}

	marshalledConfig, err = yaml.Marshal(unmarshaledConfigToMap)
	if err != nil {
		return errors.Wrap(err, "failed to marshal initial config to yaml")
	}

	var unmarshalledDocumentNode yaml.Node
	if err := yaml.Unmarshal(marshalledConfig, &unmarshalledDocumentNode); err != nil {
		return errors.Wrap(err, "failed to unmarshal initial config to yaml node")
	}

	if unmarshalledDocumentNode.Kind != yaml.DocumentNode {
		return errors.New("expected a single YAML document node")
	}

	if len(unmarshalledDocumentNode.Content) > 1 {
		return errors.New("should only have one YAML node in document")
	}

	currentNode := unmarshalledDocumentNode.Content[0]

	if pkgType.Doc != "" {
		currentNode.HeadComment = pkgType.Doc + "\n\n"
	}

	if err := visitContentNodes(currentNode, pkgType, packages); err != nil {
		return errors.Wrap(err,"failed to visit all nodes")
	}

	marshalledConfig, err = yaml.Marshal(&unmarshalledDocumentNode)
	if err != nil {
		return errors.Wrap(err,"failed to marshal commented yaml node")
	}

	fmt.Fprintln(w, string(marshalledConfig))

	return nil
}

func zeroUnsetFields(unmarshaledConfigToMap map[string]interface{}, fields []loader.Field, packages []loader.Package) error {
	for _, field := range fields {
		if _, ok := unmarshaledConfigToMap[field.JSONProperty]; !ok {
			zeroValue, err := zeroPropertyForType(field.Type)
			if err != nil {
				return errors.Wrapf(err, "failed to set zero property value for %s", field.Name)
			}
			unmarshaledConfigToMap[field.JSONProperty] = zeroValue
		}
		switch t := field.Type.(type) {
		case *types.Named:
			switch t.Underlying().(type) {
			case *types.Struct:
				fieldType, err := fieldPkgPathAndName(t, packages)
				if err != nil {
					return err
				}
				zeroUnsetFields(unmarshaledConfigToMap[field.JSONProperty].(map[string]interface{}), fieldType.Fields, packages)
			}
		case *types.Pointer:
			switch t := t.Elem().(type) {
			case *types.Struct:
				fieldType, err := fieldPkgPathAndName(t, packages)
				if err != nil {
					return err
				}
				zeroUnsetFields(unmarshaledConfigToMap[field.JSONProperty].(map[string]interface{}), fieldType.Fields, packages)
			default:
				fieldType, err := fieldPkgPathAndName(t, packages)
				if err != nil {
					return err
				}
				zeroUnsetFields(unmarshaledConfigToMap[field.JSONProperty].(map[string]interface{}), fieldType.Fields, packages)
			}
		}
	}
	return nil
}

func zeroPropertyForType(fieldType types.Type) (interface{}, error) {
	switch t := fieldType.(type) {
	case *types.Named:
		return zeroPropertyForType(t.Underlying())
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return false, nil
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
			return 0, nil
		case types.String:
			return "", nil
		default:
			return nil, fmt.Errorf("unhandled field type: %v", t.Kind())
		}
	case *types.Pointer:
		return zeroPropertyForType(t.Elem())
	case *types.Slice:
		return []struct{}{}, nil
	case *types.Map, *types.Struct:
		return map[string]interface{}{}, nil
	default:
		return nil, fmt.Errorf("unhandled node content type: %s", reflect.TypeOf(t))
	}
}

func visitContentNodes(node *yaml.Node, pkgType loader.Type, packages []loader.Package) error {
	for i, contentNode := range node.Content {
		if i%2 != 0 {
			continue
		}
		contentNodeName := contentNode.Value
		if contentNodeName == "" {
			continue
		}
		field, found := filterField(contentNodeName, pkgType)
		if !found {
			return errors.Errorf("failed to find field %s in type %s.%s", contentNodeName, pkgType.Package, pkgType.Name)
		}
		doc := field.Doc
		if strings.HasPrefix(doc, field.Name + " ") {
			doc = field.JSONProperty + doc[len(field.Name):]
		}
		contentNode.HeadComment = doc
		valueContentNode := node.Content[i+1]
		if valueContentNode.Kind == yaml.MappingNode {
			pkgType, err := fieldPkgPathAndName(field.Type, packages)
			if err != nil {
				return err
			}
			if err := visitContentNodes(valueContentNode, pkgType, packages); err != nil {
				return err
			}
		}
	}
	node.Content = sortContentNodes(pkgType.Fields, node.Content)
	return nil
}

func sortContentNodes(fields []loader.Field, contentNodes []*yaml.Node) []*yaml.Node {
	sortedContentNodes := make([]*yaml.Node, 0, len(contentNodes))
	for _, field := range fields {
		for i, contentNode := range contentNodes {
			if i%2 != 0 {
				continue
			}
			if contentNode.Value == field.JSONProperty {
				sortedContentNodes = append(sortedContentNodes, contentNodes[i:i+2]...)
			}
		}
	}
	return sortedContentNodes
}

func filterField(fieldJSONTagName string, pkgType loader.Type) (field loader.Field, found bool) {
	for _, f := range pkgType.Fields {
		if f.JSONProperty == fieldJSONTagName {
			return f, true
		}
	}
	return loader.Field{}, false
}

func filterPackage(packagePath string, packages []loader.Package) (pkg loader.Package, found bool) {
	for _, p := range packages {
		if packagePath == p.Path {
			return p, true
		}
	}

	return loader.Package{}, false
}

func filterType(typeName string, pkg loader.Package) (t loader.Type, found bool) {
	for _, p := range pkg.Types {
		if typeName == p.Name {
			return p, true
		}
	}

	return loader.Type{}, false
}

func fieldPkgPathAndName(fieldType types.Type, packages []loader.Package) (loader.Type, error) {
	switch t := fieldType.(type) {
	case *types.Named:
		pkgPath := t.Obj().Pkg().Path()
		pkg, found := filterPackage(pkgPath, packages)
		if !found {
			return loader.Type{}, errors.Errorf("package %s could not be found", pkgPath)
		}
		pkgTypeName := t.Obj().Name()
		pkgType, found := filterType(pkgTypeName, pkg)
		if !found {
			return loader.Type{}, errors.Errorf("type %s.%s could not be found", pkgPath, pkgTypeName)
		}
		return pkgType, nil
	case *types.Pointer:
		return fieldPkgPathAndName(t.Elem(), packages)
	default:
		return loader.Type{}, fmt.Errorf("unhandled node content type: %s", reflect.TypeOf(t))
	}
}
