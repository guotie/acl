package acl

import "testing"

// business hall 营业厅
type Hall struct {
	Sid  string
	City string
}

func (h Hall) GetSid() string { return h.Sid }
func (h Hall) GetTyp() string { return "HallTyp" }

var (
	AT = "AccountTyp"

	objs = []Hall{
		{"南京营业厅01", "南京"},
		{"南京营业厅02", "南京"},
		{"南京营业厅03", "南京"},
		{"苏州营业厅01", "苏州"},
		{"镇江营业厅01", "镇江"},
		{"无锡营业厅01", "无锡"},
	}

	rs = Rules{
		// 0
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 10, "10a"},
		// 1
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 7, "7a"},
		// 2
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 3, "3a"},
		// 3
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 5, "5a"},
		// 4
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 7, "7b"},
		// 5
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 10, "10b"},
		// 6
		&Rule{func(who Principal, what AclObject, perm Permission) int {
			if who.Sid == "user01" {
				return 1
			}
			return 0
		}, 1, "1c"},

		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 1, "1a"},
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 1 }, 1, "1b"},
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 100, "100a"},
		&Rule{func(who Principal, what AclObject, perm Permission) int { return -1 }, 1, "1d"},
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 100, "100b"},
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 3, "3b"},
		&Rule{func(who Principal, what AclObject, perm Permission) int { return 0 }, 5, "5b"},
	}

	ps = []Principal{
		Principal{"user01", AT}, // 超级管理员
		Principal{"user02", AT}, // 管理员
		Principal{"user03", AT}, // 普通人员
		Principal{"user04", AT}, // 普通人员
		Principal{"user05", AT}, // 无权限人员
		Principal{"prov01", RoleTyp},
		Principal{"prov02", RoleTyp},
		Principal{"city01", RoleTyp},
		Principal{"city02", RoleTyp},
	}

	testRules = []struct {
		p     Principal
		rules Rules
	}{
		{ps[0], Rules{rs[6]}},
	}
)

// TestRules insert rule, delete rule, get rules
func TestRules(t *testing.T) {
	/*
		for _, r := range rs {
			RegisterRule(ps[0].Sid, *r)
		}

		rules := GetRules(ps[0].Sid)
		assert.Assertf(len(rules) == len(rs), "user01 rules count wrong: should be %d, but %d", len(rs), len(rules))
		for i := 1; i < len(rules); i++ {
			assert.Assertf(rules[i].order >= rules[i-1].order, "rules order wrong")
		}
		assert.Assertf(rules[0].name == "1c", "rules 0 name wrong, should be 1c, but %s", rules[0].name)
		assert.Assert(rules[1].name == "1a", "rules 1 name wrong")
		assert.Assert(rules[2].name == "1b", "rules 2 name wrong")
		assert.Assert(rules[3].name == "1d", "rules 3 name wrong")
		assert.Assert(rules[4].name == "3a", "rules 4 name wrong")
		assert.Assert(rules[5].name == "3b", "rules 5 name wrong")
		assert.Assert(rules[6].name == "5a", "rules 6 name wrong")
		assert.Assert(rules[7].name == "5b", "rules 7 name wrong")
		assert.Assert(rules[8].name == "7a", "rules 8 name wrong")
		assert.Assert(rules[9].name == "7b", "rules 9 name wrong")

		for i := 1; i < len(ps); i++ {
			rules := GetRules(ps[i].Sid)
			assert.Assert(len(rules) == 0, "rules should be empty")
		}
	*/
}
