package acl

import (
	"testing"

	"sort"

	"reflect"

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

type user struct {
	name string
}

func (u user) GetSid() string {
	return u.name
}

func (u user) GetTyp() string {
	return AccountTyp
}

func TestRole(t *testing.T) {
	var (
		users = []user{
			{"a"}, // 0
			{"b"}, // 1
			{"c"}, // 2
			{"d"}, // 3
		}
		roleCases = []*Role{
			&Role{Name: "admin", Sid: "admin"},         // 0
			&Role{Name: "user", Sid: "user"},           // 1
			&Role{Name: "root", Sid: "root"},           // 2
			&Role{Name: "hallAdmin", Sid: "hallAdmin"}, // 3
			&Role{Name: "test", Sid: "test"},           // 4
		}
		urCases = []struct {
			u      user
			rid    string
			expect []string
		}{
			{users[0], roleCases[0].Sid, []string{"admin", "a"}},
			{users[1], roleCases[1].Sid, []string{"user", "b"}},
			{users[1], roleCases[3].Sid, []string{"user", "hallAdmin", "b"}},
			{users[1], roleCases[2].Sid, []string{"user", "hallAdmin", "root", "b"}},
			{users[2], roleCases[4].Sid, []string{"test", "c"}},
		}
	)

	db, err := openDB()
	assert.Assert(err == nil, "opendb")
	pool, err := openRedis()
	assert.Assert(err == nil, "openredis")

	// clean data
	cleanRoles(db, pool)

	mgr := CreateAclManager(db, pool)
	for _, r := range roleCases {
		_, err := mgr.CreateRole(r.Name, r.Sid)
		assert.Assert(err == nil, "create role")
	}

	for idx, ur := range urCases {
		err := mgr.AddUserRoleRelation(ur.u.name, ur.rid)
		assert.Assertf(err == nil, "AddUserRole")

		res := GetPrincipals(db, pool.Get(), ur.u)
		roleExpect(t, res, ur.expect, idx)

		// 从缓存中取
		res, err = getUserRolesFromRedis(pool.Get(), ur.u)
		assert.Assertf(err == nil, "getUserRolesFromRedis should success: %v", err)
		roleExpect(t, res, ur.expect, idx)
	}
}

func roleExpect(t *testing.T, roles []Principal, expect []string, idx int) {
	assert.Assertf(len(roles) == len(expect),
		"index: %d role expect length NOT equal: %d expect=%d roles=%v expect=%v",
		idx, len(roles), len(expect), roles, expect)

	sort.Sort(sort.StringSlice(expect))
	rs := []string{}
	for _, r := range roles {
		rs = append(rs, r.Sid)
	}
	sort.Sort(sort.StringSlice(rs))

	if reflect.DeepEqual(rs, expect) == false {
		t.Fatalf("roles NOT equal with expect: idx=%d roles=%v expect=%v",
			idx, rs, expect)
	}
}
