package printer

import (
	"fmt"
	"io"
	"reflect"

	"golang.org/x/tools/go/loader"
)

// PrettyPrint prints the passed in conf to the writer w, including all fields and comments if
// parsed from the package.
func PrettyPrint(conf interface{}, w io.Writer) error {
	confType := reflect.TypeOf(conf)

	confTypePkgPath := confType.PkgPath()
	var loaderConf loader.Config
	loaderConf.Import(confTypePkgPath)
	loaderProgram, err := loaderConf.Load()
	if err != nil {
		return fmt.Errorf("failed to load go package %s: %w", confTypePkgPath, err)
	}

	fmt.Println(loaderProgram.Package(confTypePkgPath))

	return nil
}
