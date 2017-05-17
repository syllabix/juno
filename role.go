package juno

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

//UserRole is an interface designed to be implemented by a role to expose its id in the context of a user object
type UserRole interface {
	ID() string
}

// Role is an interface
type Role interface {
	UserRole
	Has(Permission) bool
	Assign(Permission) error
	Revoke(Permission) error
}

// RolePermission is an interface to full fill the exposes role and associated granted permission
type RolePermission interface {
	RoleID() string
	PermissionID() string
}

//Roles is a map of string keys to Role
type Roles map[string]Role

func NewStdRole(name string) *StdRole {
	role := new(StdRole)
	role.RoleName = name
	role.permissions = make(Permissions)
	return role
}

type StdUserRole struct {
	//ID of the StdRole
	RoleID int `json:"id" db:"RoleID"`
	//Name of the StdRole
	RoleName string `json:"name" db:"RoleName"`
}

//ID implements the juno.Role interface for exposing an identifier for the role
func (r *StdUserRole) ID() string {
	return strconv.Itoa(r.RoleID)
}

//StdRole is an Implementation of the Role
type StdRole struct {
	sync.RWMutex

	StdUserRole

	CreatedDate time.Time `json:"created" db:"created"`

	permissions Permissions
}

//Assign a permission to a juno.StdRole
func (r *StdRole) Assign(p Permission) error {
	r.Lock()
	defer r.Unlock()
	if r.Has(p) {
		return errors.New("Role already has permission assigned")
	}
	if r.permissions == nil {
		r.permissions = make(Permissions)
	}
	r.permissions[p.ID()] = p
	return nil
}

//Has verifies of a StdRole has permission
func (r *StdRole) Has(p Permission) bool {
	if r.permissions == nil {
		return false
	}
	_, exists := r.permissions[p.ID()]
	return exists
}

//Revoke a permission from a StdRole - return an error if the StdRole is currently not granted the permission attempted to revoke
func (r *StdRole) Revoke(p Permission) error {
	r.Lock()
	defer r.Unlock()
	if !r.Has(p) {
		return errors.New("Role does not have permission assigned")
	}
	delete(r.permissions, p.ID())
	return nil
}
