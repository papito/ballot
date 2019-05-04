package main

import (
	"ballot/ballot/hub"
	"github.com/desertbit/glue"
	"log"
)

// FIXME: env  var
const redisUrl = "redis://localhost:6379"

func main() {
	var err error
	/* Initiate the hub that connects sessions and sockets
	 */
	log.Println("Creating hub")
	err = hub.InitHub(redisUrl)
	if err != nil {
		log.Fatal(err)
	}

	/* Create the Glue server
	 */
	glueSrv := glue.NewServer(glue.Options{
		HTTPSocketType: glue.HTTPSocketTypeNone,
	})
	defer glueSrv.Release()

	glueSrv.OnNewSocket(hub.HandleSocket)

	NewServer(glueSrv)
}