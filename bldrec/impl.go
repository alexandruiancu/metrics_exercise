package bldrec

import (
	"fmt"
	"log"
	"path/filepath"

	zmq "github.com/pebbe/zmq4"
)

func Process(config map[string]string) error {
	inDir := config["in_dir"]
	if !filepath.IsAbs(inDir) && inDir != "" {
		inDir = filepath.Join(config["configDir"], config["in_dir"])
	}
	historyDir := config["history_dir"]
	if !filepath.IsAbs(historyDir) && historyDir != "" {
		historyDir = filepath.Join(config["configDir"], config["history_dir"])
	}

	port := config["frontend_port"]
	socket, _ := zmq.NewSocket(zmq.REQ)
	defer socket.Close()
	socket.Connect(fmt.Sprintf("tcp://localhost:%s", port))
	for true {
		fp := FileProcessor{}
		records, err := fp.ProcessFiles(inDir, historyDir)
		if err != nil {
			return err
		}

		for _, record := range records {
			// Serialize to a byte slice
			data, err := record.Message().Marshal()
			if err != nil {
				log.Fatalf("marshal: %v", err)
			}
			socket.SendBytes(data, 0)
			// Receive reply
			reply, _ := socket.Recv(0)
			fmt.Printf("Received reply: %s\n", reply)
		}
	}

	return nil
}

// will add to record batch
func arrowRecord(fields []string) error {
	//TODO: Implement Arrow record creation and population based on the fields array
	return nil
}
