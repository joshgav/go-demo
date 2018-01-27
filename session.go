package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	model "github.com/joshgav/go-demo/model"
	"github.com/subosito/gotenv"
)

const (
	envvarName  = "COOKIE_KEY"
	sessionName = "vanpool_user"

	selfKey          = "self"
	authenticatedKey = "authenticated"
	stateKey         = "state"
)

var store sessions.Store

func init() {
	gotenv.Load()
	key := os.Getenv(envvarName)
	if len(key) == 0 {
		log.Printf("Session (init): use envvar %v for cookie key\n", envvarName)
		key = "makemerandom"
	}

	log.Printf("Session (init): registering Rider type\n")
	gob.Register(&model.Rider{})
	log.Printf("Session (init): creating new cookie store\n")
	store = sessions.NewCookieStore([]byte(key))
}

// Session is middleware which creates/restores session info
// session vars: state string, self *model.Rider, authenticated bool
// use in a later handler: // TODO: helper methods
//   `rider, ok := r.Context().Value(selfKey).(*model.Rider)`
//   `authenticated, ok := r.Context().Value(authenticatedKey).(bool)`
//   `state, ok := r.Context().Value(stateKey).(string)`
func Session(next http.Handler) http.Handler {
	log.Printf("Session: hello")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("SetSession: getting session %v\n", sessionName)
		s, err := store.Get(r, sessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if _, ok := s.Values[stateKey].(string); ok == false {
			log.Printf("Session: no state currently in session, adding it\n")
			// getting state creates it and adds to state if necessary
			state, err := GetState(w, r)
			if err != nil {
				http.Error(w, "failed to get state", http.StatusInternalServerError)
				log.Printf("Session: failed to get/create state: %v\n")
			}
			log.Printf("Session: state set to: %v\n", state)
		}

		if _, ok := s.Values[authenticatedKey].(bool); ok == false {
			log.Printf("Session: user not previously authenticated\n")
			log.Printf("Session: marking user as not authenticated\n")
			s.Values[authenticatedKey] = false
		}

		if _, ok := s.Values[selfKey].(*model.Rider); ok == false {
			log.Printf("Session: user not available from session\n")
			s.Values[selfKey] = &model.Rider{}
			s.Values[authenticatedKey] = false
		} else {
			log.Printf("Session: found user in session, marking as authenticated\n")
			s.Values[authenticatedKey] = true
		}

		log.Printf("Session: saving session (%v)\n", s)
		s.Save(r, w)

		log.Printf("Session: adding session data to context for ensuing modules\n")
		var ctx context.Context
		ctx = context.WithValue(r.Context(), selfKey, s.Values[selfKey])
		ctx = context.WithValue(ctx, authenticatedKey, s.Values[authenticatedKey])
		ctx = context.WithValue(ctx, stateKey, s.Values[stateKey])

		log.Printf("Session: done, calling next with context (%v)\n", ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetSession(token string, w http.ResponseWriter, r *http.Request) error {
	log.Printf("SetSession: getting session %v\n", sessionName)
	s, err := store.Get(r, sessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	t, err := jwt.Parse(token, nil)
	if err != nil {
		log.Printf("SetSession: failed to parse JWT: %v\n", err)
	}
	log.Printf("SetSession: parsed jwt: %+v\n", t)
	s.Values[selfKey] = &model.Rider{
	// TODO: populate based on info from JWT
	}
	s.Save(r, w)
	return err
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("UserHandler: checking for session user\n")
	rider, ok := r.Context().Value(selfKey).(*model.Rider)
	if ok == false {
		log.Printf("UserHandler: no session user found\n")
		// send an error object
	}
	log.Printf("UserHandler: responding with session user: %+v\n", rider)
	json, err := json.Marshal(rider)
	if err != nil {
		log.Printf("UserHandler: failed to marshal json: %v\n", err)
		// send an error object
	}
	w.Write(json)
}

func GetState(w http.ResponseWriter, r *http.Request) (string, error) {
	// log.Printf("GetState: getting session %v\n", sessionName)
	// s, err := store.Get(r, sessionName)
	// if err != nil {
	//   http.Error(w, err.Error(), http.StatusInternalServerError)
	// }

	// build new state if necessary and save to request session

	return "makemerandom", nil
}
