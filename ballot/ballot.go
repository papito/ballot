package main

import (
	"ballot/ballot/hub"
	"github.com/desertbit/glue"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	/* Connect to Redis
	 */
	var err error
	redisConn, err = redis.DialURL(redisUrl)
	if err != nil {
		log.Fatal("Error getting index view ", err)
	}

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

	http.Handle("/glue/ws", glueSrv)

	/* Serve static files
	 */
	fs := http.FileServer(http.Dir("ui/dist/js"))
	http.Handle("/ui/js/",http.StripPrefix("/ui/js/", fs))

	/* Handlers
	 */
	r := mux.NewRouter()
	r.HandleFunc("/", indexHttpHandler).Methods("GET")
	r.HandleFunc("/api/session", createSessionHttpHandler).Methods("POST")
	r.HandleFunc("/api/user", createUserHttpHandler).Methods("POST")
	r.HandleFunc("/api/vote/start", startVoteHttpHandler).Methods("PUT")
	r.HandleFunc("/api/vote/cast", castVoteHttpHandler).Methods("PUT")
	http.Handle("/", r)

	log.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}