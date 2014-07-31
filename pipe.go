package main

func pipe(in <-chan string, out chan<- string) {
	var data string

	for {
		select {
		case data = <-in:
			out <- data
		}
	}
}

// vi: ts=4
