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
	return AT
}

func TestRole(t *testing.T) {
	var (
		users = []user{
			{"a"},
			{"b"},
			{"c"},
			{"d"},
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
			{users[0], roleCases[0].Sid, []string{}},
			{users[1], roleCases[1].Sid, []string{}},
			{users[1], roleCases[3].Sid, []string{}},
			{users[1], roleCases[2].Sid, []string{}},
			{users[2], roleCases[4].Sid, []string{}},
		}
	)

	db, err := openDB()
	assert.Assert(err == nil, "opendb")
	pool, err := openRedis()
	assert.Assert(err == nil, "openredis")

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
	}
}

func roleExpect(t *testing.T, roles []Principal, expect []string, idx int) {
	assert.Assertf(len(roles) == len(expect),
		"index: %d role expect length NOT equal: %d expect=%d",
		idx, len(roles), len(expect))

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
