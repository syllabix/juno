package mssqlrepo

import (
	"database/sql"

	"errors"

	"github.com/syllabix/juno"
)

//NewUserAuthenticationRepo constructor
func NewUserAuthenticationRepo(db *sql.DB) *UserAuthenticationRepo {
	return &UserAuthenticationRepo{
		db: db,
	}
}

//UserAuthenticationRepo is the mssql implementation of the juno.UserAuthRepo
type UserAuthenticationRepo struct {
	db *sql.DB
}

const selectbyusername = `
    SELECT UserID, Email, Password, UserRoles.RoleID, UserRoles.RoleName, Users.Created, Users.Modified, Users.LastLogin 
    FROM Users
    JOIN UserRoles ON Users.RoleID = UserRoles.RoleID
    WHERE Users.Email = ?`

//GetUserByCredentials returns a juno.User for the the provided juno.Credentials
func (repo *UserAuthenticationRepo) GetUserByCredentials(creds juno.Credentials) (juno.User, error) {
	email := creds.GetUsername()
	user := juno.StdUser{}
	err := repo.db.QueryRow(selectbyusername, email).Scan(&user.UserID, &user.Email, &user.Password, &user.RoleID, &user.RoleName, &user.Created, &user.Modified, &user.LastLogin)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

const selectbyid = `
    SELECT UserID, Email, UserRoles.RoleID, UserRoles.RoleName 
    FROM Users
    JOIN UserRoles ON Users.RoleID = UserRoles.RoleID
    WHERE Users.UserID = ?`

//GetUserFromSession returns a juno.User from a provided user.Session
func (repo *UserAuthenticationRepo) GetUserFromSession(s juno.Session) (juno.User, error) {
	id, ok := s.Get(juno.USER_ID_SESSION_KEY)
	if !ok {
		return nil, errors.New("Session is not authenticated")
	}
	user := new(juno.StdUser)
	err := repo.db.QueryRow(selectbyid, id).Scan(&user.UserID, &user.Email, &user.RoleID, &user.RoleName)
	if err != nil {
		return nil, err
	}
	return user, nil
}
