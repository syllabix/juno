package juno

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/securecookie"
	"github.com/satori/go.uuid"
)

const (
	sessionID = "sessionID"
)

//NewStdCookieProvider is a factory constructor for returning a standard cookie provider
func NewStdCookieProvider(hashKey, blockKey []byte, cookieName string) *StdCookieProvider {
	return &StdCookieProvider{
		secure: securecookie.New(hashKey, blockKey),
		name:   cookieName,
	}
}

//StdCookieProvider implements the juno.Cookie provider interface, and
//is to be used in the context of a juno.SessionProvider to augment
//persistance operations in the provider with cookie handling
type StdCookieProvider struct {
	secure *securecookie.SecureCookie
	name   string
}

func (c *StdCookieProvider) Read(req *http.Request) (Session, error) {
	cookie, err := req.Cookie(c.name)
	if err != nil {
		return nil, err
	}
	value := make(map[string]string)
	err = c.secure.Decode(c.name, cookie.Value, &value)
	if err != nil {
		return nil, err
	}
	id, hasID := value[sessionID]
	if !hasID {
		return nil, ErrNoSessionID
	}
	guid, err := uuid.FromString(id)
	if err != nil {
		return nil, ErrInvalidSessionID
	}

	session := new(StdSession)
	session.ID = guid
	session.Expiration = cookie.Expires
	return session, nil
}

//Set the session id securely on a cookie in the response
func (c *StdCookieProvider) Set(w http.ResponseWriter, s Session) error {
	stdSession, ok := s.(*StdSession)
	if !ok {
		return fmt.Errorf("Expecting juno.StdSession, but got %s", reflect.TypeOf(s))
	}
	value := map[string]string{
		sessionID: s.SessionID(),
	}
	encoded, err := c.secure.Encode(c.name, value)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     c.name,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Expires:  stdSession.Expiration,
	}
	http.SetCookie(w, cookie)
	return nil
}

//Invalidate cookie by setting mage age -1
func (c *StdCookieProvider) Invalidate(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     c.name,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
