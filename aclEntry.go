package acl

import (
	"encoding/json"
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/smtc/glog"
)

// AclEntry Acl Entry, 简称ACE
// AclEntry 描述谁 对 何物 有 何种权限
type AclEntry struct {
	ID           int64  `gorm:"column:id"`
	ObjID        string `gorm:"size:100;column:obj_id;unique_index;"` // 对象
	Sid          string `gorm:"size:100;column:sid;unique_index;"`    // the user, or role, or any AclObject
	SidTyp       string `gorm:"size:100;column:sid_typ;"`             // indicate user or role or anything else
	Order        int    // 规则的顺序, 判断权限时, 按照order从小到大逐个判断
	Grant        bool   //
	AuditSuccess bool
	AuditFailure bool
}

// GetPricipalACE 获得主体的ACL entry
func GetPricipalACE(db *gorm.DB, rc redis.Conn, sid string) []*AclEntry {
	var (
		aces []*AclEntry
		err  error
	)

	aces, err = getPricipalACEFromRedis(rc, sid)
	if err == nil {
		return aces
	}

	aces, err = getPricipalACEFromDB(db, sid)
	if err != nil {
		glog.Error("getPricipalACEFromDB: get pricipal ACE failed: %v\n", err)
		return aces
	}

	setPricipalACEToRedis(rc, sid, aces)
	return aces
}

// getPricipalACEFromRedis get principal ACE from redis
func getPricipalACEFromRedis(rc redis.Conn, sid string) ([]*AclEntry, error) {
	key := KeyAclEntry(sid)
	exist, err := redis.Bool(rc.Do("EXISTS", key))
	if err != nil {
		return nil, err
	}
	if exist == false {
		return nil, fmt.Errorf("getPricipalACEFromRedis: key %s NOT exist", key)
	}

	vals, err := redis.ByteSlices(rc.Do("HGETALL", key))
	if err != nil {

	}

	var aces = []*AclEntry{}
	for _, val := range vals {
		var ace AclEntry
		err = json.Unmarshal(val, &ace)
		if err != nil {
			glog.Error("getPricipalACEFromRedis: unmarshal AclEntry failed: bytes=%v error=%v\n",
				val, err)
			return nil, err
		}
		aces = append(aces, &ace)
	}

	return aces, nil
}

// getPricipalACEFromDB get pricipal ACE from DB
func getPricipalACEFromDB(db *gorm.DB, sid string) ([]*AclEntry, error) {
	var aces []*AclEntry

	err := db.Where("sid=?", sid).Find(&aces).Error
	return aces, err
}

// setPricipalACEToRedis set pricipal ACE to redis
func setPricipalACEToRedis(rc redis.Conn, sid string, aces []*AclEntry) {
	key := KeyAclEntry(sid)

	for _, ace := range aces {
		val, err := json.Marshal(ace)
		if err != nil {
			glog.Error("setPricipalACEToRedis: marshal AclEntry %v failed: %v\n", ace, err)
			rc.Do("HDEL", key)
			return
		}

		_, err = rc.Do("HSET", redis.Args{}.Add(key).Add(ace.ObjID).Add(val))
		if err != nil {
			rc.Do("HDEL", key)
			return
		}
	}
}
