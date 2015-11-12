package main

import (
	"log"
	"os"
	"time"

	"github.com/karrick/goperconn"
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
