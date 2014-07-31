package main

import (
	"github.com/codegangsta/cli"
	"log"
	"net"
	"os"
	"strings"
)

const (
	Name        string = "nbnc"
	Version     string = "0.0.1"
	Description string = "simple null (transparent) bnc"
)

var (
	opt Options
)

type Options struct {
	ListenAddr  string
	ListenPort  string
	ConnectAddr string
	ConnectPort string

	Password string
}

func _main() {
	// Print program information.
	log.Printf("%s %s",
		Name, Version)
	// Also print exactly what we're doing.
	log.Printf("Proxying %s:%s -> %s:%s",
		opt.ListenAddr, opt.ListenPort,
		opt.ConnectAddr, opt.ConnectPort)
	log.Printf("Authenticating clients with password '%s'.",
		opt.Password)

	var (
		listener net.Listener
		err      error
	)

	// Bind to the listening socket.
	listener, err = net.Listen("tcp", strings.Join([]string{opt.ListenAddr, opt.ListenPort}, ":"))
	if err != nil {
		log.Fatalf("Error binding to socket: %s", err)
	}

	// For and accept connections.
	for {
		var (
			conn net.Conn
			err  error
		)

		// Block until we get a connection.
		conn, err = listener.Accept()
		if err != nil {
			log.Printf("Error accepting client connection: %s", err)
		}

		// Hand the connection off to a goroutine.
		go func(conn net.Conn) {
			var (
				clientSock net.Conn
				client     *Connection
				remoteSock net.Conn
				remote     *Connection
			)

			// Pick up the accepted connection.
			clientSock = conn

			// Create a new connection object for the client.
			client = NewConnection(clientSock)
			// Always close the client socket.
			defer func(client *Connection) {
				log.Printf("Closed client connection from %s.", client.Address)
				client.Close()
			}(client)

			log.Printf("Accepted client connection from %s.", client.Address)

			// Attempt to authenticate the client.
			if !authConnection(client) {
				log.Printf("Got bad authentication from %s.", client.Address)
				return
			}

			log.Printf("Got good authentication from %s.", client.Address)

			// Spawn the remote connection.
			remoteSock, err = net.Dial("tcp", strings.Join([]string{opt.ConnectAddr, opt.ConnectPort}, ":"))
			if err != nil {
				log.Printf("Error opening connection: %s", err)
				return
			}

			// Create a new connection object for the remote.
			remote = NewConnection(remoteSock)
			// Always close the remote socket.
			defer func(remote *Connection) {
				log.Printf("Closed remote connection to %s.", remote.Address)
				remote.Close()
			}(remote)

			log.Printf("Opened remote connection to %s.", remote.Address)

			// Link the client to the server.
			go pipe(remote.Incoming, client.Outgoing)
			go pipe(client.Incoming, remote.Outgoing)

		For:
			// Block until the connection completes.
			for {
				select {
				case <-client.Complete:
					log.Print("Client hung up, tearing down.")
					break For
				case <-remote.Complete:
					log.Print("Remote hung up, tearing down.")
					break For
				}
			}
		}(conn)
	}
}

func main() {
	app := cli.NewApp()

	app.Name = Name
	app.Version = Version
	app.Usage = Description
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "l, laddr", Value: "0.0.0.0", Usage: "Address to listen on."},
		cli.StringFlag{Name: "L, lport", Value: "1337", Usage: "Port to listen on."},
		cli.StringFlag{Name: "c, caddr", Value: "127.0.0.1", Usage: "Address to connect to."},
		cli.StringFlag{Name: "C, cport", Value: "6667", Usage: "Port to connect to."},
		cli.StringFlag{Name: "p, pass", Value: "opensesame", Usage: "Password to authenticate against."},
	}
	app.Action = func(c *cli.Context) {
		// Parse options.
		opt.ListenAddr = c.String("laddr")
		opt.ListenPort = c.String("lport")
		opt.ConnectAddr = c.String("caddr")
		opt.ConnectPort = c.String("cport")
		opt.Password = c.String("pass")

		// Call real main().
		_main()
	}

	app.Run(os.Args)
}

// vi: ts=4
