package juno

import "time"

type (

	//User interface is to be implemented in a way that provided access to the necessary values related to a user for utilization throughout the juno Authenticator and Authorizer workflows
	User interface {
		Credentials
		ID() int
		Role() UserRole
	}
)

//StdUser implements the user interface and is meant to be thought of as a base user struct, and embedded in more involved and/or specfic user structs
type StdUser struct {
	UserID      int    `db:"UserID" json:"userId"`
	Email       string `db:"Email" json:"email"`
	Password    string `db:"Password" json:"password,omitempty"`
	StdUserRole `json:"role"`
	Created     time.Time              `db:"Created" json:"created"`
	Modified    time.Time              `db:"Modified" json:"modified"`
	LastLogin   time.Time `db:"LastLogin" json:"lastLogin"`
}

//GetUsername implements the Credentials interface and returns the users email
func (u *StdUser) GetUsername() string {
	return u.Email
}

//GetPassword implements the Credentials interface and returns the users password
func (u *StdUser) GetPassword() string {
	return u.Password
}

//ID implements the User interface and exposes the users id
func (u *StdUser) ID() int {
	return u.UserID
}

//Role implements the User interface and exposes the user's role
func (u *StdUser) Role() UserRole {
	return &u.StdUserRole
}
