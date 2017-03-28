package acl

import (
	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
)

// Principal 主体
// Pricipal 可以是一个用户或者一个role, 是权限主体的最小单位
type Principal string

// isGrant principal 是否被授权或拒绝
func (sid Principal) isGrant(db *gorm.DB, rc redis.Conn, obj AclObject) int {
	return 0
}
