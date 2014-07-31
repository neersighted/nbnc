package main

func pipe(in, out chan string) {
	var data string

	for {
		select {
		case data = <-in:
			out <- data
		}
	}
}

// vi: ts=4
