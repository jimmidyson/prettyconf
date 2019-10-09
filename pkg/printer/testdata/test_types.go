package testdata

// TopLevel holds the details for top level config.
type TopLevel struct {
	// A is field for AStruct.
	A AStruct `json:"a"`
	C CStruct `json:"cnocomment"`
	// B holds the comment here.
	B *BStruct `json:"b,omitempty"`
	// I holds a slice.
	I []*BStruct `json:"bs,omitempty"`
}

// AStruct holds some fields.
type AStruct struct {
	// E comment.
	E NestedStruct `json:"enested"`
	// D comment.
	D int `json:"d,omitempty"`
}

// NestedStruct holds nested struct fields.
type NestedStruct struct {
	// F comment.
	F string `json:"f,omitempty"`
}

// BStruct holds B fields.
type BStruct struct {
	// G comment.
	G string `json:"g,omitempty"`
}

// CStruct holds C fields.
type CStruct struct {
	// H comment.
	H string `json:"h,omitempty"`
}
