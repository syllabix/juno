package juno

import (
	"testing"

	"time"

	"log"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewStdSession(t *testing.T) {
	assert := assert.New(t)

	session := NewStdSession()

	_, err := uuid.FromString(session.SessionID())
	assert.NoError(err, "juno.StdSession should have an id that is a valid uuid")
	assert.WithinDuration(time.Now(), session.Expiration, time.Minute*30, "When passed no arguments, a new session should default to a length of 30 minutes")
	assert.NotNil(session.store, "A session instantiated with the NewStdSession constructor should have a non nil store")

	timeTestSession := NewStdSession(time.Second * 2)
	log.Println("Pausing execution for 3 seconds to verify a time based test")
	time.Sleep(time.Second * 3)
	assert.True(timeTestSession.Expired(), "Sessions should properly expire after their set duration has elapsed")
}

func TestSessionGetSetDelete(t *testing.T) {
	assert := assert.New(t)

	session := NewStdSession()
	testID := 120
	testRoleName := "admin"
	session.Set("userid", testID)
	session.Set("role", testRoleName)
	assert.Equal(2, len(session.store))
	userID, found := session.Get("userid")
	assert.True(found, "A valid call to Get in a session store, when used on existing set value, should return found.")
	assert.Equal(testID, userID, "A call to get should return the last set value against the key in a session")

	session.Delete("role")
	assert.Equal(1, len(session.store), "A delete against a valid key in a session store should remove the record properly")

	nilStoreSession := new(StdSession)
	nilStoreSession.Set("test", "foo")
	assert.Equal(1, len(nilStoreSession.store), "Session not instantiated via the factory constructor should still properly instantiate the underlying map and store values")

}
