package verve

import (
	"application/api"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bsm/redislock"
	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
)

type RequestCounter struct {
	redis  *redis.Client
	locker *redislock.Client
}

func NewRequestCounter(rdb *redis.Client) *RequestCounter {
	return &RequestCounter{
		redis:  rdb,
		locker: redislock.New(rdb),
	}
}

func (rc *RequestCounter) GetUniqReqCount() int {
	ctx := context.Background()
	count := 0
	iter := rc.redis.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		if iter.Val() != "my-key" {
			count++
			log.Println("key ", iter.Val())
		}
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}
	return count
}

func (rc *RequestCounter) Reset(t time.Time) {
	ctx := context.Background()
	// Try to obtain lock.
	lock, err := rc.locker.Obtain(ctx, "my-key", 5000*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock!")
	} else if err != nil {
		log.Fatalln(err)
	}

	// Don't forget to defer Release.
	defer lock.Release(ctx)
	fmt.Println("I have a lock!")

	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("At ", t)

	iter := rc.redis.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		if iter.Val() != "my-key" {
			log.Println("key ", iter.Val())
		}
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}

	// rc.c = make(map[string]*http.Cookie)
	err = rc.redis.FlushDB(context.Background()).Err()
	if err != nil {
		log.Fatalf("error on resetting db %v", err)
	}
}

func (rc *RequestCounter) CheckCookie(key string) bool {
	result, err := rc.redis.Get(context.Background(), key).Result()
	if err != nil {
		fmt.Println("Key not found in Redis cache")
		return false
	}
	fmt.Printf("key has value %s\n", result)
	return true
}

func (rc *RequestCounter) UpdateCounter(cookie *http.Cookie) {
	_, err := rc.redis.Set(context.Background(), cookie.Value, true, 0).Result()
	if err != nil {
		fmt.Println("Failed to add key-value pair", err)
		return
	}
}

func (rc *RequestCounter) HandleCookie(w http.ResponseWriter, r *http.Request, id string) {

	ctx := context.Background()
	// Try to obtain lock.
	lock, err := rc.locker.Obtain(ctx, "my-key", 5000*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock!")
	} else if err != nil {
		log.Fatalln(err)
	}

	// Don't forget to defer Release.
	defer lock.Release(ctx)
	fmt.Println("I have a lock!")

	var newCookie *http.Cookie
	cookie, err := r.Cookie("exampleCookie")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			log.Println("no cookie err", err)
			newCookie = &http.Cookie{
				Name:     "exampleCookie",
				Value:    id,
				Path:     "/api/verve/accept",
				MaxAge:   60,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			rc.UpdateCounter(newCookie)
			http.SetCookie(w, newCookie)
		default:
			log.Println("server error", err)
			api.Error(w, r, errors.New("server error"), http.StatusBadRequest)
		}
	}

	if cookie != nil {
		if !rc.CheckCookie(id) {
			newCookie = &http.Cookie{
				Name:     "exampleCookie",
				Value:    id,
				Path:     "/api/verve/accept",
				MaxAge:   60,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			rc.UpdateCounter(newCookie)
			http.SetCookie(w, newCookie)
		} else if rc.CheckCookie(id) {
			newCookie = &http.Cookie{
				Name:     "exampleCookie",
				Value:    id,
				Path:     "/api/verve/accept",
				MaxAge:   60,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, newCookie)
		} else {
			http.SetCookie(w, cookie)
		}
	}
}

type Handler struct {
	service Service
	rc      *RequestCounter
}

func NewHandler(service Service, rc *RequestCounter) *Handler {
	return &Handler{service: service, rc: rc}
}

func (h *Handler) VerveAccept(w http.ResponseWriter, r *http.Request, urlParams httprouter.Params) {
	id := urlParams.ByName("id")
	err := validateId(id)
	fmt.Println("err", err)
	if err != nil {
		api.Error(w, r, err, http.StatusBadRequest)
		return
	}

	h.rc.HandleCookie(w, r, id)

	endpoint := r.URL.Query().Get("endpoint")
	result, err := h.service.Accept(id, endpoint, h.rc.GetUniqReqCount())
	if err != nil {
		api.Error(w, r, err, 0)
		return
	}
	api.SuccessJson(w, r, result)
}
