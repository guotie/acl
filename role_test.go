package acl

import (
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/guotie/assert"
	"github.com/jinzhu/gorm"
)

// 删除 数据库 和 redis 中的数据
func cleanRoles(db *gorm.DB, pool *redis.Pool) {
	db.Exec("Truncate roles")
	db.Exec("Truncate user_roles")
	pool.Get().Do("flushdb")
}

func TestRole(t *testing.T) {
	var (
		users = []struct {
			name string
		}{
			{"a"},
			{"b"},
			{"c"},
			{"d"},
		}
		roleCases = []struct {
			name string
			sid  string
		}{
			{"admin", "admin"},
			{"user", "user"},
			{"root", "root"},
			{"hallAdmin", ""},
		}
		urCases = []struct {
			uid    string
			rid    string
			expect []interface{}
		}{
			{},
		}
	)

	db, err := openDB()
	assert.Assert(err == nil, "opendb")
	pool, err := openRedis()
	assert.Assert(err == nil, "openredis")

	mgr := CreateAclManager(db, pool)
	for _, r := range roleCases {
		_, err := mgr.CreateRole(r.name, r.sid)
		assert.Assert(err == nil, "create role")
	}

	for _, ur := range urCases {
		err := mgr.AddUserRoleRelation(ur.uid, ur.rid)
		assert.Assertf(err == nil, "AddUserRole")
	}
}

func roleExpect(t *testing.T, roles []Principal, expect []interface{}, idx int) {
	assert.Assertf(len(roles) == len(expect),
		"index: %d role expect length NOT equal: %d expect=%d", idx, len(roles), len(expect))
}
