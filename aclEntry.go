package acl

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/smtc/glog"
	"github.com/smtc/goutils"
)

// AclEntry Acl Entry, 简称ACE
// AclEntry 描述谁 对 何物 有 何种权限
type AclEntry struct {
	ID           int64  `gorm:"column:id"`
	ObjID        string `gorm:"size:100;column:obj_id;index;"` // 对象
	ObjTyp       string `gorm:"size:100;column:obj_typ"`       // Object Type
	Sid          string `gorm:"size:100;column:sid;index;"`    // the user, or role, or any AclObject
	SidTyp       string `gorm:"size:100;column:sid_typ;"`      // indicate user or role or anything else
	Pos          int    // 规则的顺序, 判断权限时, 按照Pos从小到大逐个判断
	Mask         uint32 // permission
	IsGrant      bool   //
	AuditSuccess bool
	AuditFailure bool
	CreatedAt    time.Time
}

// ACESlice 排序用的AclEntry slice
type ACESlice []*AclEntry

// Len length of slice
func (s ACESlice) Len() int {
	return len(s)
}

// Less less function
func (s ACESlice) Less(i, j int) bool {
	if s[i].Pos < s[j].Pos {
		return true
	} else if s[i].Pos == s[j].Pos {
		return s[i].ID < s[j].ID
	}
	return false
}

// Swap swap function
func (s ACESlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// GetPricipalACE 获得主体的所有 ACL entry
func GetPricipalACE(db *gorm.DB, rc redis.Conn, sid string) ([]*AclEntry, error) {
	var (
		aces []*AclEntry
		err  error
	)

	aces, err = getPricipalACEFromRedis(rc, sid)
	if err == nil {
		return aces, nil
	}

	aces, err = getPricipalACEFromDB(db, sid)
	if err != nil {
		glog.Error("getPricipalACEFromDB: get pricipal ACE failed: %v\n", err)
		return nil, err
	}

	setPricipalACEToRedis(rc, sid, aces)
	return aces, nil
}

// getPricipalACEFromRedis get principal ACE from redis
func getPricipalACEFromRedis(rc redis.Conn, sid string) ([]*AclEntry, error) {
	key := KeyAclEntry(sid)
	exist, err := redis.Bool(rc.Do("EXISTS", key))
	if err != nil {
		return nil, err
	}
	if exist == false {
		return nil, ErrRedisKeyNotExist
	}

	vals, err := redis.ByteSlices(rc.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}

	// 从redis 的 hset 中取出的 entry 无顺序区分, 因此, 需要重新排序
	var aces = []*AclEntry{}
	for i := 0; i < len(vals); i += 2 {
		var ace []*AclEntry
		val := vals[i+1]
		err = json.Unmarshal(val, &ace)
		if err != nil {
			glog.Error("getPricipalACEFromRedis: unmarshal AclEntry failed: bytes=%v error=%v\n",
				string(val), err)
			return nil, err
		}

		aces = append(aces, ace...)
	}

	sort.Sort(ACESlice(aces))
	return aces, nil
}

// getPricipalACEFromDB get pricipal ACE from DB
func getPricipalACEFromDB(db *gorm.DB, sid string) ([]*AclEntry, error) {
	var aces []*AclEntry

	err := db.Where("sid=?", sid).Order("pos,id").Find(&aces).Error
	if err == gorm.ErrRecordNotFound {
		return []*AclEntry{}, nil
	}
	return aces, err
}

// setPricipalACEToRedis set pricipal ACE to redis
// 一个objectID 可能对应多条记录, 因此, 首先把同一个objectID的记录合并成一个slice, 然后再写入redis
func setPricipalACEToRedis(rc redis.Conn, sid string, aces []*AclEntry) {
	if len(aces) == 0 {
		return
	}

	key := KeyAclEntry(sid)

	maces := map[string][]*AclEntry{}
	for _, ace := range aces {
		if _, ok := maces[ace.ObjID]; ok {
			maces[ace.ObjID] = append(maces[ace.ObjID], ace)
		} else {
			maces[ace.ObjID] = []*AclEntry{ace}
		}
	}

	for subkey, aceV := range maces {
		val, err := json.Marshal(aceV)
		if err != nil {
			glog.Error("setPricipalACEToRedis: marshal AclEntry %v failed: %v\n", aceV, err)
			rc.Do("HDEL", key)
			return
		}

		// 写入redis, 任何一次写入失败都将导致整个函数失败, 并删除主键
		_, err = rc.Do("HSET", key, subkey, string(val))
		if err != nil {
			glog.Error("setPricipalACEToRedis: HSET %s failed: %v\n", key, err)
			//rc.Do("HDEL", key)
			return
		}
	}
}

///////////////////////////////////获得主体对于Object的权限////////////////////////////////////////

// GetPricipalObjectACE 获得主体对于obj的 ACL entry
func GetPricipalObjectACE(db *gorm.DB, rc redis.Conn, sid, obj string) ([]*AclEntry, error) {
	var (
		aces []*AclEntry
		err  error
	)

	aces, err = getPricipalObjectACEFromRedis(rc, sid, obj)
	if err == nil {
		return aces, nil
	}

	aces, err = getPricipalObjectACEFromDB(db, sid, obj)
	if err != nil {
		glog.Error("getPricipalObjectACEFromDB: get pricipal ACE failed: %v\n", err)
		return nil, err
	}

	setPricipalObjectACEToRedis(rc, sid, aces)
	return aces, nil
}

// getPricipalObjectACEFromRedis get principal ACE from redis
func getPricipalObjectACEFromRedis(rc redis.Conn, sid, obj string) ([]*AclEntry, error) {
	key := KeyAclEntry(sid)
	exist, err := redis.Bool(rc.Do("HEXISTS", key, obj))
	if err != nil {
		return nil, err
	}
	if exist == false {
		return nil, ErrRedisSubkeyNotExist
	}

	vals, err := redis.Bytes(rc.Do("HGET", key, obj))
	if err != nil {
		return nil, err
	}

	// 从redis 的 hset 中取出的 entry 无顺序区分, 因此, 需要重新排序
	var aces []*AclEntry

	err = json.Unmarshal(vals, &aces)
	if err != nil {
		glog.Error("getPricipalObjectACEFromRedis: unmarshal AclEntry failed: bytes=%v error=%v\n",
			string(vals), err)
		return nil, err
	}

	sort.Sort(ACESlice(aces))
	return aces, nil
}

// getPricipalObjectACEFromDB get pricipal ACE from DB
func getPricipalObjectACEFromDB(db *gorm.DB, sid, obj string) ([]*AclEntry, error) {
	var aces []*AclEntry

	err := db.Where("sid=? AND obj_id=?", sid, obj).Order("pos,id").Find(&aces).Error
	if err == gorm.ErrRecordNotFound {
		return []*AclEntry{}, nil
	}
	return aces, err
}

// setPricipalObjectACEToRedis set pricipal ACE to redis
// 一个objectID 可能对应多条记录, 因此, 首先把同一个objectID的记录合并成一个slice, 然后再写入redis
func setPricipalObjectACEToRedis(rc redis.Conn, sid string, aces []*AclEntry) {
	if len(aces) == 0 {
		return
	}

	key := KeyAclEntry(sid)
	val, err := json.Marshal(aces)
	if err != nil {
		glog.Error("setPricipalObjectACEToRedis: marshal AclEntry %v failed: %v\n", aces, err)
		rc.Do("HDEL", key)
		return
	}

	// 写入redis, 任何一次写入失败都将导致整个函数失败, 并删除主键
	_, err = rc.Do("HSET", key, aces[0].ObjID, string(val))
	if err != nil {
		glog.Error("setPricipalObjectACEToRedis: HSET %s failed: %v\n", key, err)
		return
	}

}

//////////////////////////////////数据库增删改查操作//////////////////////////////////////

// CreateACE 创建AclEntry
// 如果已经存在相同的who和what的AclEntry, 则更新该条记录, 但会覆盖audit策略
func (mgr *AclManager) CreateACE(whoSid, whoTyp, whatSid, whatTyp string,
	mask uint32, order int,
	grant, auditSuccess, auditFailure bool) (*AclEntry, error) {
	// 删除缓存
	mgr.EvictPrincipalACECache(whoSid)

	entry, err := mgr.getACE(whoSid, whatSid, grant, order)
	if entry != nil && err == nil {
		// update entry
		entry.Mask |= mask
		entry.AuditSuccess = auditSuccess
		entry.AuditFailure = auditFailure

		err = mgr.db.Save(entry).Error
		return entry, err
	}

	// 创建新的AclEntry
	ace := AclEntry{
		ObjTyp:       whatTyp,
		ObjID:        whatSid,
		Sid:          whoSid,
		SidTyp:       whoTyp,
		Mask:         mask,
		IsGrant:      grant,
		Pos:          order,
		AuditFailure: auditFailure,
		AuditSuccess: auditSuccess,
	}
	err = mgr.db.Create(&ace).Error

	return &ace, err
}

// DeleteACE 根据id删除AclEntry
func (mgr *AclManager) DeleteACE(id int64) {
	mgr.EvictACECache()
	mgr.db.Where("id=?", id).Delete(AclEntry{})
}

// DeleteACEByPrincipal 根据权限的主体删除AclEntry
func (mgr *AclManager) DeleteACEByPrincipal(who string) {
	mgr.EvictPrincipalACECache(who)
	mgr.db.Where("sid=?", who).Delete(AclEntry{})
}

// DeleteACEByTarget 根据权限的目标删除AclEntry
func (mgr *AclManager) DeleteACEByTarget(what string) {
	mgr.EvictACECache()
	mgr.db.Where("obj_id=?", what).Delete(AclEntry{})
}

// GetAllACEs get all AclEntry from DB by who & what
func (mgr *AclManager) GetAllACEs(who, what string) ([]*AclEntry, error) {
	var aces []*AclEntry

	err := mgr.db.Where("obj_id=? AND sid=?", what, who).Order("pos,id").Find(&aces).Error
	if err == gorm.ErrRecordNotFound {
		return []*AclEntry{}, nil
	}
	return aces, err
}

// GetACE 获取AclEntry, 根据 who, what, grant, order 组合条件查询一个AclEntry
func (mgr *AclManager) getACE(who, what string, grant bool, order int) (*AclEntry, error) {
	var entry AclEntry

	err := mgr.db.Where("obj_id=? AND sid=? AND is_grant=? AND pos=?", what, who, grant, order).Find(&entry).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetACEByPrincipal 根据用户或群组获取AclEntry
func (mgr *AclManager) GetACEByPrincipal(who string) ([]*AclEntry, error) {
	var aces []*AclEntry

	err := mgr.db.Where("sid=?", who).Order("pos,id").Find(&aces).Error
	if err == gorm.ErrRecordNotFound {
		return []*AclEntry{}, nil
	}
	return aces, err
}

// GetACEByTarget 根据权限标的获取AclEntry
func (mgr *AclManager) GetACEByTarget(what string) ([]*AclEntry, error) {
	var aces []*AclEntry

	err := mgr.db.Where("obj_id=?", what).Order("pos,id").Find(&aces).Error
	if err == gorm.ErrRecordNotFound {
		return []*AclEntry{}, nil
	}
	return aces, err
}

// UpdateACE 更新AclEntry
// 可以更新:
//     Mask
//     Order
//     Grant
//     AuditSuccess
//     AuditFailure
func (mgr *AclManager) UpdateACE(id int64, opt map[string]interface{}) error {
	ace := AclEntry{
		ID: id,
	}
	if err := mgr.db.Find(&ace).Error; err != nil {
		glog.Error("UpdateACE: Not Found AclEntry by id %d\n", id)
		return err
	}

	// 删除 AclEntry cache
	mgr.EvictACECache()

	if opt["mask"] != nil {
		ace.Mask = uint32(goutils.ToInt(opt["mask"], int(ace.Mask)))
	}
	if opt["pos"] != nil {
		ace.Pos = goutils.ToInt(opt["pos"], ace.Pos)
	}
	if opt["isGrant"] != nil {
		ace.IsGrant = goutils.ToBool(opt["isGrant"], ace.IsGrant)
	}
	if opt["auditSuccess"] != nil {
		ace.AuditSuccess = goutils.ToBool(opt["auditSuccess"], ace.AuditSuccess)
	}
	if opt["auditFailure"] != nil {
		ace.AuditFailure = goutils.ToBool(opt["auditFailure"], ace.AuditFailure)
	}

	return mgr.db.Save(&ace).Error
}
