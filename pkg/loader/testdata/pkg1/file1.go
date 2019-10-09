package pkg1

// Type1 is a normal type
// with a single field and a description.
type Type1 struct {
	// Some doc.
	Field1 int
	Field2 string `json:"f2"`
	// Some more doc.
	Field3 []string `json:"-"`
	// Even more doc.
	Field4 []string `json:",omitempty"`
	// And some
	// more doc.
	Field5 map[string]bool `json:"f5,omitempty"`
	Type5  `json:",omitempty"`
	Type5s []Type5 `json:"t5s,omitempty"`
}

// Type2 is a normal type
// with a single unexported field and a description.
type Type2 struct {
	field1 int
}

// type3 is an unexported type
// with a single unexported field and a description.
type type3 struct {
	field1 int
}

// type4 is an unexported type
// with a single exported field and a description.
type type4 struct {
	Field1 int
}

type Type5 struct {
	// Something.
	Type5Field uint32 `json:"t5"`

	// Something else.
	Type5Field2 []uint32 `json:"t6"`
}
