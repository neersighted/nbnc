package main

import (
	"bufio"
	"log"
	"net"
	"regexp"
	"time"
)

func mux(client net.Conn, config *Config) (bouncer *BouncerConfig) {
	// Attempt to handshake with the client and determine the requested config.
	var match chan *BouncerConfig = make(chan *BouncerConfig, 1)
	go handshake(client, config, match)

	// Block until handshake is done or we hit our timeout.
	select {
	case bouncer = <-match:
		return bouncer
	case <-time.After(time.Second * config.Auth.Timeout):
		log.Printf("TIMEOUT %s", client.RemoteAddr())
		return nil
	}
}

var shake *regexp.Regexp = regexp.MustCompile(`^(?i)PASS(?-i) (\w+):?(.+)?\r?\n?$`)

func handshake(client net.Conn, config *Config, match chan *BouncerConfig) {
	var attempt int

	// Read lines from the client and try to match them against our handshake.
	var scanner *bufio.Scanner = bufio.NewScanner(client)
	for scanner.Scan() {
		if shake.MatchString(scanner.Text()) {
			var matches [][]byte = shake.FindSubmatch(scanner.Bytes())

			var bouncer BouncerConfig
			for i := range config.Bouncer {
				if string(matches[1]) == i {
					bouncer = config.Bouncer[i]
					if string(matches[2]) == bouncer.Password {
						log.Printf("LOGIN %s %s", client.RemoteAddr(), i)
						match <- &bouncer
						return
					}
				} else {
					continue
				}
			}
		}

		if attempt > config.Auth.Attempts {
			log.Printf("REJECT %s", client.RemoteAddr())
			match <- nil
			return
		}

		attempt++
	}
}

// vi: ts=4
