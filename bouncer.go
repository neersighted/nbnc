package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
)

func bouncer(config *Config) {
	var err error

	// Start listening on the designated address.
	var server net.Listener
	if config.Cert != "" && config.Key != "" {
		var certificate, err = tls.LoadX509KeyPair(config.Cert, config.Key)
		if err != nil {
			log.Println(err)
			return
		}
		var tlsconfig *tls.Config = &tls.Config{Certificates: []tls.Certificate{certificate}}

		server, err = tls.Listen("tcp", config.Listen, tlsconfig)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("SLISTEN %s", server.Addr())
	} else {
		server, err = net.Listen("tcp", config.Listen)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("LISTEN %s", server.Addr())
	}

	for {
		// Block and accept incoming connections.
		var client net.Conn
		client, err = server.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Bounce off the connection.
		go bounce(client, config)
	}
}

func bounce(client net.Conn, config *Config) {
	var err error

	// Always clean up...
	defer func() {
		log.Printf("CLOSE %s", client.RemoteAddr())
		client.Close()
	}()
	log.Printf("OPEN %s", client.RemoteAddr())

	// Determine which bouncer config to use.
	var (
		bouncer   *BouncerConfig
		remainder []byte
	)
	if bouncer, remainder = handshake(client, config); bouncer == nil {
		return
	}

	// Create a correctly bound local dialer.
	var dialer net.Dialer = net.Dialer{LocalAddr: &net.TCPAddr{IP: net.ParseIP(bouncer.Bind)}}

	// Connect to the remote server.
	var server net.Conn
	if bouncer.Secure {
		// Use TLS if configured.
		var tlsconfig *tls.Config = &tls.Config{InsecureSkipVerify: bouncer.NoVerify}

		server, err = tls.DialWithDialer(&dialer, "tcp", bouncer.Target, tlsconfig)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("SDIAL %s", server.RemoteAddr())
	} else {
		// Otherwise make a plaintext connection.
		server, err = dialer.Dial("tcp", bouncer.Target)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("DIAL %s", server.RemoteAddr())
	}

	// Always clean up...
	defer func() {
		log.Printf("HANGUP %s", server.RemoteAddr())
		server.Close()
	}()

	// Write the part of the handshake not directed at us.
	server.Write(remainder)

	// Copy data client <-> server, blocking until one of them passes 'done' back.
	var done chan bool = make(chan bool)
	go relay(server, client, done)
	go relay(client, server, done)
	<-done
}

func relay(conn1 net.Conn, conn2 net.Conn, done chan<- bool) {
	io.Copy(conn1, conn2)
	done <- true
}
