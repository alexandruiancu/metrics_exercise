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

func getSchema() *arrow.Schema {

	// Combine arrays into a schema
	fields := []arrow.Field{
		{Name: "uDateTime", Type: arrow.PrimitiveTypes.Int64},
		{Name: "sDescription", Type: arrow.BinaryTypes.String},
		{Name: "fValue", Type: arrow.PrimitiveTypes.Float32},
		{Name: "sDontCare", Type: arrow.BinaryTypes.String},
	}

	schema := arrow.NewSchema(fields, nil)
	return schema
}

func getLocalAllocator() memory.Allocator {

	// Create a memory allocator
	mem := memory.NewGoAllocator()
	return mem
}

// Define a function to create an Arrow Table
func createArrowTable() (array.Table, error) {
	schema := getSchema()
	arrowTable := array.NewTable(schema, nil, 0)

	return arrowTable, nil
}

// Define a function to create an Arrow Table
func toArrowRecord(rec bldrec.Record) (array.Record, error) {
	mem := getLocalAllocator()
	schema := getSchema()
	builder := array.NewRecordBuilder(mem, schema)
	defer builder.Release()

	handle_err := func(description string, err error) string {
		if err != nil {
			return ""
		}

		return description
	}
	builder.Field(0).(*array.Int64Builder).Append(rec.UDateTime())
	builder.Field(1).(*array.StringBuilder).Append(handle_err(rec.SDescription()))
	builder.Field(2).(*array.Float32Builder).Append(rec.FValue())
	builder.Field(3).(*array.StringBuilder).Append(handle_err(rec.SDontCare()))

	return builder.NewRecord(), nil
}

// Define a function to create an Arrow Table
func buildRecords(bldRcrds []bldrec.Record) ([]array.Record, error) {

	var records []array.Record

	for _, r := range bldRcrds {
		record, err := toArrowRecord(r)
		if err != nil {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

func startWorker(id int) {
	socket, _ := zmq.NewSocket(zmq.REP)
	defer socket.Close()
	socket.Connect("tcp://localhost:5556")

	//table, _ := createArrowTable()
	// buffering
	var allRecords []array.Record

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
		newArrowRecords, _ := buildRecords([]bldrec.Record{record})
		for _, r := range newArrowRecords {
			allRecords = append(allRecords, r)
		}

		socket.Send(fmt.Sprintf("Reply from worker %d", tmp), 0)
	}

	//TODO
	//defer mem...
	//defer table...
}
