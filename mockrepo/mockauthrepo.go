package mockrepo

import (
	"strconv"

	"github.com/syllabix/juno"
)

//Mock Permissions
var (
	update *juno.StdPermission
	delete *juno.StdPermission
	create *juno.StdPermission
	read   *juno.StdPermission
)

//Mock Roles

var (
	admin   *juno.StdRole
	blogger *juno.StdRole
	manager *juno.StdRole
	sales   *juno.StdRole
)

func init() {
	update = juno.NewStdPermission("update", "You can update things")
	update.PermissionID = 1

	delete = juno.NewStdPermission("delete", "You can delete things")
	delete.PermissionID = 2

	create = juno.NewStdPermission("create", "You can create things")
	create.PermissionID = 3

	read = juno.NewStdPermission("read", "You can read things")
	read.PermissionID = 4

	admin = juno.NewStdRole("admin")
	admin.RoleID = 1

	blogger = juno.NewStdRole("blogger")
	blogger.RoleID = 2

	manager = juno.NewStdRole("manager")
	manager.RoleID = 3

	sales = juno.NewStdRole("sales")
	manager.RoleID = 4
}

//RolePermission is an implementation of juno.RolePersmission, and used to expose the role/permission grant relationships to Authorizer
type RolePermission struct {
	RID int `db:"RoleId"`
	PID int `db:"PermissionID"`
}

//RoleID implemetnts the juno.RolePermission RoldID getter
func (rp *RolePermission) RoleID() string {
	return strconv.Itoa(rp.RID)
}

//PermissionID implemetnts the juno.RolePermission PermissionID getter
func (rp *RolePermission) PermissionID() string {
	return strconv.Itoa(rp.PID)
}

type MockAuthRepo struct{}

func (repo *MockAuthRepo) GetPermissions() ([]juno.Permission, error) {
	return []juno.Permission{update, delete, create, read}, nil
}

func (repo *MockAuthRepo) CreatePermission(p juno.Permission) (juno.Permission, error) {
	return juno.NewStdPermission(p.ID(), "Mock permission"), nil
}

func (repo *MockAuthRepo) GetRoles() ([]juno.Role, error) {
	return []juno.Role{admin, blogger, sales, manager}, nil
}

func (repo *MockAuthRepo) GetRole(r juno.Role) (juno.Role, error) {
	return juno.NewStdRole(r.ID()), nil
}

func (repo *MockAuthRepo) CreateRole(r juno.Role) (juno.Role, error) {
	return juno.NewStdRole(r.ID()), nil
}

func (repo *MockAuthRepo) GetPermission(p juno.Permission) (juno.Permission, error) {
	return juno.NewStdPermission("Test", "A description for the test permission"), nil
}

func (repo *MockAuthRepo) GetRolePermissions() ([]juno.RolePermission, error) {
	rp := new(RolePermission)
	rp.PID = 1
	rp.RID = 1
	return []juno.RolePermission{rp}, nil
}

func (repo *MockAuthRepo) AssignPermissionToRole(r juno.Role, p juno.Permission) error {
	return nil
}

func (repo *MockAuthRepo) RevokePermissionFromRole(r juno.Role, p juno.Permission) error {
	return nil
}
