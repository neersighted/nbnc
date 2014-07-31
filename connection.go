package main

import (
	"bufio"
	"log"
	"net"
)

type Connection struct {
	Socket   net.Conn
	Address  net.Addr
	Reader   *bufio.Reader
	Writer   *bufio.Writer
	Incoming chan string
	Outgoing chan string
	Complete chan bool
}

func (conn *Connection) Read(done, writeDone chan bool) {
	var (
		data string
		err  error
	)

For:
	// For to check if we're finished.
	for {
		select {
		case <-writeDone:
			break For
		default:
			data, err = conn.Reader.ReadString('\n')
			if err != nil {
				break For
			}

			log.Printf("[%s] %s", conn.Socket.RemoteAddr(), data)
			conn.Incoming <- data
		}
	}

	done <- true
}

func (conn *Connection) Write(done, readDone chan bool) {
	var (
		data string
		err  error
	)

	// For to check if we're finished.
For:
	for {
		select {
		case <-readDone:
			break For
		case data = <-conn.Outgoing:
			_, err = conn.Writer.WriteString(data)
			if err != nil {
				break For
			}

			err = conn.Writer.Flush()
			if err != nil {
				break For
			}
		}
	}

	done <- true
}

func (conn *Connection) Listen(readDone, writeDone chan bool) {
	go conn.Read(readDone, writeDone)
	go conn.Write(writeDone, readDone)
}

func (conn *Connection) Close() {
	conn.Socket.Close()
}

func NewConnection(sock net.Conn) *Connection {
	var (
		writer *bufio.Writer
		reader *bufio.Reader
		conn   *Connection

		complete  chan bool
		writeDone chan bool
		readDone  chan bool
	)

	// Create bufio readers/writers.
	writer = bufio.NewWriter(sock)
	reader = bufio.NewReader(sock)

	// Create a master notification channel.
	complete = make(chan bool, 1)

	// Initialize the connection object.
	conn = &Connection{
		Socket:   sock,
		Address:  sock.RemoteAddr(),
		Reader:   reader,
		Writer:   writer,
		Incoming: make(chan string),
		Outgoing: make(chan string),
		Complete: complete,
	}

	// Create locks.
	readDone = make(chan bool, 1)
	writeDone = make(chan bool, 1)

	conn.Listen(readDone, writeDone)

	go func(complete, readDone, writeDone chan bool) {
	For:
		for {
			select {
			case <-readDone:
				complete <- true
				break For
			case <-writeDone:
				complete <- true
				break For
			}
		}
	}(complete, readDone, writeDone)

	// Return the pointer to the connection.
	return conn
}

// vi: ts=4
