package acl

import (
	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/smtc/glog"
)

// Principal 主体
// Pricipal 可以是一个用户或者一个role, 是权限主体的最小单位
//type Principal struct {
//	Sid string
//	Typ string
//}
type Principal interface {
	AclObject
}

// isGrant principal 是否被授权或拒绝
// 没有策略: 返回0
// 拒绝:     返回Reject
// 允许:     返回Grant
//
// 0. 从rule中判定是否被允许或拒绝
// 1. 从acl中判定是否被允许或拒绝
//
func isGrant(db *gorm.DB, rc redis.Conn, sid Principal, obj AclObject, perm Permission) int {
	// 通过rules来判定策略
	if ret := JudgePolicyByRules(sid, obj, perm); ret != 0 {
		//fmt.Printf("  JudgePolicyByRules: who: %v what: %v perm: %v result: %v\n", sid, obj, perm, ret)
		return ret
	}

	// 根据acls来判定策略
	if ret := JudgePolicyByAcls(db, rc, sid, obj, perm); ret != 0 {
		//fmt.Printf("  JudgePolicyByAcls: who: %v what: %v perm: %v result: %v\n", sid, obj, perm, ret)
		return ret
	}

	return 0
}

// JudgePolicyByRules 根据rules来判定策略
// Grant:  授权
// Reject: 拒绝
// 0:      无法判断
func JudgePolicyByRules(sid Principal, obj AclObject, perm Permission) int {
	// rules 的规则
	rules := GetRules(sid.GetSid())
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

// JudgePolicyByAcls 根据acls来判定策略
// Grant:  授权
// Reject: 拒绝
// 0:      无法判断
func JudgePolicyByAcls(db *gorm.DB, rc redis.Conn, sid Principal, obj AclObject, perm Permission) int {
	// acl的规则
	aces, err := GetPricipalACE(db, rc, sid.GetSid())
	if err != nil {
		glog.Error("JudgePolicyByAcls: GetPricipalACE failed: error=%v\n", err)
		return 0
	}

	for _, ace := range aces {
		// 策略不是作用于obj
		if ace.ObjID != obj.GetSid() {
			continue
		}

		if ace.IsGrant == false {
			return Reject
		}
		// ace.Grant == true 的情况
		if ace.Mask&perm.Mask == perm.Mask {
			return Grant
		}
	}

	return 0
}
