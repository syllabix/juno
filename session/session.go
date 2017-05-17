package session

import (
	"context"

	"github.com/syllabix/juno"
)

type key int

const sessionKey key = 000

//NewContext returns a new context with a session id
func NewContext(ctx context.Context, session juno.Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

//FromContext takes a context as an argument and extracts the sessionID from it if set
func FromContext(ctx context.Context) (juno.Session, bool) {
	session, ok := ctx.Value(sessionKey).(juno.Session)
	return session, ok
}
