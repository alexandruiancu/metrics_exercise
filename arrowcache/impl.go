// ////////////////////
// Apache Arrow loader
package laarrow

import (
	"fmt"
	"log"

	"me/bldrec"

	capnp "capnproto.org/go/capnp/v3"
	zmq "github.com/pebbe/zmq4"

	"github.com/apache/arrow/go/arrow/array"
)

// buildRecords converts a slice of bldrec.Record using the provided ArrowStore.
func buildRecords(as *ArrowStore, bldRcrds []bldrec.Record) ([]array.Record, error) {

	var records []array.Record

	for _, r := range bldRcrds {
		record, err := as.ToArrowRecord(r)
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

	store := NewArrowStore()
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
		newArrowRecords, _ := buildRecords(store, []bldrec.Record{record})
		for _, r := range newArrowRecords {
			allRecords = append(allRecords, r)
		}

		socket.Send(fmt.Sprintf("Reply from worker %d", tmp), 0)
	}

	//TODO: release store resources if necessary
}
