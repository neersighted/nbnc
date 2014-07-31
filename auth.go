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
	case <-time.After(time.Second * AuthTimeout):
		log.Printf("Authentication timed out, tearing down.")
		return false
	}
}

func checkAuth(client <-chan string, result chan<- bool) {
	var (
		attempt int
		data    string

		pass *regexp.Regexp
	)

	pass = regexp.MustCompile(`^(?i)PASS(?-i) ` + opt.Password + `\r?\n?$`)

	attempt = 0
	for data = range client {
		if attempt > AuthAttempts {
			log.Print("Authentication bad, tearing down.")
			result <- false
			return
		}

		if pass.MatchString(data) {
			log.Print("Authentication good, open sesame.")
			result <- true
			return
		}

		attempt++
	}
}

// vi: ts=4
