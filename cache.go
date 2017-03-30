package acl

import (
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/smtc/glog"
)

// 缓存的key

const (
	PrefixUserRole = "UserRole_" // 用户对应的role
	PrefixAclEntry = "AclEntry_" // pricipal对应的Acl Entry
)

// KeyUserRole UserRole key
// 用户角色对应表
// 在redis cache中是一个set
func KeyUserRole(sid string) string {
	return PrefixUserRole + sid
}

// KeyAclEntry AclEntry key
// sid: principal string
// HashMap
// 用户或角色对应的AclEntrys
// Hset AclEntry_{{ user or role sid}} AclEntry.Sid AclEntry
func KeyAclEntry(sid string) string {
	return PrefixAclEntry + sid
}

/////////////////////////////////Cache相关操作///////////////////////////////////////

// CacheKeys 获得所有的已prefix开头的key
// 如果prefix为空, 获取所有的keys
func (mgr *AclManager) CacheKeys(prefix string) ([]string, error) {
	rc := mgr.rp.Get()
	vs, err := redis.Strings(rc.Do("KEYS", "*"))

	if err != nil {
		return nil, err
	}
	if prefix == "" {
		return vs, nil
	}
	// 根据prefix过滤
	res := []string{}

	for _, s := range vs {
		if strings.HasPrefix(s, prefix) {
			res = append(res, s)
		}
	}

	return res, nil
}

// CleanCache 清除ACL的缓存
func (mgr *AclManager) CleanCache() error {
	rc := mgr.rp.Get()
	keys, err := mgr.CacheKeys("")
	if err != nil {
		return err
	}

	for _, key := range keys {
		rc.Do("DEL", key)
	}
	return nil
}

// EvictRoleCache 删除Role相关的缓存
// 找到所有的UserRole_开头的key, 删掉
func (mgr *AclManager) EvictRoleCache() {
	keys, err := mgr.CacheKeys(PrefixUserRole)
	if err != nil {
		glog.Error("EvictRoleCache: Get User-Role keys failed: %v\n", err)
		return
	}

	rc := mgr.rp.Get()
	for _, key := range keys {
		_, err := rc.Do("DEL", key)
		if err != nil {
			glog.Error("EvictRoleCache: Delete key %s failed: %v\n", key, err)
		}
	}
}

// EvictACECache 删除所有AclEntry缓存
func (mgr *AclManager) EvictACECache() {
	keys, err := mgr.CacheKeys(PrefixAclEntry)
	if err != nil {
		glog.Error("EvictACECache: Get User-ACE keys failed: %v\n", err)
		return
	}

	rc := mgr.rp.Get()
	for _, key := range keys {
		_, err := rc.Do("DEL", key)
		if err != nil {
			glog.Error("EvictACECache: Delete key %s failed: %v\n", key, err)
		}
	}
}

// EvictUserRoleCache 删除User Role
// 找到所有的UserRole_开头的key, 删掉
func (mgr *AclManager) EvictUserRoleCache(sid string) {
	rc := mgr.rp.Get()
	key := KeyUserRole(sid)
	_, err := rc.Do("DEL", key)
	if err != nil {
		glog.Error("EvictUserRoleCache: Delete key %s failed: %v\n", key, err)
	}
}

// EvictPrincipalACECache 删除Principal 的 AclEntry缓存
func (mgr *AclManager) EvictPrincipalACECache(sid string) {
	rc := mgr.rp.Get()
	key := KeyAclEntry(sid)
	_, err := rc.Do("DEL", key)
	if err != nil {
		glog.Error("EvictPrincipalACECache: Delete key %s failed: %v\n", key, err)
	}
}
