package mssqlrepo

import (
	"database/sql"
	"fmt"
	"time"

	"reflect"

	"strconv"

	"github.com/syllabix/juno"
)

//The NewAuthRepo func return a fully instantiated auth repository that implements the juno.AuthRepo interface
func NewAuthRepo(db *sql.DB) *AuthRepo {
	return &AuthRepo{
		db: db,
	}
}

//AuthRepo is a the struct the implements the AuthRepo interface for MSSQL
type AuthRepo struct {
	db *sql.DB
}

const getpermissions = `SELECT PermissionID, Label, Description FROM dbo.Permissions`

//GetPermissions returns all permissions
func (r *AuthRepo) GetPermissions() ([]juno.Permission, error) {
	rows, err := r.db.Query(getpermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []juno.Permission{}
	for rows.Next() {
		permission := new(juno.StdPermission)
		err := rows.Scan(
			&permission.PermissionID,
			&permission.Label,
			&permission.Description,
		)
		if err == nil {
			results = append(results, permission)
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return results, nil
}

const getpermbyname = `SELECT PermissionID, Label, Description FROM dbo.Permissions WHERE Label = ?`

func (r *AuthRepo) GetPermission(p juno.Permission) (juno.Permission, error) {
	stdPerm, ok := p.(*juno.StdPermission)
	if !ok {
		return nil, fmt.Errorf("Unexpected error type of %s recieved, expected %s", reflect.TypeOf(p), "*juno.StdPermission")
	}
	permission := new(juno.StdPermission)
	err := r.db.QueryRow(getpermbyname, stdPerm.Label).Scan(&permission.PermissionID, &permission.Label, &permission.Description)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

const insertpermission = `INSERT INTO dbo.Permissions (Label, Description) VALUES (?, ?)`

//AddPermission takes an implementation of the juno.Permission interface to create the permission
func (r *AuthRepo) CreatePermission(p juno.Permission) (juno.Permission, error) {
	if stdPerm, ok := p.(*juno.StdPermission); ok {
		result, err := r.db.Exec(insertpermission, stdPerm.Label, stdPerm.Description)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		stdPerm.PermissionID = int(id)
		return stdPerm, nil
	}
	return nil, fmt.Errorf("Invalid Permissions type of %s passed to add function. Expecting juno.StdPermission", reflect.TypeOf(p))
}

const getroles = `SELECT * FROM dbo.UserRoles`

func (r *AuthRepo) GetRoles() ([]juno.Role, error) {
	rows, err := r.db.Query(getroles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []juno.Role{}
	for rows.Next() {
		role := new(juno.StdRole)
		err := rows.Scan(
			&role.RoleID,
			&role.RoleName,
			&role.CreatedDate,
		)
		if err == nil {
			results = append(results, role)
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return results, nil
}

const insertrole = `INSERT INTO dbo.UserRoles (RoleName) VALUES (?)`

func (r *AuthRepo) CreateRole(role juno.Role) (juno.Role, error) {
	if stdrole, ok := role.(*juno.StdRole); ok {
		stdrole.CreatedDate = time.Now()
		result, err := r.db.Exec(insertrole, stdrole.RoleName)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		stdrole.RoleID = int(id)
		return stdrole, nil
	}
	return nil, fmt.Errorf("Invalid Role type of %s passed to CreateRole. Expecting juno.StdRole", reflect.TypeOf(role))
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

const getrolepermissions = `
        SELECT UserRoles.RoleID, Permissions.PermissionID
        FROM UserRolePermissionsMap
        JOIN UserRoles ON UserRolePermissionsMap.RoleID = UserRoles.RoleID
        JOIN Permissions ON UserRolePermissionsMap.PermissionID = Permissions.PermissionID`

//GetRolePermissions returns a slice of RolePermission which is intended to associate a role with a granted permission
func (r *AuthRepo) GetRolePermissions() ([]juno.RolePermission, error) {
	rows, err := r.db.Query(getrolepermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []juno.RolePermission{}
	for rows.Next() {
		rp := new(RolePermission)
		err := rows.Scan(
			&rp.RID,
			&rp.PID,
		)
		if err == nil {
			results = append(results, rp)
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return results, nil
}

const insertrolepermission = `INSERT INTO dbo.UserRolePermissionsMap (RoleID, PermissionID) VALUES (?, ?)`

//AssignPermissionToRole grants a role a permission
func (r *AuthRepo) AssignPermissionToRole(role juno.Role, perm juno.Permission) error {
	stdRole, ok := role.(*juno.StdRole)
	if !ok {
		return fmt.Errorf("Invalid Role type of %s passed to CreateRole. Expecting juno.StdRole", reflect.TypeOf(role))
	}
	stdPerm, ok := perm.(*juno.StdPermission)
	if !ok {
		return fmt.Errorf("Invalid Permissions type of %s passed to add function. Expecting juno.StdPermission", reflect.TypeOf(perm))
	}
	_, err := r.db.Exec(insertrolepermission, stdRole.RoleID, stdPerm.PermissionID)
	return err
}

const revokepermission = `DELETE FROM dbo.UserRolePermissionsMap WHERE RoleID = ? AND PermissionID = ?`

//RevokePermissionFromRole takes a juno.StdRole and juno.StdPermission (implementations of the respective interface) and removes their grant relationship in the database.
func (r *AuthRepo) RevokePermissionFromRole(role juno.Role, perm juno.Permission) error {
	stdRole, ok := role.(*juno.StdRole)
	if !ok {
		return fmt.Errorf("Invalid Role type of %s passed to CreateRole. Expecting juno.StdRole", reflect.TypeOf(role))
	}
	stdPerm, ok := perm.(*juno.StdPermission)
	if !ok {
		return fmt.Errorf("Invalid Permissions type of %s passed to add function. Expecting juno.StdPermission", reflect.TypeOf(perm))
	}
	_, err := r.db.Exec(revokepermission, stdRole.RoleID, stdPerm.PermissionID)
	return err
}

const getrole = `SELECT * FROM dbo.UserRoles WHERE RoleName = ?`

func (r *AuthRepo) GetRole(role juno.Role) (juno.Role, error) {
	stdRole, ok := role.(*juno.StdRole)
	if !ok {
		return nil, fmt.Errorf("Invalid Role type of %s passed to GetRole. Expecting juno.StdRole", reflect.TypeOf(role))
	}
	var retRole juno.StdRole
	err := r.db.QueryRow(getrole, stdRole.RoleName).Scan(&retRole.RoleID, &retRole.RoleName, &retRole.CreatedDate)
	if err != nil {
		return nil, err
	}
	return &retRole, nil
}
