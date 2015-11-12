package goperconn

type opcode byte

const (
	_close opcode = iota
	_read
	_write
)

// rillJob represents a job to perform either a read or write operation to a stream
type rillJob struct {
	op      opcode
	data    []byte
	results chan rillResult
}

func newRillJob(op opcode, data []byte) *rillJob {
	return &rillJob{op: op, data: data, results: make(chan rillResult, 1)}
}

// rillResult represents the return values for a read or write operation to a stream
type rillResult struct {
	n   int
	err error
}

// ErrClosedConnection is returned when I/O operation attempted on closed connection.
type ErrClosedConnection struct{}

func (e ErrClosedConnection) Error() string {
	return "cannot perform I/O on closed connection"
}
