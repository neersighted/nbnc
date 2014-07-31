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
	go checkAuth(conn.Incoming, auth)

	select {
	case res = <-auth:
		return res
	case <-time.After(time.Second * 10):
		log.Printf("Authentication for %s timed out.", conn.Address)
		return false
	}
}

func checkAuth(client <-chan string, result chan<- bool) {
	var (
		attempt int
		data    string

		pass *regexp.Regexp
	)

	pass = regexp.MustCompile(`^PASS ` + opt.Password + `\r?\n?$`)

	attempt = 0
	for data = range client {
		if pass.MatchString(data) {
			result <- true
			return
		}

		if attempt > 1 {
			result <- false
			return
		}

		attempt++
	}
}

// vi: ts=4
