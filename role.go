package acl

// 定义Role
// Role 与 User的关系

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/smtc/glog"
)

// RoleTyp Role Type
const RoleTyp = "Role"

// Role role
type Role struct {
	ID        int64  `gorm:"column:id"`
	Name      string `gorm:"column:name size:100"`
	Sid       string `gorm:"size:100"`
	CreatedAt time.Time
}

// GetSid get role sid
func (r *Role) GetSid() string {
	return r.Sid
}

// GetTyp get role type
func (*Role) GetTyp() string {
	return RoleTyp
}

// UserRole user 和role的对应关系表
type UserRole struct {
	ID        int64  `gorm:"column:id"`
	UID       string `gorm:"column:uid size:100"`
	Rid       string `gorm:"size:100"`
	CreatedAt time.Time
}

// GetPrincipals get principals
func GetPrincipals(db *gorm.DB, rc redis.Conn, who AclObject) []Principal {
	// 如果是Role, 由于目前没有层级role, 这里直接返回role sid
	if who.GetTyp() == RoleTyp {
		return []Principal{Principal{who.GetSid(), who.GetTyp()}}
	}

	// who为用户的情况
	if who.GetTyp() == "Account" {
		return GetUserPrincipals(db, rc, who)
	}

	return []Principal{Principal{who.GetSid(), who.GetTyp()}}
}

// GetUserPrincipals Get User roles
func GetUserPrincipals(db *gorm.DB, rc redis.Conn, who AclObject) []Principal {
	roles := getUserRoles(db, rc, who)
	return append(roles, Principal{who.GetSid(), who.GetTyp()})
}

// getUserRoles get user roles
func getUserRoles(db *gorm.DB, rc redis.Conn, who AclObject) []Principal {
	sids, err := getUserRolesFromRedis(rc, who)
	if err == nil {
		return sids
	}
	var res []Principal
	res, err = getUserRolesFromDB(db, who)
	if err != nil {
		glog.Error("getUserRolesFromDB failed: %v\n", err)
		return []Principal{}
	}

	// 保存到缓存中
	setUserRolesToRedis(rc, who.GetSid(), res)

	return res
}

// getUserRolesFromRedis 从redis中获取用户roles
func getUserRolesFromRedis(rc redis.Conn, who AclObject) ([]Principal, error) {
	var sids []Principal

	key := KeyUserRole(who.GetSid())
	exist, err := redis.Bool(rc.Do("EXISTS", key))
	if err != nil {
		// redis出错
		return []Principal{}, err
	}
	if exist == false { // key 不存在
		return []Principal{}, fmt.Errorf("UserRole cache %s Not exist", key)
	}

	// 取出 key 下的所有 members
	reply, err := redis.ByteSlices(rc.Do("SMEMBERS", key))
	if err != nil {
		return []Principal{}, err
	}

	for _, r := range reply {
		var p Principal
		err = json.Unmarshal(r, &p)
		if err != nil {
			return nil, err
		}
		sids = append(sids, p)
	}

	return sids, nil
}

// getUserRolesFromDB 从数据库中获取用户roles
func getUserRolesFromDB(db *gorm.DB, who AclObject) ([]Principal, error) {
	var (
		ur   []UserRole
		sids []Principal
	)

	if err := db.Where("uid=?", who.GetSid()).Find(&ur).Error; err != nil {
		return nil, err
	}

	for _, r := range ur {
		sids = append(sids, Principal{r.Rid, RoleTyp})
	}

	return sids, nil
}

// setUserRolesToRedis 设置用户roles到redis中
func setUserRolesToRedis(rc redis.Conn, sid string, sids []Principal) {
	rc.Do("SADD", redis.Args{}.Add(sid).AddFlat(sids))
}
