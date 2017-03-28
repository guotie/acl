package acl

const (
	// permissionRead Read Permission
	permissionRead = "READ"
	// permissionWrite Write or update Permission
	permissionWrite = "WRITE"
	// permissionCreate Create Permission
	permissionCreate = "CREATE"
	// permissionDelete Delete Permission
	permissionDelete = "DELETE"
	// permissionManage Manage Permission
	permissionManage = "Manage"

	// permission index
	permIndexRead = iota
	permIndexWrite
	permIndexCreate
	permIndexDelete
	permIndexManage

	PermissionRead   = 1 << permIndexRead   // 读权限
	PermissionWrite  = 1 << permIndexWrite  // 写权限
	PermissionCreate = 1 << permIndexCreate // 创建权限
	PermissionDelete = 1 << permIndexDelete // 删除权限
	PermissionManage = 1 << permIndexManage // 管理权限
)

// Permission Permission definition
type Permission struct {
	Mask uint32
	Name string
}
