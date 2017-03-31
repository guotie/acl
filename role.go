package acl

// 定义Role
// Role 与 User的关系

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"github.com/smtc/glog"
)

// RoleTyp Role Type
const (
	RoleTyp    = "RoleTyp"
	AccountTyp = "AccountTyp"
)

// Role role
type Role struct {
	ID        int64  `gorm:"column:id"`
	Name      string `gorm:"column:name;size:100"`
	Sid       string `gorm:"size:100;unique_index"`
	Depart    string
	Corp      string
	City      string
	Province  string
	Level     int
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
	UID       string `gorm:"column:uid;size:100;index"`
	Rid       string `gorm:"size:100;index"`
	CreatedAt time.Time
}

// GetPricipals get principals
func (mgr *AclManager) GetPricipals(who AclObject) []Principal {
	return GetUserPrincipals(mgr.db, mgr.rp.Get(), who)
}

// GetPrincipals get principals
func GetPrincipals(db *gorm.DB, rc redis.Conn, who AclObject) []Principal {
	// 如果是Role, 由于目前没有层级role, 这里直接返回role sid
	if who.GetTyp() == RoleTyp {
		return []Principal{Principal(who)}
	}

	// who为用户的情况
	if who.GetTyp() == AccountTyp {
		ps := GetUserPrincipals(db, rc, who)
		return ps
	}

	glog.Warn("GetPrincipals: unknown Principal type %s, sid=%s\n", who.GetTyp(), who.GetSid())
	return []Principal{Principal(who)}
}

// GetUserPrincipals Get User roles
func GetUserPrincipals(db *gorm.DB, rc redis.Conn, who AclObject) []Principal {
	roles, err := getUserRoles(db, rc, who)
	if err != nil {
		glog.Error("GetUserPrincipals: getUserRoles failed: %v\n", err)
		return []Principal{Principal(who)}
	}
	return append(roles, Principal(who))
}

// getUserRoles get user roles
func getUserRoles(db *gorm.DB, rc redis.Conn, who AclObject) ([]Principal, error) {
	sids, err := getUserRolesFromRedis(rc, who)
	if err == nil {
		return sids, nil
	}
	var res []Principal
	res, err = getUserRolesFromDB(db, who)
	if err != nil {
		glog.Error("getUserRolesFromDB failed: %v\n", err)
		return []Principal{}, err
	}

	// 保存到缓存中
	setUserRolesToRedis(rc, who.GetSid(), res)

	return res, nil
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
		return []Principal{}, ErrRedisKeyNotExist
	}

	// 取出 key 下的所有 members
	reply, err := redis.ByteSlices(rc.Do("SMEMBERS", key))
	if err != nil {
		return []Principal{}, err
	}

	for _, r := range reply {
		var p Role
		err = json.Unmarshal(r, &p)
		if err != nil {
			return nil, err
		}
		sids = append(sids, &p)
	}
	sids = append(sids, Principal(who))

	return sids, nil
}

// getUserRolesFromDB 从数据库中获取用户roles
func getUserRolesFromDB(db *gorm.DB, who AclObject) ([]Principal, error) {
	var (
		rs []*Role
		ps []Principal
	)

	// 关联查找
	//db.LogMode(true)
	err := db.Raw("select roles.* from roles left join user_roles AS ur on ur.rid=roles.sid where ur.uid=?",
		who.GetSid()).Scan(&rs).Error
	if err != nil {
		return nil, err
	}

	for _, r := range rs {
		//fmt.Println(r)
		ps = append(ps, Principal(r))
	}
	return ps, err
}

// setUserRolesToRedis 设置用户roles到redis中
func setUserRolesToRedis(rc redis.Conn, sid string, sids []Principal) {
	key := KeyUserRole(sid)
	for _, sid := range sids {
		val, err := json.Marshal(sid)
		if err != nil {
			glog.Error("setUserRolesToRedis Marshal failed: %v\n", err)
			rc.Do("DEL", key)
			return
		}
		_, err = rc.Do("SADD", key, val)
		if err != nil {
			glog.Error("setUserRolesToRedis failed: %v\n", err)
			rc.Do("DEL", key)
			return
		}
	}
}

////////////////////////////Role//////////////////////////////

// CreateRole 创建role
func (mgr *AclManager) CreateRole(name, sid, depart, corp, city, province string, level int) (*Role, error) {
	var r Role

	// 删除缓存
	mgr.EvictRoleCache()

	sid = strings.TrimSpace(sid)
	if sid == "" {
		sid = uuid.NewV4().String()
	}

	r = Role{
		Name:     name,
		Sid:      sid,
		Depart:   depart,
		Corp:     corp,
		City:     city,
		Province: province,
		Level:    level,
	}
	err := mgr.db.Create(&r).Error

	return &r, err
}

// GetRole get role from db
func (mgr *AclManager) GetRole(sid string) (*Role, error) {
	var r Role

	err := mgr.db.Where("sid=?", sid).Find(&r).Error
	return &r, err
}

// DeleteRole 删除role
func (mgr *AclManager) DeleteRole(sid string) {
	// 删除缓存
	mgr.EvictRoleCache()

	mgr.db.Where("sid=?", sid).Delete(&Role{})
}

// RenameRole 修改Role的名字
// nname: 新名字
func (mgr *AclManager) RenameRole(nname, sid string) error {
	// 删除缓存 仅修改role name, 不需要删除缓存
	// mgr.DeleteRoleCache()

	err := mgr.db.Model(Role{}).Where("sid=?", sid).Update("name", nname).Error
	return err
}

////////////////////////////UserRole//////////////////////////////

// AddUserRoleRelation 增加用户到Role中
func (mgr *AclManager) AddUserRoleRelation(uid, rid string) error {
	var (
		count int
		ur    UserRole
	)
	// 先检查记录是否已经存在, 如果存在, 不做任何操作
	mgr.db.Model(&UserRole{}).Where("uid = ? AND rid = ?", uid, rid).Count(&count)
	if count > 0 {
		glog.Warn("AddUserRoleRelation: uid %s rid %s has exist.\n", uid, rid)
		return nil
	}

	// 删除缓存
	mgr.EvictUserRoleCache(uid)
	// 创建
	ur.UID = uid
	ur.Rid = rid
	err := mgr.db.Create(&ur).Error
	if err != nil {
		glog.Error("AddUserRoleRelation: Create uid %s rid %s failed: %v\n", uid, rid, err)
	}
	return err
}

// DelUserRoleRelation 从Role中删除用户
// byUser  User的sid
// byRole  Role的sid
//
func (mgr *AclManager) DelUserRoleRelation(byUser, byRole string) {
	if byUser == "" && byRole == "" {
		glog.Warn("DelUserRoleRelation: invalid param, byUser & byRole should NOT be all empty\n")
		return
	}
	// 删除缓存
	if byRole != "" {
		mgr.EvictRoleCache()
	} else if byUser != "" {
		mgr.EvictUserRoleCache(byUser)
	}

	var clause = mgr.db
	if byRole != "" {
		clause = clause.Where("rid=?", byRole)
	}
	if byUser != "" {
		clause = clause.Where("uid=?", byUser)
	}

	err := clause.Delete(UserRole{}).Error
	if err != nil {
		glog.Error("DelUserRoleRelation: byUser=%s byRole=%s error=%v\n", byUser, byRole, err)
	}
}
