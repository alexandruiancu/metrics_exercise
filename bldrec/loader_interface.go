package bldrec

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

// LoaderClient defines the interface for communicating with the loader
type LoaderClient interface {
	SendRecord(record Record) (string, error)
	Close() error
}

// RealLoaderClient implements the interface using ZMQ
type RealLoaderClient struct {
	socket *zmq.Socket
}

// NewRealLoaderClient creates a new ZMQ-based loader client
func NewRealLoaderClient(port string) (*RealLoaderClient, error) {
	socket, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZMQ socket: %w", err)
	}
	socket.Connect(fmt.Sprintf("tcp://localhost:%s", port))
	return &RealLoaderClient{socket: socket}, nil
}

// SendRecord sends a record to the loader and waits for a reply
func (r *RealLoaderClient) SendRecord(record Record) (string, error) {
	data, err := record.Message().Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal record: %w", err)
	}
	bytes, err := r.socket.SendBytes(data, 0)
	if err != nil || bytes != len(data) {
		return "", fmt.Errorf("failed to send record: %w", err)
	}
	reply, err := r.socket.Recv(0)
	if err != nil {
		return "", fmt.Errorf("failed to receive reply: %w", err)
	}
	return reply, nil
}

// Close closes the ZMQ socket
func (r *RealLoaderClient) Close() error {
	if r.socket != nil {
		return r.socket.Close()
	}
	return nil
}

// MockLoaderClient is a test double for testing without a real loader
type MockLoaderClient struct {
	RecordsSent []Record
	Replies     []string
	Error       error
}

// NewMockLoaderClient creates a new mock loader client
func NewMockLoaderClient() *MockLoaderClient {
	return &MockLoaderClient{
		RecordsSent: make([]Record, 0),
		Replies:     make([]string, 0),
	}
}

// SendRecord records the sent record and returns a predefined reply
func (m *MockLoaderClient) SendRecord(record Record) (string, error) {
	m.RecordsSent = append(m.RecordsSent, record)
	if m.Error != nil {
		return "", m.Error
	}
	if len(m.Replies) > 0 {
		reply := m.Replies[0]
		m.Replies = m.Replies[1:]
		return reply, nil
	}
	return "ACK", nil
}

// Close does nothing for the mock client
func (m *MockLoaderClient) Close() error {
	return nil
}

// SetReplies sets the replies that will be returned by SendRecord
func (m *MockLoaderClient) SetReplies(replies []string) {
	m.Replies = replies
}

// SetError sets an error that will be returned by SendRecord
func (m *MockLoaderClient) SetError(err error) {
	m.Error = err
}
