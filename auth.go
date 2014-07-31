package main

import (
	"log"
	"regexp"
	"time"
)

func authConnection(conn *Connection) bool {
	var (
		res  bool
		auth chan bool
	)

	// Create the channel to get our result back from.
	auth = make(chan bool, 1)

	// Run the authentication task as a goroutine so we can have a timeout.
	go checkAuth(conn, auth)

	select {
	case res = <-auth:
		return res
	case <-time.After(time.Second * 10):
		log.Printf("Timed out when authenticating %s, tearing down.", conn.Address)
		return false
	}
}

func checkAuth(conn *Connection, result chan<- bool) {
	var (
		attempt int
		data    string

		pass *regexp.Regexp
	)

	pass = regexp.MustCompile(`^PASS ` + opt.Password + `\r?\n?$`)

	attempt = 0
	for data = range conn.Incoming {
		if pass.MatchString(data) {
			log.Printf("Got good authentication from %s, open sesame.", conn.Address)
			result <- true
			return
		}

		if attempt > 1 {
			log.Printf("Got bad authentication from %s, tearing down.", conn.Address)
			result <- false
			return
		}

		attempt++
	}
}

// vi: ts=4
