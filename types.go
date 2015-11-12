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

// ErrDialFailure is optionally sent to the configured warning hookback when net.DialTimeout fails.
// The library will continue to attempt to reestablish the connection, but this error is useful for
// client application logging purposes.
type ErrDialFailure struct {
	Address string
	Err     error
}

func (e ErrDialFailure) Error() string {
	// NOTE: Address is stored for error for the client, but Err also includes the string, so we
	// don't need to print it here.
	return "cannot connect: " + e.Err.Error()
}

// ErrIOError is optionally sent to the configured warning hookback when an I/O operation fails. The
// library will close and attempt to reestablish the connection, but this error is useful for client
// application logging purposes.
type ErrIOError struct {
	Op  opcode
	Err error
}

func (e ErrIOError) Error() string {
	switch e.Op {
	case _read:
		return "cannot read: " + e.Err.Error()
	case _write:
		return "cannot write: " + e.Err.Error()
	case _close:
		return "cannot close: " + e.Err.Error()
	default:
		return "unknown error: " + e.Err.Error()
	}
}
