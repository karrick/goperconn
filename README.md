# goperconn

Goperconn is a Go library for maintaining pseudo-persitent client
network connections.

## Usage

Documentation is available via
[![GoDoc](https://godoc.org/github.com/karrick/goperconn?status.svg)](https://godoc.org/github.com/karrick/goperconn).

### Basic Example

Only the Address of the remote host is required. All other parameters
have reasonalbe defaults and may be elided.

```Go
        package main

        import (
            "log"
            "os"
            "time"

            goperconn "gopkg.in/karrick/goperconn.v1"
        )

        func main() {
            // NOTE: Address is required, but all other parameters have defaults.
            conn, err := goperconn.New(goperconn.Address("echo-server.example.com:7"))
            if err != nil {
                log.Fatal(err)
            }

            // later ...

            _, err = conn.Write([]byte("hello, world"))
            if err != nil {
                log.Fatal(err)
            }

            buf := make([]byte, 512)
            _, err = conn.Read(buf)
            if err != nil {
                log.Fatal(err)
            }
        }
```

### DialTimeout

By default the library uses `net.Dial` to establish a new network
connection. If you specify a custom dial timeout, it will use
`net.DialTimeout` instead.

```Go
        package main

        import (
            "log"
            "os"
            "time"

            goperconn "gopkg.in/karrick/goperconn.v1"
        )

        func main() {
            // NOTE: Address is required, but all other parameters have defaults.
            conn, err := goperconn.New(goperconn.Address("echo-server.example.com:7"),
                goperconn.DialTimeout(5*time.Second))
            if err != nil {
                log.Fatal(err)
            }

            // later ...

            _, err = conn.Write([]byte("hello, world"))
            if err != nil {
                log.Fatal(err)
            }

            buf := make([]byte, 512)
            _, err = conn.Read(buf)
            if err != nil {
                log.Fatal(err)
            }
        }
```

### Logger

If the library receives an error when attempting to dial the remote
host, or if there is an error that takes place during an I/O
operation, the library handles the error. Additionally, when the error
takes place during an I/O operation, the library will return the error
to the client just like the Read or Write methods are expected
to. However, when a custom Logger is provided, the library will also
call the logger's Print function.

```Go
        package main

        import (
            "log"
            "os"
            "time"

            goperconn "gopkg.in/karrick/goperconn.v1"
        )

        func main() {
            printer := log.New(os.Stderr, "WARNING: ", 0)

            // NOTE: Address is required, but all other parameters have defaults.
            conn, err := goperconn.New(goperconn.Address("echo-server.example.com:7"),
                goperconn.Logger(printer))
            if err != nil {
                log.Fatal(err)
            }

            // later ...

            _, err = conn.Write([]byte("hello, world"))
            if err != nil {
                log.Fatal(err)
            }

            buf := make([]byte, 512)
            _, err = conn.Read(buf)
            if err != nil {
                log.Fatal(err)
            }
        }
```
### RetryMin and RetryMax

The library will attempt to reestablish the network connection if it
breaks down. It uses an exponential backoff retry approach, bounded by
RetryMin and RetryMax. You can override the minimum and maximum amount
of time between connection retry attempts.

```Go
        package main

        import (
            "log"
            "os"
            "time"

            goperconn "gopkg.in/karrick/goperconn.v1"
        )

        func main() {
            // NOTE: Address is required, but all other parameters have defaults.
            conn, err := goperconn.New(goperconn.Address("echo-server.example.com:7"),
                goperconn.RetryMin(time.Second),
                goperconn.RetryMax(30*time.Second))
            if err != nil {
                log.Fatal(err)
            }

            // later ...

            _, err = conn.Write([]byte("hello, world"))
            if err != nil {
                log.Fatal(err)
            }

            buf := make([]byte, 512)
            _, err = conn.Read(buf)
            if err != nil {
                log.Fatal(err)
            }
        }
```


### The Whole Enchalata

The Address of the remote connection must be specified, but all other
customizable parameters are optional, and may be given in any order,
in the form of a slice of function calls in the `goperconn.New`
invocation.

```Go
        package main

        import (
            "log"
            "os"
            "time"

            goperconn "gopkg.in/karrick/goperconn.v1"
        )

        func main() {
            printer := log.New(os.Stderr, "WARNING: ", 0)

            // NOTE: Address is required, but all other parameters have defaults.
            conn, err := goperconn.New(goperconn.Address("echo-server.example.com:7"),
                goperconn.DialTimeout(5*time.Second),
                goperconn.Logger(printer),
                goperconn.RetryMin(time.Second),
                goperconn.RetryMax(30*time.Second))
            if err != nil {
                log.Fatal(err)
            }

            // later ...

            _, err = conn.Write([]byte("hello, world"))
            if err != nil {
                log.Fatal(err)
            }

            buf := make([]byte, 512)
            _, err = conn.Read(buf)
            if err != nil {
                log.Fatal(err)
            }
        }
```
