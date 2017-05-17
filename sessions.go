package juno

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

type (
	//The SessionProvider is to be implemented by the persistance mechanism for sessions and injected into the session manager
	SessionProvider interface {
		GetSession(*http.Request) (Session, error)
		SetSession(Session) error
		EndSession(http.ResponseWriter, Session) error
		UpdateSession(Session) error
		WriteCookie(http.ResponseWriter, Session) error
	}

	CookieProvider interface {
		//Read returns a session, or error if empty
		Read(*http.Request) (Session, error)
		//Write takes a response writer and sets a cookie to it
		Set(http.ResponseWriter, Session) error
		//Invalidate a the session cookie
		Invalidate(http.ResponseWriter)
	}

	//Session is an interface to be implemented by a general session object - provides access to the id, expiration state, and general setter, getters, and delete against session values
	Session interface {
		SessionID() string
		Set(key string, value interface{})
		Get(key string) (interface{}, bool)
		Delete(key string)
		Expired() bool
		Store() map[string]interface{}
		StoreDirty() bool
		ReplaceStore(map[string]interface{})
	}
)

var (
	ErrNoSessionID      = errors.New("Cookie does not have valid session id")
	ErrInvalidSessionID = errors.New("Session ID present is not valid")
	ErrSessionExpired   = errors.New("Your session has expired.")
)

const USER_ID_SESSION_KEY = "userid"

//NewStdSession is factory constructor for returning a brand new session
func NewStdSession(duration ...time.Duration) *StdSession {
	var exp time.Duration
	if len(duration) < 1 {
		//default to 30 minute session
		exp = time.Minute * 30
	} else {
		exp = duration[0]
	}
	return &StdSession{
		ID:         uuid.NewV4(),
		Expiration: time.Now().Add(exp),
		store:      make(map[string]interface{}),
	}
}

//StdSession is an implementation of the juno.Session interface
type StdSession struct {
	ID         uuid.UUID `db:"GUID"`
	Expiration time.Time `db:"Expiration"`
	store      map[string]interface{}
	storeDirty bool
	sync.RWMutex
}

//SessionID retrieves the session id
func (s *StdSession) SessionID() string {
	return s.ID.String()
}

//Expired returns true if the session is expired, false if it is still valid
func (s *StdSession) Expired() bool {
	return time.Now().After(s.Expiration)
}

//Get a value off the session
func (s *StdSession) Get(key string) (interface{}, bool) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.store[key]
	return val, ok
}

//Set a value on the session
func (s *StdSession) Set(key string, value interface{}) {
	s.Lock()
	defer s.Unlock()
	if s.store == nil {
		s.store = make(map[string]interface{})
	}
	s.store[key] = value
	s.storeDirty = true
}

//Delete a value from the session
func (s *StdSession) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.store, key)
	s.storeDirty = true
}

//Store returns the key/value map of stored contents
func (s *StdSession) Store() map[string]interface{} {
	return s.store
}

//StoreDirty is true when something has been set or deleted on the store but it hasn't been persisted
func (s *StdSession) StoreDirty() bool {
	return s.store != nil && s.storeDirty
}

//ReplaceStore replaces the entire contents of the store. It is not marked dirty since this
//should only be used when loading from a db.
func (s *StdSession) ReplaceStore(store map[string]interface{}) {
	s.Lock()
	defer s.Unlock()
	s.store = store
}
