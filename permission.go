package juno

import (
	"strconv"
	"strings"
)

//The Permission interface is to be implemented in a way that exports an identifier via the ID method, as well as defines equality with another permission via implementing the Match Method
type Permission interface {
	ID() string
	Equals(Permission) bool
}

//Permissions is a map of string keys to Permission values
type Permissions map[string]Permission

//NewStdPermission is a factory constructor for returning a standard implementation of the Permission Interface
func NewStdPermission(label, description string) *StdPermission {
	return &StdPermission{
		Label:       label,
		Description: description,
	}
}

//StdPermission is the Juno implementation of the permission interface
type StdPermission struct {
	//The PermissionID as it is in the database
	PermissionID int    `json:"id" db:"PermissionID"`
	Label        string `json:"label" db:"Label"`
	Description  string `json:"description" db:"Description"`
}

//ID implements the Permission interface, exposing the value intended to represent the StdPermissions ID
func (p *StdPermission) ID() string {
	return strconv.Itoa(p.PermissionID)
}

//Equals is the implementation of the Permission interface Equal method
func (p *StdPermission) Equals(perm Permission) bool {
	if _, ok := perm.(*StdPermission); ok {
		return strings.ToLower(strconv.Itoa(p.PermissionID)) == strings.ToLower(perm.ID())
	}
	return false
}
