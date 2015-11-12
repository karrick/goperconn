package goperconn

import (
	"fmt"
	"io"
	"net"
	"time"
)

// DefaultJobQueueSize specifies the size of the job queue created to support IO operations on the
// Conn.
const DefaultJobQueueSize = 10

// DefaultRetryMin is the default minimum amount of time the client will wait to reconnect to a
// remote host if the connection drops.
const DefaultRetryMin = time.Second

// DefaultRetryMax is the default maximum amount of time the client will wait to reconnect to a
// remote host if the connection drops.
const DefaultRetryMax = time.Minute

// Configurator is a function that modifies a Conn structure during initialization.
type Configurator func(*Conn) error

// Address changes the network address used by a .
func Address(address string) Configurator {
	return func(c *Conn) error {
		c.address = address
		return nil
	}
}

// DialTimeout specifies the timeout to use when establishing the connection to the remote host.
func DialTimeout(duration time.Duration) Configurator {
	return func(c *Conn) error {
		c.dialTimeout = duration
		return nil
	}
}

// Printer interface exposes the Print method.
type Printer interface {
	Print(...interface{})
}

// Logger specifies an optional logger to invoke to log warning messages.
func Logger(printer Printer) Configurator {
	return func(c *Conn) error {
		c.printer = printer
		return nil
	}
}

// RetryMin controls the minimum amount of time a Conn will wait between connection attempts to the
// remote host.
func RetryMin(duration time.Duration) Configurator {
	return func(c *Conn) error {
		c.retryMin = duration
		return nil
	}
}

// RetryMax controls the maximum amount of time a Conn will wait between connection attempts to the
// remote host.
func RetryMax(duration time.Duration) Configurator {
	return func(c *Conn) error {
		c.retryMax = duration
		return nil
	}
}

// Conn wraps a net.Conn, providing a pseudo-persistent network connection.
type Conn struct {
	net.Conn
	address     string
	dialTimeout time.Duration
	jobs        chan *rillJob
	printer     Printer
	retryMax    time.Duration
	retryMin    time.Duration
}

// New returns a Conn structure that wraps the net.Conn connection, and attempts to provide a
// pseudo-persistent connection to a remote host.
//
//	package main
//
//	import (
//		"log"
//		"os"
//		"time"
//
//		"github.com/karrick/goperconn"
//	)
//
//	func main() {
//		printer := log.New(os.Stderr, "WARNING: ", 0)
//
//		// NOTE: Address is required, but all other parameters have defaults.
//		conn, err := goperconn.New(goperconn.Address("echo-server.example.com:7"),
//			goperconn.DialTimeout(5*time.Second),
//			goperconn.Logger(printer),
//			goperconn.RetryMin(time.Second),
//			goperconn.RetryMax(30*time.Second))
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// later ...
//
//		_, err = conn.Write([]byte("hello, world"))
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		buf := make([]byte, 512)
//		_, err = conn.Read(buf)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
func New(setters ...Configurator) (*Conn, error) {
	client := &Conn{
		retryMin: DefaultRetryMin,
		retryMax: DefaultRetryMax,
		jobs:     make(chan *rillJob, DefaultJobQueueSize),
	}
	for _, setter := range setters {
		if err := setter(client); err != nil {
			return nil, err
		}
	}
	if client.retryMin == 0 {
		return nil, fmt.Errorf("cannot create Conn with retry: %d", client.retryMin)
	}
	if client.retryMax == 0 {
		return nil, fmt.Errorf("cannot create Conn with retry: %d", client.retryMax)
	}
	if client.retryMax < client.retryMin {
		return nil, fmt.Errorf("cannot create Conn with retry max (%d) less than retry min (%d)", client.retryMax, client.retryMin)
	}
	if client.address == "" {
		return nil, fmt.Errorf("cannot create Conn with address: %q", client.address)
	}
	go func(wrapper *Conn) {
		var conn net.Conn
		var err error
		retry := client.retryMin
		for {
			if client.dialTimeout == 0 {
				conn, err = net.Dial("tcp", client.address)
			} else {
				conn, err = net.DialTimeout("tcp", client.address, client.dialTimeout)
			}
			if err != nil {
				if client.printer != nil {
					client.printer.Print(ErrDialFailure{client.address, err})
				}
				time.Sleep(retry)
				retry *= 2
				if retry > client.retryMax {
					retry = client.retryMax
				}
				continue
			}

			closed, err := wrapper.proxy(conn)
			if err != nil && client.printer != nil {
				client.printer.Print(err)
			}
			if closed {
				return
			}
			retry = client.retryMin
			time.Sleep(retry)
		}
	}(client)
	return client, nil
}

func (client *Conn) proxy(rwc io.ReadWriteCloser) (bool, error) {
	var closed bool
	for job := range client.jobs {
		if closed {
			job.results <- rillResult{err: ErrClosedConnection{}}
			continue
		}
		switch job.op {
		case _read:
			n, err := rwc.Read(job.data)
			job.results <- rillResult{n, err}
			if err != nil {
				rwc.Close()
				return false, err
			}
		case _write:
			n, err := rwc.Write(job.data)
			job.results <- rillResult{n, err}
			if err != nil {
				rwc.Close()
				return false, err
			}
		case _close:
			closed = true
			err := rwc.Close()
			job.results <- rillResult{err: err}
			return true, err
		}
	}
	return false, nil
}

// Read reads data from the connection.
func (client *Conn) Read(data []byte) (int, error) {
	job := newRillJob(_read, make([]byte, len(data)))
	client.jobs <- job

	result := <-job.results
	copy(data, job.data)
	if result.err != nil {
		result.err = ErrIOError{_read, result.err}
	}
	return result.n, result.err
}

// Write writes data to the connection.
func (client *Conn) Write(data []byte) (int, error) {
	job := newRillJob(_write, data)
	client.jobs <- job

	result := <-job.results
	if result.err != nil {
		result.err = ErrIOError{_write, result.err}
	}
	return result.n, result.err
}

// Close closes the connection.
func (client *Conn) Close() error {
	job := newRillJob(_close, nil)
	client.jobs <- job

	result := <-job.results
	if result.err != nil {
		result.err = ErrIOError{_close, result.err}
	}
	return result.err
}
