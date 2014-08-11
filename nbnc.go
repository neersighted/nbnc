package main

import (
	"crypto/tls"
	"fmt"
	"github.com/codegangsta/cli"
	"log"
	"net"
	"os"
	"time"
)

const (
	Name        string = "nbnc"
	Version     string = "0.2.3"
	Description string = "simple null (transparent) bnc"

	AuthTimeout  time.Duration = 15
	AuthAttempts int           = 2
)

var (
	opt Options
)

type Options struct {
	ListenAddr string
	ListenPort int
	RemoteAddr string
	RemotePort int
	OutAddr    string

	RemoteSSL    bool
	RemoteVerify bool

	ForceV4 bool
	ForceV6 bool

	Password string

	Log bool
}

func _main() {
	var (
		proto string

		lsrv  net.Listener
		laddr *net.TCPAddr

		err error
	)

	// Print program information.
	log.Printf("%s %s", Name, Version)
	// Also print exactly what we're doing.
	log.Printf("Proxying %s:%d -> %s:%d, using password '%s'.",
		opt.ListenAddr, opt.ListenPort,
		opt.RemoteAddr, opt.RemotePort,
		opt.Password)

	// Let us know if we're using a specific address.
	if len(opt.OutAddr) != 0 {
		log.Printf("Connecting using %s only.", opt.OutAddr)
	}

	// Figure out which ip version we're using.
	if opt.ForceV4 && !opt.ForceV6 {
		log.Print("Connecting using IPv4 only.")
		proto = "tcp4"
	} else if opt.ForceV6 && !opt.ForceV4 {
		log.Print("Connecting using IPv6 only.")
		proto = "tcp6"
	} else {
		proto = "tcp"
	}

	// Resolve the listen address.
	laddr, err = net.ResolveTCPAddr(proto, fmt.Sprintf("%s:%d", opt.ListenAddr, opt.ListenPort))
	if err != nil {
		log.Fatalf("Error resolving listen address: %s", err)
	}

	// Bind to the listening socket.
	lsrv, err = net.ListenTCP("tcp", laddr)
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
		conn, err = lsrv.Accept()
		if err != nil {
			log.Printf("Error accepting client connection: %s", err)
		}

		// Hand the connection off to a goroutine.
		go func(conn net.Conn) {
			var (
				raddr *net.TCPAddr
				oaddr *net.TCPAddr

				csock net.Conn
				rsock net.Conn

				cconn *Connection
				rconn *Connection

				tconf tls.Config
			)

			// Pick up the accepted connection.
			csock = conn

			// Create a new connection object for the client.
			cconn = NewConnection(csock)
			// Always close the client socket.
			defer func(cconn *Connection) {
				log.Printf("Closed client connection from %s.", cconn.Address)
				cconn.Close()
			}(cconn)

			log.Printf("Accepted client connection from %s.", cconn.Address)

			// Attempt to authenticate the client.
			if !authConnection(cconn) {
				return
			}

			// Resolve the outgoing address.
			oaddr, err = net.ResolveTCPAddr(proto, fmt.Sprintf("%s:", opt.OutAddr))
			if err != nil {
				log.Fatalf("Error resolving connection address: %s", err)
			}
			// Resolve the remote address.
			raddr, err = net.ResolveTCPAddr(proto, fmt.Sprintf("%s:%d", opt.RemoteAddr, opt.RemotePort))
			if err != nil {
				log.Fatalf("Error resolving remote address: %s", err)
			}

			// Open the remote connection.
			rsock, err = net.DialTCP(proto, oaddr, raddr)
			if err != nil {
				log.Printf("Error opening connection: %s", err)
				return
			}

			if opt.RemoteSSL {
				tconf = tls.Config{ServerName: opt.RemoteAddr, InsecureSkipVerify: opt.RemoteVerify}
				rsock = tls.Client(rsock, &tconf)
			}

			// Create a new connection object for the remote.
			rconn = NewConnection(rsock)
			// Always close the remote socket.
			defer func(rconn *Connection) {
				log.Printf("Closed remote connection to %s.", rconn.Address)
				rconn.Close()
			}(rconn)

			log.Printf("Opened remote connection to %s.", rconn.Address)

			// Link the client to the server.
			go pipe(rconn.Incoming, cconn.Outgoing)
			go pipe(cconn.Incoming, rconn.Outgoing)

		For:
			// Block until the connection completes.
			for {
				select {
				case <-cconn.Complete:
					log.Print("Client hung up, tearing down.")
					break For
				case <-rconn.Complete:
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
		cli.StringFlag{Name: "l, laddr", Usage: "Local address to listen on.", Value: "0.0.0.0"},
		cli.IntFlag{Name: "L, lport", Usage: "Local port to listen on.", Value: 1337},
		cli.StringFlag{Name: "r, raddr", Usage: "Remote address to connect to.", Value: "127.0.0.1"},
		cli.IntFlag{Name: "R, rport", Usage: "Remote port to connect to.", Value: 6667},
		cli.StringFlag{Name: "o, oaddr", Usage: "Outgoing address to connect with."},
		cli.BoolFlag{Name: "s, ssl", Usage: "Connect with SSL."},
		cli.BoolFlag{Name: "S, verify", Usage: "Don't verify SSL certificates."},
		cli.BoolFlag{Name: "4", Usage: "Force connection to use IPv4."},
		cli.BoolFlag{Name: "6", Usage: "Force connection to use IPv6."},
		cli.StringFlag{Name: "p, pass", Usage: "Password to authenticate against.", Value: "opensesame"},
		cli.BoolFlag{Name: "V, log", Usage: "Log IRC traffic."},
	}
	app.Action = func(c *cli.Context) {
		// Parse options.
		opt.ListenAddr = c.String("laddr")
		opt.ListenPort = c.Int("lport")
		opt.RemoteAddr = c.String("raddr")
		opt.RemotePort = c.Int("rport")
		opt.OutAddr = c.String("oaddr")

		opt.RemoteSSL = c.Bool("ssl")
		opt.RemoteVerify = c.Bool("verify")

		opt.ForceV4 = c.Bool("4")
		opt.ForceV6 = c.Bool("6")

		opt.Password = c.String("pass")

		opt.Log = c.Bool("log")

		// Call real main().
		_main()
	}

	app.Run(os.Args)
}

// vi: ts=4
