package logutil

import "log"

func Logger(_ int, err error) {
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
}
