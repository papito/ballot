package main

import "log"

func logerr(_ int, err error) {
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
}
