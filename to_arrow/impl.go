// ////////////////////
// Apache Arrow loader
package laarrow

import (
	"fmt"
	"log"

	"me/bldrec"

	capnp "capnproto.org/go/capnp/v3"
	zmq "github.com/pebbe/zmq4"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
)

// Define a function to create an Arrow Table
func createArrowTable() (*array.Table, error) {

	// Combine arrays into a schema
	fields := []arrow.Field{
		{Name: "uDateTime", Type: arrow.PrimitiveTypes.Int64},
		{Name: "sDescription", Type: arrow.BinaryTypes.String},
		{Name: "fValue", Type: arrow.PrimitiveTypes.Float32},
		{Name: "sDontCare", Type: arrow.BinaryTypes.String},
	}

	schema := arrow.NewSchema(fields, nil)
	arrowTable := array.NewTable(schema, nil, 0)

	return arrowTable, nil
}

// Define a function to create an Arrow Table
func toArrowRecord(rec bldrec.Record) (array.Record, error) {
	// Create a memory allocator
	mem := memory.NewGoAllocator()

	// Combine arrays into a schema
	fields := []arrow.Field{
		{Name: "uDateTime", Type: arrow.PrimitiveTypes.Int64},
		{Name: "sDescription", Type: arrow.BinaryTypes.String},
		{Name: "fValue", Type: arrow.PrimitiveTypes.Float32},
		{Name: "sDontCare", Type: arrow.BinaryTypes.String},
	}

	schema := arrow.NewSchema(fields, nil)
	builder := array.NewRecordBuilder(mem, schema)
	defer builder.Release()

	builder.Field(0).(*array.Int64Builder).AppendValues({rec.UDateTime()}, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues({rec.SDescription()}, nil)
	builder.Field(2).(*array.Float32Builder).AppendValues({rec.FValue()}, nil)
	builder.Field(3).(*array.StringBuilder).AppendValues({rec.SDontCare()}, nil)

	return builder.NewRecord(), nil
}

// Define a function to create an Arrow Table
func buildRecords(bldRcrds []bldrec.Record) ([]array.Record, error) {

	var records []array.Record

	for i, r := range bldRcrds {
		records = append(toArrowRecord(r))
	}

	return records, nil
}

func startWorker(id int) {
	socket, _ := zmq.NewSocket(zmq.REP)
	defer socket.Close()
	socket.Connect("tcp://localhost:5556")

	table, _ := createArrowTable()
	for {
		zmqMsgBytes, _ := socket.RecvBytes(0)
		// Wrap in a Cap’n Proto message (read‑only)
		msg, err := capnp.Unmarshal(zmqMsgBytes)
		if err != nil {
			log.Fatalf("capnp message: %v", err)
		}
		record, err := bldrec.ReadRootRecord(msg)
		if err != nil {
			log.Fatalf("read struct: %v", err)
		}
		desc, _ := record.SDescription()
		tmp, err := fmt.Printf("Worker %d received: %s\n", id, desc)
		if err != nil {
			continue
		}
		//println(desc)
		arrowRec, _ = toArrowRecord(record)
		socket.Send(fmt.Sprintf("Reply from worker %d", tmp), 0)
	}
}
