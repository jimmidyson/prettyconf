package printer_test

import (
	"bytes"
	"testing"

	"github.com/jimmidyson/prettyconf/pkg/printer"
)

func TestPrettyPrint(t *testing.T) {
	type args struct {
		conf interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			"a",
			args{
				conf: ClusterProvisioner{},
			},
			"hello",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := printer.PrettyPrint(tt.args.conf, w); (err != nil) != tt.wantErr {
				t.Errorf("PrettyPrint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("PrettyPrint() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
