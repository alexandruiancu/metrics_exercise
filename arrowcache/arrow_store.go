package laarrow

import (
	"me/bldrec"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
)

// ArrowStore encapsulates arrow-related helpers and caching for schema and allocator.
type ArrowStore struct {
	mem    memory.Allocator
	schema *arrow.Schema
}

// NewArrowStore constructs a fresh ArrowStore with defaults.
func NewArrowStore() *ArrowStore {
	fields := []arrow.Field{
		{Name: "uDateTime", Type: arrow.PrimitiveTypes.Int64},
		{Name: "sDescription", Type: arrow.BinaryTypes.String},
		{Name: "fValue", Type: arrow.PrimitiveTypes.Float32},
		{Name: "sDontCare", Type: arrow.BinaryTypes.String},
	}
	return &ArrowStore{
		mem:    memory.NewGoAllocator(),
		schema: arrow.NewSchema(fields, nil),
	}
}

// Schema returns the immutable schema for records.
func (as *ArrowStore) Schema() *arrow.Schema {
	return as.schema
}

// Allocator returns the memory.Allocator used by this store.
func (as *ArrowStore) Allocator() memory.Allocator {
	return as.mem
}

// CreateArrowTable makes an empty table based on the store's schema.
func (as *ArrowStore) CreateArrowTable() (array.Table, error) {
	// Build a slice of empty columns matching the schema so that the
	// table validator does not panic. Each column uses an empty
	// Chunked array of the appropriate type.
	cols := make([]array.Column, len(as.schema.Fields()))
	for i, f := range as.schema.Fields() {
		colPtr := array.NewColumn(f, array.NewChunked(f.Type, nil))
		cols[i] = *colPtr
	}
	return array.NewTable(as.schema, cols, 0), nil
}

// ToArrowRecord converts a single bldrec.Record into an Arrow record.
func (as *ArrowStore) ToArrowRecord(rec bldrec.Record) (array.Record, error) {
	builder := array.NewRecordBuilder(as.mem, as.schema)
	defer builder.Release()

	handleErr := func(description string, err error) string {
		if err != nil {
			return ""
		}
		return description
	}

	builder.Field(0).(*array.Int64Builder).Append(rec.UDateTime())
	builder.Field(1).(*array.StringBuilder).Append(handleErr(rec.SDescription()))
	builder.Field(2).(*array.Float32Builder).Append(rec.FValue())
	builder.Field(3).(*array.StringBuilder).Append(handleErr(rec.SDontCare()))

	return builder.NewRecord(), nil
}
