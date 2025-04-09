package router

import (
	"application/api"
	"application/endpoints/verve"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
)

func NewServer(c chan os.Signal, rdb *redis.Client) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}
	return &http.Server{Addr: "localhost:" + port, Handler: newHandler(c, rdb)}
}

func addRoutes(r *httprouter.Router, c chan os.Signal, rdb *redis.Client) {
	requestCounter := verve.NewRequestCounter(rdb)
	ticker := time.NewTicker(time.Minute)

	go func() {
		for {
			select {
			case <-c:
				return
			case t := <-ticker.C:
				requestCounter.Reset(t)
				fmt.Println("Tick at", t)
			}
		}
	}()

	verveHandler := verve.NewHandler(verve.NewService(), requestCounter)

	r.GET("/api/verve/accept/:id", verveHandler.VerveAccept)
}

func newHandler(c chan os.Signal, rdb *redis.Client) http.Handler {
	r := httprouter.New()
	addRoutes(r, c, rdb)

	r.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			header := w.Header()
			header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			header.Set("Access-Control-Allow-Methods", header.Get("Allow"))
			header.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		log.Printf("panic: %+v", err)
		api.Error(w, r, fmt.Errorf("whoops! My handler has run into a panic"), http.StatusInternalServerError)
	}
	r.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Error(w, r, fmt.Errorf("we have OPTIONS for youm but %v is not among them", r.Method), http.StatusMethodNotAllowed)
	})
	r.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Error(w, r, fmt.Errorf("whatever route you've been looking for, it's not here"), http.StatusNotFound)
	})

	return r
}
