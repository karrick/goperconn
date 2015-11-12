package main

import (
	"log"

	"github.com/karrick/goperconn"
)

func main() {
	warning := func(s string) {
		log.Printf("WARNING: %s", s)
	}

	conn, err := goperconn.New(goperconn.Address("localhost:8080"),
		goperconn.Warning(warning))
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
