package main

import (
	"regexp"
)

func authConnection(conn *Connection) bool {
	var (
		data    string
		attempt int

		pass *regexp.Regexp
	)

	pass = regexp.MustCompile(`^PASS ` + opt.Password + `\r?\n?$`)

	attempt = 0
	for data = range conn.Incoming {
		if pass.MatchString(data) {
			return true
		}

		if attempt > 1 {
			break
		}

		attempt++
	}

	return false
}

// vi: ts=4
