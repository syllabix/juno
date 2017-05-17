package userrole

import (
	"context"

	"github.com/syllabix/juno"
)

type key string

const userkey key = "user_role"

//NewContext returns a new context with a session id
func NewContext(ctx context.Context, role juno.UserRole) context.Context {
	return context.WithValue(ctx, userkey, role)
}

//FromContext takes a context as an argument and extracts the sessionID from it if set
func FromContext(ctx context.Context) (juno.UserRole, bool) {
	session, ok := ctx.Value(userkey).(juno.UserRole)
	return session, ok
}
