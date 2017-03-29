package acl

import (
	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
)

// Principal 主体
// Pricipal 可以是一个用户或者一个role, 是权限主体的最小单位
type Principal struct {
	Sid string
	Typ string
}

// isGrant principal 是否被授权或拒绝
// 没有策略: 返回0
// 拒绝:     返回Reject
// 允许:     返回Grant
//
// 0. 从rule中判定是否被允许或拒绝
// 1. 从acl中判定是否被允许或拒绝
//
func (sid Principal) isGrant(db *gorm.DB, rc redis.Conn, obj AclObject, perm Permission) int {
	rules := GetRules(sid.Sid)

	for _, r := range rules {
		ret := r.fn(sid, obj, perm)
		if ret > 0 {
			return Grant
		} else if ret < 0 {
			return Reject
		}
	}

	return 0
}
