package acl

import "sync"

// 根据规则判断一个用户/角色对某个物体有何种权限
// acl在很多时候

// RuleFn 权限规则判断函数
// who:  谁, 可以是user或role
// what: 权限target
// perm: 权限
type RuleFn func(who interface{}, what interface{}, perm Permission) bool

// Rule rule
type Rule struct {
	fn    RuleFn
	order int    // 按照从小到大的顺序排列
	name  string // rule的标示, 两个rule name相同则认为两个rule相同
}

// Rules rules
type Rules []*Rule

var (
	rulesMap = struct {
		sync.Mutex
		rules map[string]Rules
	}{
		rules: map[string]Rules{},
	}
)

// RegisterRule 注册rule
func RegisterRule(who string, r Rule) error {
	return nil
}

// GetRules get rules
// 复制一份[]*Rule
func GetRules(who string) []*Rule {

	return []*Rule{}
}

// Name rule name
func (r *Rule) Name() string {
	return r.name
}

// Order rule order
func (r *Rule) Order() int {
	return r.order
}

func insertRule(rs Rules, r *Rule) Rules {
	return rs
}

func deleteRule(rs Rules, name string) Rules {
	return rs
}
