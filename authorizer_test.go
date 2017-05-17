package juno

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"log"
)

//Mock Permissions
var (
	update    *StdPermission
	canDelete *StdPermission
	create    *StdPermission
	read      *StdPermission
)

//Mock Roles

var (
	admin   *StdRole
	blogger *StdRole
	manager *StdRole
	sales   *StdRole
)

func init() {
	log.Println("testing...")
	update = NewStdPermission("update", "You can update things")
	update.PermissionID = 1

	canDelete = NewStdPermission("delete", "You can delete things")
	canDelete.PermissionID = 2

	create = NewStdPermission("create", "You can create things")
	create.PermissionID = 3

	read = NewStdPermission("read", "You can read things")
	read.PermissionID = 4

	admin = NewStdRole("admin")
	admin.RoleID = 1

	blogger = NewStdRole("blogger")
	blogger.RoleID = 2

	manager = NewStdRole("manager")
	manager.RoleID = 3

	sales = NewStdRole("sales")
	sales.RoleID = 4
}

type MockAuthRepo struct{}

func (repo *MockAuthRepo) GetPermissions() ([]Permission, error) {
	return []Permission{update, canDelete, create, read}, nil
}

func (repo *MockAuthRepo) CreatePermission(p Permission) (Permission, error) {
	return NewStdPermission(p.ID(), "Mock permission"), nil
}

func (repo *MockAuthRepo) GetRoles() ([]Role, error) {
	return []Role{admin, blogger, sales, manager}, nil
}

func (repo *MockAuthRepo) GetRole(r Role) (Role, error) {
	return nil, nil
}

func (repo *MockAuthRepo) CreateRole(r Role) (Role, error) {
	return r, nil
}

//RolePermission is an implementation of RolePersmission, and used to expose the role/permission grant relationships to Authorizer
type MockRolePermission struct {
	RID int `db:"RoleId"`
	PID int `db:"PermissionID"`
}

//RoleID implemetnts the RolePermission RoldID getter
func (rp *MockRolePermission) RoleID() string {
	return strconv.Itoa(rp.RID)
}

//PermissionID implemetnts the RolePermission PermissionID getter
func (rp *MockRolePermission) PermissionID() string {
	return strconv.Itoa(rp.PID)
}

func (repo *MockAuthRepo) GetRolePermissions() ([]RolePermission, error) {
	rp := new(MockRolePermission)
	rp.PID = 1
	rp.RID = 1
	return []RolePermission{rp}, nil
}

func (repo *MockAuthRepo) AssignPermissionToRole(r Role, p Permission) error {
	return nil
}

func (repo *MockAuthRepo) RevokePermissionFromRole(r Role, p Permission) error {
	return nil
}

func (repo *MockAuthRepo) GetPermission(p Permission) (Permission, error) {
	return NewStdPermission("Test", "A description for the test permission"), nil
}

func mockAuthorizer() *Authorizer {
	repo := new(MockAuthRepo)
	return NewAuthorizer(repo)
}

func TestNewAuthorizer(t *testing.T) {
	assert := assert.New(t)

	authorizer := mockAuthorizer()

	assert.Equal(4, len(authorizer.permissions), "Authorizer should have 4 permissions")
	assert.Equal(4, len(authorizer.roles), "Authorizer should have 4 roles")
	assert.True(admin.Has(update))
}

func TestCreateSuperAdmin(t *testing.T) {
	assert := assert.New(t)

	authorizer := mockAuthorizer()

	superadmin := NewStdRole("SuperAdmin")
	superadmin.RoleID = 0
	err := authorizer.CreateSuperAdmin(superadmin)
	assert.Equal(5, len(authorizer.roles), "Authorizer should have 5 roles, including SuperAdmin")
	assert.Nil(err, "Super admin should be created without error")
	assert.True(authorizer.Granted(superadmin, update), "Super admin should be granted permission to update")
	assert.True(authorizer.Granted(superadmin, create), "Super admin should be granted permission to create")
	assert.True(authorizer.Granted(superadmin, canDelete), "Super admin should be granted permission to canDelete")
	assert.True(authorizer.Granted(superadmin, read), "Super admin should be granted permission to read")

	fun := authorizer.AddPermission(NewStdPermission("fun", "Allows user to have fun"))
	assert.True(authorizer.Granted(superadmin, fun), "Super admin should be granted permissions if they are added after the role is created")
}

func TestRevokePermissionFromRole(t *testing.T) {
	assert := assert.New(t)

	authorizer := mockAuthorizer()
	err := authorizer.RevokePermissionFromRole(admin, update)
	assert.Nil(err, "Revoking permission on role previously granted permission should work without error")
	assert.False(authorizer.Granted(admin, update))

	fakeRole := NewStdRole("fake")
	fakeRole.RoleID = 9999
	err = authorizer.RevokePermissionFromRole(fakeRole, update)
	assert.Error(err, "Revoking permission on non existant role should return an error")
}

func TestAddPermission(t *testing.T) {
	assert := assert.New(t)

	authorizer := mockAuthorizer()
	doStuff := authorizer.AddPermission(NewStdPermission("Do Stuff", "Allows the user to do stuff"))
	assert.NotNil(doStuff, "Do stuff should be successfully created")
	assert.True(authorizer.hasPermission(doStuff), "Instance of authorizer should have have the doStuff permission")
}

func TestAssignPermissionToRole(t *testing.T) {
	assert := assert.New(t)

	authorizer := mockAuthorizer()
	err := authorizer.AssignPermissionToRole(blogger, canDelete)
	assert.Nil(err, "Assigning permission to existing role should work")
	assert.True(authorizer.Granted(blogger, canDelete), "Blogger should have canDelete permissions after being assigned")

}
