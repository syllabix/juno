package juno

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"reflect"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
)

var blockKey = securecookie.GenerateRandomKey(32)
var hashKey = securecookie.GenerateRandomKey(32)

func TestNewStdCookieProvider(t *testing.T) {
	assert := assert.New(t)
	name := "test-cookie"
	cookieProvider := NewStdCookieProvider(hashKey, blockKey, name)
	assert.Equal(name, cookieProvider.name, "The NewStdCookieProvider factory constructor sets a the cookie name properly")
	cookieType := reflect.TypeOf(cookieProvider.secure).String()
	assert.Equal(cookieType, "*securecookie.SecureCookie", "Cookie provider should use an implementation of the secure cookie package for encoding and decoding cookie values")
}

func TestCookieProviderSetRead(t *testing.T) {
	assert := assert.New(t)

	cookieName := "test-cookie"
	cookieProvider := NewStdCookieProvider(hashKey, blockKey, cookieName)
	recorder := httptest.NewRecorder()
	session := NewStdSession()
	cookieProvider.Set(recorder, session)
	header := recorder.HeaderMap["Set-Cookie"]

	request := &http.Request{Header: http.Header{"Cookie": recorder.HeaderMap["Set-Cookie"]}}

	//Tests for cookie provider Set
	assert.Equal(1, len(header), "Cookie provider Set method should properly set a cookie on the http response")
	cookie, err := request.Cookie(cookieName)
	assert.NoError(err, "Cookie should contain encrypted cookie set by the provider")
	value := make(map[string]string)
	err = cookieProvider.secure.Decode(cookieName, cookie.Value, &value)
	assert.NoError(err, "Cookie provider should properly decrypt cookie value")
	assert.Equal(session.SessionID(), value[sessionID])

	//Tests for cookie provider Read
	readSession, err := cookieProvider.Read(request)
	assert.Equal(session.SessionID(), readSession.SessionID(), "Session read from the request should match the session id from the intial set session")
}
