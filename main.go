package main

import (
	"log"
	"os"
)

func main() {
	var err error

	// Log to stderr until our new logger is setup.
	log.SetOutput(os.Stderr)

	// Check we have the correct number of arguments.
	if len(os.Args) < 3 {
		log.Fatalf("%s <config> <log>", os.Args[0])
	}

	// Load the config from the first argument.
	var config *Config
	config, err = loadConfig(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Open the logfile from the second argument.
	var logfile *os.File
	if os.Args[2] == "-" {
		logfile = os.Stdout
	} else {
		logfile, err = os.OpenFile(os.Args[2], os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Redirect logging to the new logfile.
	log.SetOutput(logfile)
	// Dump config if in debug mode.
	if config.Debug {
		log.Println(config)
	}

	// Start listening...
	bouncer(config)
}
