package acl

import (
	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
)

// AclObject 代表user, role或者权限的target对象类型
type AclObject interface {
	GetSid() string // 通常是 uuid
	GetTyp() string // object类型, user, role, post, etc...
}

// AclManager acl manager
type AclManager struct {
	db *gorm.DB
	rp *redis.Pool
}

// CreateAclManager create ACL manager
func CreateAclManager(db *gorm.DB, pool *redis.Pool) *AclManager {
	return &AclManager{
		db: db,
		rp: pool,
	}
}

// IsGrant 判断是否有权限
// who:  主体
// what: 权限target
// perm: 要求的权限
//
// 0. 如果who不是role，查找who关联的roles
// 1. 从who关联的rules逐条执行, 如果有结果, 返回结果;
// 2. 从who关联的acl entry中逐条判断，如果有结果，返回结果；
// 3. 返回false
func (mgr *AclManager) IsGrant(who AclObject, what AclObject, perm Permission) bool {
	if perm.Mask == 0 {
		// 非法的Mask
		return false
	}

	rc := mgr.rp.Get()
	sids := GetPrincipals(mgr.db, rc, who)
	//sids = append(sids, Principal(who))

	//fmt.Printf("who: %v  principals: %v what: %v perm: %v\n", who, sids, what, perm)

	for _, sid := range sids {
		result := isGrant(mgr.db, rc, sid, what, perm)
		if result == 0 {
			continue
		} else if result == Grant {
			return true
		} else {
			return false
		}
	}

	// 默认返回拒绝
	return false
}

//
// func (mgr *AclManager)  {
// }
