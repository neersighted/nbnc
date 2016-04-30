package main

import (
	"bufio"
	"log"
	"net"
	"regexp"
	"time"
)

func handshake(client net.Conn, config *Config) (bouncer *BouncerConfig, remainder []byte) {
	// Setup a buffered reader over the client.
	var reader *bufio.Reader = bufio.NewReader(client)

	// Attempt to handshake with the client and determine the requested config.
	var match chan *BouncerConfig = make(chan *BouncerConfig, 1)
	go seekshake(reader, config, match)

	// Block until handshake is done or we hit our timeout.
	select {
	case bouncer = <-match:
		// Grab the remaining portion of the handshake.
		remainder = make([]byte, reader.Buffered())
		reader.Read(remainder)

		if bouncer != nil {
			log.Printf("LOGIN %s %s", client.RemoteAddr())
		} else {
			log.Printf("REJECT %s", client.RemoteAddr())
		}
		return bouncer, remainder
	case <-time.After(time.Duration(config.Auth.Timeout) * time.Second):
		log.Printf("TIMEOUT %s", client.RemoteAddr())
		return nil, []byte{}
	}
}

var shake *regexp.Regexp = regexp.MustCompile(`^(?i)PASS(?-i) (\w+):?(.+)?\r?\n?$`)

func seekshake(reader *bufio.Reader, config *Config, match chan *BouncerConfig) {
	// Read lines from the client and try to match them against our handshake.
	var attempt int
	for attempt < config.Auth.Attempts {
		data, _, err := reader.ReadLine()
		if err != nil {
			log.Println(err)
		}

		if id := matchshake(data, config); id != "" {
			var bouncer BouncerConfig = config.Bouncer[id]
			match <- &bouncer
			return
		} else {
			attempt++
		}
	}

	match <- nil
}

func matchshake(data []byte, config *Config) string {
	// Attempt to match data against the handshake pattern.
	var matches [][]byte = shake.FindSubmatch(data)

	// Convert matches to strings.
	var (
		name     string = string(matches[1])
		password string = string(matches[2])
	)

	// Check that the config exists and that the password matches.
	if bouncer, ok := config.Bouncer[name]; ok && password == bouncer.Password {
		return name
	}

	return ""
}

// vi: ts=4
