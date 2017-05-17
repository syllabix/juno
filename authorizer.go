package juno

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

//AuthRepo is an interface that should be implemented by a repository that provides persistance to the data used in an Authorizer
type AuthRepo interface {
	GetPermissions() ([]Permission, error)
	GetPermission(Permission) (Permission, error)
	CreatePermission(Permission) (Permission, error)

	GetRoles() ([]Role, error)
	GetRole(Role) (Role, error)
	CreateRole(Role) (Role, error)

	GetRolePermissions() ([]RolePermission, error)
	AssignPermissionToRole(Role, Permission) error
	RevokePermissionFromRole(Role, Permission) error
}

//Authorizer is the struct (with intended use as a singleton) for handling all things authorization
type Authorizer struct {
	sync.RWMutex
	repo        AuthRepo
	roles       Roles
	permissions Permissions
	superadmin  Role
}

//NewAuthorizer is a factory constructor for getting a properly instantiated Authorizer
func NewAuthorizer(repo AuthRepo) *Authorizer {
	//set up an instance of Authorizer
	mngr := &Authorizer{
		repo:        repo,
		roles:       make(Roles),
		permissions: make(Permissions),
	}

	//get permissions in the DB
	perms, err := repo.GetPermissions()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	//assign permissions to the cache
	for _, p := range perms {
		mngr.permissions[p.ID()] = p
	}

	//get roles
	roles, err := repo.GetRoles()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	//assign roles to cache
	for _, r := range roles {
		mngr.roles[r.ID()] = r
	}

	//get role/permission relationships
	rolePerms, err := repo.GetRolePermissions()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	//registers role assignments in the cache
	for _, rp := range rolePerms {
		if perm, ok := mngr.permissions[rp.PermissionID()]; ok {
			mngr.roles[rp.RoleID()].Assign(perm)
		}
	}

	return mngr
}

func (mngr *Authorizer) hasPermission(p Permission) bool {
	_, exists := mngr.permissions[p.ID()]
	return exists
}

func (mngr *Authorizer) hasRole(r UserRole) bool {
	_, exists := mngr.roles[r.ID()]
	return exists
}

//Granted verifies if a role specified by role name is currently granted a permission
func (mngr *Authorizer) Granted(role UserRole, p Permission) bool {
	mngr.Lock()
	defer mngr.Unlock()
	if role, exists := mngr.roles[role.ID()]; exists {
		return role.Has(p)
	}
	return false
}

//GetPermissions returns all permissions
func (mngr *Authorizer) GetPermissions() ([]Permission, error) {
	return mngr.repo.GetPermissions()
}

//GetRoles returns all roles
func (mngr *Authorizer) GetRoles() ([]Role, error) {
	return mngr.repo.GetRoles()
}

//AddPermission adds a permission the auth mngr
func (mngr *Authorizer) AddPermission(p Permission) Permission {
	mngr.Lock()
	defer mngr.Unlock()
	newPerm, err := mngr.repo.CreatePermission(p)
	if err != nil {
		//check for duplicate error and ignore if that's what sql server returns
		if !strings.Contains(strings.ToLower(err.Error()), "cannot insert duplicate key") {
			log.Fatalf("Unable to create permission: %s\n%s", p.ID(), err.Error())
			return nil
		}
		newPerm, err = mngr.repo.GetPermission(p)
		if err != nil {
			log.Fatal("A fatal error occurred setting up permissions:", err.Error())
			return nil
		}
	}
	if mngr.superadmin != nil {
		mngr.roles[mngr.superadmin.ID()].Assign(newPerm)
	}
	mngr.permissions[p.ID()] = newPerm
	return newPerm
}

func (mngr *Authorizer) CreateRole(r Role) (Role, error) {
	mngr.Lock()
	defer mngr.Unlock()
	if _, exists := mngr.roles[r.ID()]; !exists {
		newrole, err := mngr.repo.CreateRole(r)
		if err != nil {
			return nil, err
		}
		mngr.roles[newrole.ID()] = newrole
		return newrole, nil
	}
	return nil, fmt.Errorf("Role with ID %s already exists", r.ID())
}

func (mngr *Authorizer) assignSuperAdmin(admin Role) {
	mngr.superadmin = admin
	for _, perm := range mngr.permissions {
		mngr.superadmin.Assign(perm)
	}
	mngr.roles[admin.ID()] = mngr.superadmin
}

//RevokePermissionFromRole removes roles grant to a permission
func (mngr *Authorizer) RevokePermissionFromRole(role Role, perm Permission) error {
	mngr.Lock()
	defer mngr.Unlock()
	if role, exists := mngr.roles[role.ID()]; exists {
		err := mngr.repo.RevokePermissionFromRole(role, perm)
		if err != nil {
			return err
		}
		return role.Revoke(perm)
	}
	return fmt.Errorf("RoleID with ID '%s' does not exist", role.ID())
}

//CreateSuperAdmin is a method on Authorizor to create a Role that is granted all permissions
func (mngr *Authorizer) CreateSuperAdmin(r Role) error {
	superAdmin, _ := mngr.repo.GetRole(r)
	if superAdmin != nil {
		mngr.assignSuperAdmin(superAdmin)
		return nil
	}
	superAdmin, err := mngr.CreateRole(r)
	if err != nil {
		return err
	}
	mngr.assignSuperAdmin(superAdmin)
	return nil
}

//AssignPermissionToRole takes a role and grants access to the provided permission
func (mngr *Authorizer) AssignPermissionToRole(role Role, perm Permission) error {
	if !mngr.hasRole(role) {
		return fmt.Errorf("RoleID with ID '%s' does not exist", role.ID())
	}
	err := mngr.repo.AssignPermissionToRole(role, perm)
	if err != nil {
		return err
	}
	err = mngr.roles[role.ID()].Assign(perm)
	return err
}
