package acl

// 缓存的key

// KeyUserRole UserRole key
// 用户角色对应表
// 在redis cache中是一个set
func KeyUserRole(sid string) string {
	return "UserRole_" + sid
}

// KeyAclEntry AclEntry key
// sid: principal string
func KeyAclEntry(sid string) string {
	return "AclEntry_" + sid
}
