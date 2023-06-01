package fleetpkg

import "reflect"

type FileMetadata struct {
	file   string `json:"-"` // file from which field was read.
	line   int    `json:"-"` // line from which field was read.
	column int    `json:"-"` // column from which field was read.
}

func (m FileMetadata) Path() string { return m.file }
func (m FileMetadata) Line() int    { return m.line }
func (m FileMetadata) Column() int  { return m.column }

// annotateFileMetadata sets the file name on any types that contains FileMetadata.
func annotateFileMetadata(file string, v any) {
	fileAnnotator{Name: file}.Annotate(reflect.ValueOf(v))
}

type fileAnnotator struct {
	Name string
}

func (a fileAnnotator) Annotate(val reflect.Value) {
	// Need an addressable value in order to edit the metadata value.
	if val.CanAddr() && val.CanSet() {
		if m, ok := val.Addr().Interface().(*FileMetadata); ok {
			m.file = a.Name
			return
		}
	}

	switch val.Kind() {
	case reflect.Pointer:
		a.Annotate(val.Elem())
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			valueField := val.Field(i)
			a.Annotate(valueField)
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			a.Annotate(val.Index(i))
		}
	case reflect.Map:
		itr := val.MapRange()
		for itr.Next() {
			// NOTE: This can only edit the map value if it is addressable (aka a pointer).
			a.Annotate(itr.Value())
		}
	}
}
