package juno

import "golang.org/x/crypto/bcrypt"

import "errors"

type (

	//UserAuthRepo is the interface that is intended to be implemented by a data access struct methods
	UserAuthRepo interface {
		GetUserByCredentials(Credentials) (User, error)
		GetUserFromSession(Session) (User, error)
	}

	//The Credentials interface exposes getters for password and username
	Credentials interface {
		GetUsername() string
		GetPassword() string
	}
)

var (
	//ErrInvalidCredentials to be returned for invalid credentials
	ErrInvalidCredentials = errors.New("The provided credentials are not valid.")
)

//NewAuthenticator returns an pointer to an authenticar, taking an implemented UserRepo as it only argument
func NewAuthenticator(repo UserAuthRepo) *Authenticator {
	return &Authenticator{
		repo: repo,
	}
}

//The Authenticator is used to login in users, encrypt passwords, and validate users are authenticated
type Authenticator struct {
	repo UserAuthRepo
}

//EncryptPassword uses bcrypt to encrypt a provided password in a way that ensures decryption using respective Authenticate method works as expected
func (a *Authenticator) EncryptPassword(password string) (string, error) {
	encPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	return string(encPass), nil
}

//Authenticate takes the provided credentials and authenticates the a user, returning the full user on success, error on failure
func (a *Authenticator) Authenticate(creds Credentials) (User, error) {
	user, err := a.repo.GetUserByCredentials(creds)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.GetPassword()), []byte(creds.GetPassword()))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

//IsAuthenticatedSession takes an a current session, and return the user if the session is authenticated, otherwise return an error
func (a *Authenticator) IsAuthenticatedSession(s Session) (User, error) {
	return a.repo.GetUserFromSession(s)
}
