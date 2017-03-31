package acl

import (
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/guotie/assert"
	"github.com/jinzhu/gorm"
)

// 营业厅
type hall struct {
	name string
	corp string
	city string
	prov string
}

// GetSid get sid
func (h hall) GetSid() string {
	return h.name
}

// GetTyp get hall type
func (h hall) GetTyp() string {
	return "hall"
}

type admin struct {
	name string
	corp string
	city string
	prov string
}

func (a admin) GetSid() string {
	return a.name
}

func (a admin) GetTyp() string {
	return AccountTyp
}

// 构造数据
var (
	now = time.Now()
	pr  = Permission{PermissionRead, "read"}
	pw  = Permission{PermissionWrite, "write"}
	pc  = Permission{PermissionCreate, "create"}
	pd  = Permission{PermissionDelete, "delete"}

	users = []admin{
		{"联通南京营业厅1管理员", "unicom", "nanjing", "jiangsu"},  // 0
		{"联通南京营业厅1普通账号", "unicom", "nanjing", "jiangsu"}, // 1
		{"联通南京营业厅2管理员", "unicom", "nanjing", "jiangsu"},  // 2
		{"联通南京营业厅2普通账号", "unicom", "nanjing", "jiangsu"}, // 3
		{"联通南京市公司管理员", "unicom", "nanjing", "jiangsu"},   // 4
		{"联通江苏省公司管理员", "unicom", "nanjing", "jiangsu"},   // 5
		{"联通北京公司管理员", "unicom", "nanjing", "jiangsu"},    // 6
		{"工行北京网点管理员", "icbc", "nanjing", "jiangsu"},      // 7
		{"工行南京网点普通人员", "icbc", "nanjing", "jiangsu"},     // 8
		{"noname", "", "nanjing", "jiangsu"},             // 9
	}

	objs = []hall{
		{"联通南京营业厅1", "unicom", "nanjing", "jiangsu"}, // 0
		{"工行南京营业厅1", "icbc", "nanjing", "jiangsu"},   // 1
		{"联通南京营业厅2", "unicom", "nanjing", "jiangsu"}, // 2
		{"联通镇江营业厅1", "unicom", "镇江", "jiangsu"},      // 3
		{"联通北京营业厅1", "unicom", "beijing", "beijing"}, // 4
		{"工行北京营业厅1", "icbc", "beijing", "beijing"},   // 5
	}

	// 测试验证使用的用例及期望输出
	permCases = []struct {
		u      admin
		h      hall
		perm   uint32
		expect bool
	}{
		// 0 联通南京营业厅1管理员
		{users[0], objs[0], 0xf, true},
		{users[0], objs[1], 0x1, false},
		{users[0], objs[1], 0xf, false},
		{users[0], objs[2], 0xf, false},
		{users[0], objs[3], 0xf, false},
		{users[0], objs[4], 0xf, false},
		{users[0], objs[5], 0x1, false},

		// 1 联通南京营业厅1普通账号
		{users[1], objs[0], 0x1, true},
		{users[1], objs[0], 0x2, false},
		{users[1], objs[0], 0x3, false},
		{users[1], objs[1], 0x1, false},
		{users[1], objs[1], 0xf, false},
		{users[1], objs[2], 0xf, false},
		{users[1], objs[3], 0xf, false},
		{users[1], objs[4], 0x1, false},
		{users[1], objs[5], 0x1, false},

		// 2 联通南京营业厅2管理员
		{users[2], objs[0], 0x1, false},
		{users[2], objs[0], 0xf, false},
		{users[2], objs[1], 0x1, false},
		{users[2], objs[1], 0xf, false},
		{users[2], objs[2], 0xf, true},
		{users[2], objs[3], 0xf, false},
		{users[2], objs[4], 0x1, false},
		{users[2], objs[5], 0x1, false},

		// 3 联通南京营业厅2普通账号
		{users[3], objs[0], 0x1, false},
		{users[3], objs[0], 0xf, false},
		{users[3], objs[1], 0x1, false},
		{users[3], objs[1], 0xf, false},
		{users[3], objs[2], 0xf, false},
		{users[3], objs[2], 0x1, false},
		{users[3], objs[2], 0x3, false},
		{users[3], objs[2], 0x4, false},
		{users[3], objs[2], 0x2, true},
		{users[3], objs[2], 0x3, false},
		{users[3], objs[3], 0xf, false},
		{users[3], objs[4], 0x1, false},
		{users[3], objs[5], 0x1, false},

		// 4 联通南京市公司管理员
		{users[4], objs[0], 0x1, true},
		{users[4], objs[0], 0xf, true},
		{users[4], objs[1], 0x1, false},
		{users[4], objs[1], 0xf, false},
		{users[4], objs[2], 0xf, true},
		{users[4], objs[2], 0x3, true},
		{users[4], objs[2], 0x4, true},
		{users[4], objs[2], 0x8, true},
		{users[4], objs[3], 0xf, false},
		{users[4], objs[4], 0x1, false},
		{users[4], objs[5], 0x1, false},

		// 5 联通江苏省公司管理员
		{users[5], objs[0], 0x1, true},
		{users[5], objs[0], 0xf, true},
		{users[5], objs[1], 0x1, false},
		{users[5], objs[1], 0xf, false},
		{users[5], objs[2], 0xf, true},
		{users[5], objs[3], 0xf, true},
		{users[5], objs[2], 0x1, true},
		{users[5], objs[3], 0x2, true},
		{users[5], objs[2], 0x3, true},
		{users[5], objs[3], 0x8, true},
		{users[5], objs[2], 0x7, true},
		{users[5], objs[3], 0x9, true},
		{users[5], objs[4], 0x1, false},
		{users[5], objs[5], 0x1, false},

		// 6 联通北京公司管理员
		{users[6], objs[0], 0x1, false},
		{users[6], objs[0], 0xf, false},
		{users[6], objs[1], 0x1, false},
		{users[6], objs[1], 0xf, false},
		{users[6], objs[2], 0xf, false},
		{users[6], objs[3], 0xf, false},
		{users[6], objs[4], 0x1, true},
		{users[6], objs[4], 0x3, true},
		{users[6], objs[4], 0x4, true},
		{users[6], objs[4], 0xf, true},
		{users[6], objs[5], 0x1, false},

		// 7 工行北京网点管理员
		{users[7], objs[0], 0x1, false},
		{users[7], objs[0], 0xf, false},
		{users[7], objs[1], 0x1, false},
		{users[7], objs[1], 0xf, false},
		{users[7], objs[2], 0xf, false},
		{users[7], objs[3], 0xf, false},
		{users[7], objs[4], 0x1, false},
		{users[7], objs[5], 0x1, true},
		{users[7], objs[5], 0xf, true},
		{users[7], objs[5], 0x7, true},
		{users[7], objs[5], 0x4, true},

		// 8 工行南京网点普通人员
		{users[8], objs[0], 0x1, false},
		{users[8], objs[0], 0xf, false},
		{users[8], objs[1], 0x1, true},
		{users[8], objs[1], 0xf, false},
		{users[8], objs[2], 0xf, false},
		{users[8], objs[3], 0xf, false},
		{users[8], objs[4], 0x1, false},
		{users[8], objs[5], 0x1, false},
		{users[8], objs[0], 0xf, false},
		{users[8], objs[0], 0xf, false},

		// 9 noname
		{users[9], objs[0], 0x1, false},
		{users[9], objs[0], 0xf, false},
		{users[9], objs[1], 0x1, false},
		{users[9], objs[1], 0xf, false},
		{users[9], objs[2], 0xf, false},
		{users[9], objs[3], 0xf, false},
		{users[9], objs[4], 0x1, false},
		{users[9], objs[5], 0x1, false},
		{users[9], objs[0], 0x1, false},
		{users[9], objs[0], 0x3, false},
	}

	roles = []*Role{
		&Role{Name: "联通南京市管理员", Sid: "联通南京市管理员", Corp: "unicom", City: "nanjing", Province: "jiangsu", Level: 10}, // 0
		&Role{Name: "联通江苏省管理员", Sid: "联通江苏省管理员", Corp: "unicom", City: "nanjing", Province: "jiangsu", Level: 11}, // 1
		&Role{Name: "联通北京市管理员", Sid: "联通北京市管理员", Corp: "unicom", City: "beijing", Province: "beijing", Level: 10}, // 2
		&Role{Name: "工行北京市管理员", Sid: "工行北京市管理员", Corp: "icbc", City: "beijing", Province: "beijing", Level: 10},   // 3
		&Role{Name: "工行南京市管理员", Sid: "工行南京市管理员", Corp: "icbc", City: "nanjing", Province: "jiangsu", Level: 10},   // 4
		&Role{Name: "test", Sid: "test"}, // 4
	}

	urs = []*UserRole{
		&UserRole{UID: users[4].name, Rid: roles[0].Sid},
		&UserRole{UID: users[5].name, Rid: roles[1].Sid},
		&UserRole{UID: users[6].name, Rid: roles[2].Sid},
		&UserRole{UID: users[7].name, Rid: roles[3].Sid},
	}

	fns = Rule{
		name:  "rule",
		order: 1,

		fn: func(who Principal, what AclObject, perm Permission) int {
			if who.GetTyp() != RoleTyp {
				return 0
			}

			r, ok := who.(*Role)
			if !ok {
				return 0
			}
			h, ok := what.(hall)
			if !ok {
				return 0
			}

			// 公司不同
			if h.corp != r.Corp {
				return 0
			}
			if h.prov != r.Province {
				return 0
			}
			if h.city != r.City {
				if r.Level <= 10 {
					return 0
				}
				return 1
			}
			// 同公司 同城市 同省份
			if r.Level >= 10 {
				return 1
			}

			return 0
		},
	}

	acls = []AclEntry{
		{0, objs[0].GetSid(), objs[0].GetTyp(), users[0].GetSid(), users[0].GetTyp(), 1, 0xff, true, false, false, now}, // 1 所有权限
		{0, objs[0].GetSid(), objs[0].GetTyp(), users[1].GetSid(), users[1].GetTyp(), 6, 0x1, true, false, false, now},  // 2 读

		// 工行南京网点普通人员
		{0, objs[1].GetSid(), objs[1].GetTyp(), users[8].GetSid(), users[8].GetTyp(), 1, 0x1, true, false, false, now}, // 3 读写权限

		{0, objs[2].GetSid(), objs[2].GetTyp(), users[2].GetSid(), users[2].GetTyp(), 1, 0xff, true, false, false, now}, //
		{0, objs[2].GetSid(), objs[2].GetTyp(), users[3].GetSid(), users[3].GetTyp(), 1, 0x2, true, false, false, now},  //

		// 联通北京营业厅
		{0, objs[4].GetSid(), objs[4].GetTyp(), users[6].GetSid(), users[6].GetTyp(), 50, 0xf, true, false, false, now}, // 4
		// 工行北京营业厅
		{0, objs[5].GetSid(), objs[5].GetTyp(), users[7].GetSid(), users[7].GetTyp(), 10, 0xf, true, false, false, now}, // 5
	}
)

func cleanAllTable(db *gorm.DB, rc redis.Conn) {
	db.Exec("truncate roles")
	db.Exec("truncate acl_entries")
	db.Exec("truncate user_roles")

	rc.Do("flushdb")
}

// test acl & rules in black box mode
func TestACL(t *testing.T) {
	db, err := openDB()
	assert.Assert(err == nil, "opendb")
	pool, err := openRedis()
	assert.Assert(err == nil, "openredis")

	cleanAllTable(db, pool.Get())

	mgr := CreateAclManager(db, pool)
	for _, r := range roles {
		_, err := mgr.CreateRole(r.Name, r.Sid, r.Depart, r.Corp, r.City, r.Province, r.Level)
		assert.Assertf(err == nil, "create role failed: %v", err)
		if r.Name != "test" {
			RegisterRule(r.Sid, fns)
		}
	}

	for _, ur := range urs {
		err := mgr.AddUserRoleRelation(ur.UID, ur.Rid)
		assert.Assertf(err == nil, "create user-role-relation failed: %v", err)
	}

	for _, acl := range acls {
		_, err := mgr.CreateACE(acl.Sid, acl.SidTyp, acl.ObjID, acl.ObjTyp, acl.Mask, acl.Pos, acl.IsGrant, acl.AuditSuccess, acl.AuditFailure)
		assert.Assertf(err == nil, "createACE failed: %v", err)
	}

	for idx, c := range permCases {
		result := mgr.IsGrant(c.u, c.h, Permission{c.perm, ""})
		assert.Assertf(result == c.expect, "index=%d user %v on object %v perm %x should be %v, but %v",
			idx, c.u, c.h, c.perm, c.expect, result)
	}

	for idx, c := range permCases {
		result := mgr.IsGrant(c.u, c.h, Permission{c.perm, ""})
		assert.Assertf(result == c.expect, "index=%d user %v on object %v perm %x should be %v, but %v",
			1000+idx, c.u, c.h, c.perm, c.expect, result)
	}
}
