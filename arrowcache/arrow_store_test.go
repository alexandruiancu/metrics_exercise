package laarrow

import (
	"testing"

	"me/bldrec"

	capnp "capnproto.org/go/capnp/v3"
	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
)

func makeTestBldRec(t *testing.T) bldrec.Record {
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		t.Fatalf("failed to create capnp message: %v", err)
	}
	rec, err := bldrec.NewRootRecord(seg)
	if err != nil {
		t.Fatalf("failed to create root record: %v", err)
	}

	// populate fields
	rec.SetUDateTime(1618033988)
	rec.SetSDescription("test-desc")
	rec.SetFValue(float32(3.14))
	rec.SetSDontCare("ignore")
	return rec
}

func TestNewArrowStoreBasics(t *testing.T) {
	as := NewArrowStore()
	if as == nil {
		t.Fatalf("NewArrowStore returned nil")
	}
	if as.Schema() == nil {
		t.Error("Schema should not be nil")
	}
	if as.Allocator() == nil {
		t.Error("Allocator should not be nil")
	}
}

func TestSchemaFields(t *testing.T) {
	as := NewArrowStore()
	schema := as.Schema()
	if len(schema.Fields()) != 4 {
		t.Errorf("expected 4 fields, got %d", len(schema.Fields()))
	}
	// check each name/type pair
	exp := []struct {
		name string
		typ  arrow.DataType
	}{
		{"uDateTime", arrow.PrimitiveTypes.Int64},
		{"sDescription", arrow.BinaryTypes.String},
		{"fValue", arrow.PrimitiveTypes.Float32},
		{"sDontCare", arrow.BinaryTypes.String},
	}
	for i, e := range exp {
		f := schema.Field(i)
		if f.Name != e.name {
			t.Errorf("field %d name expected %s got %s", i, e.name, f.Name)
		}
		if f.Type != e.typ {
			t.Errorf("field %d type expected %v got %v", i, e.typ, f.Type)
		}
	}
}

func TestCreateArrowTable(t *testing.T) {
	as := NewArrowStore()
	tbl, err := as.CreateArrowTable()
	if err != nil {
		t.Fatalf("CreateArrowTable failed: %v", err)
	}
	if tbl == nil {
		t.Fatal("table should not be nil")
	}
	if !tbl.Schema().Equal(as.Schema()) {
		t.Errorf("table schema does not match store schema")
	}
	if tbl.NumRows() != 0 {
		t.Errorf("new table should have 0 rows, got %d", tbl.NumRows())
	}
}

func TestToArrowRecord(t *testing.T) {
	as := NewArrowStore()
	rec := makeTestBldRec(t)
	arrRec, err := as.ToArrowRecord(rec)
	if err != nil {
		t.Fatalf("ToArrowRecord failed: %v", err)
	}
	if arrRec == nil {
		t.Fatal("returned Arrow record is nil")
	}
	if arrRec.NumRows() != 1 {
		t.Errorf("expected 1 row, got %d", arrRec.NumRows())
	}
	// check values in each column
	v0 := arrRec.Column(0).(*array.Int64)
	if v0.Value(0) != 1618033988 {
		t.Errorf("uDateTime mismatch, got %d", v0.Value(0))
	}
	v1 := arrRec.Column(1).(*array.String)
	if v1.Value(0) != "test-desc" {
		t.Errorf("sDescription mismatch, got %s", v1.Value(0))
	}
	v2 := arrRec.Column(2).(*array.Float32)
	if v2.Value(0) != float32(3.14) {
		t.Errorf("fValue mismatch, got %f", v2.Value(0))
	}
	v3 := arrRec.Column(3).(*array.String)
	if v3.Value(0) != "ignore" {
		t.Errorf("sDontCare mismatch, got %s", v3.Value(0))
	}
}
