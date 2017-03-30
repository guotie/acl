package acl

import "sync"
import "fmt"

// 根据规则判断一个用户/角色对某个物体有何种权限
// acl在很多时候

// RuleFn 权限规则判断函数
// who:  谁, 可以是user或role
// what: 权限target
// perm: 权限
// 返回:  1: 授权
//       -1: 拒绝
//        0: 无法判断
type RuleFn func(who Principal, what AclObject, perm Permission) int

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
	var (
		rs Rules
		ok bool
	)

	rulesMap.Lock()
	defer rulesMap.Unlock()

	rs, ok = rulesMap.rules[who]
	if !ok {
		rulesMap.rules[who] = Rules{&r}
		return nil
	}

	rulesMap.rules[who] = insertRule(rs, &r)
	return nil
}

// UnregisterRule 删除rule
func UnregisterRule(who string, name string) {
	rulesMap.Lock()
	defer rulesMap.Unlock()

	if rs, ok := rulesMap.rules[who]; ok {
		rulesMap.rules[who] = deleteRule(rs, name)
	}

	return
}

// GetRules get rules
// 复制一份[]*Rule
func GetRules(who string) Rules {
	var (
		rs Rules
		ok bool
	)

	rulesMap.Lock()
	defer rulesMap.Unlock()

	rs, ok = rulesMap.rules[who]
	if !ok {
		return Rules{}
	}
	res := make(Rules, len(rs))
	copy(res, rs)
	return res
}

// Name rule name
func (r *Rule) Name() string {
	return r.name
}

// Order rule order
func (r *Rule) Order() int {
	return r.order
}

// insertRule insert rule to Rules, 按照order由低到高的顺序, 返回新的rules
func insertRule(rs Rules, r *Rule) Rules {
	nrs := make(Rules, len(rs)+1)

	idx := 0
	inserted := false
	for _, ele := range rs {
		// 校验name, rule的name不能出现重复
		if r.Name() == ele.Name() {
			// 名字重复, panic
			panic(fmt.Sprintf("insertRule: rule name %s duplicated", r.Name()))
			//return rs
		}

		if inserted || ele.Order() <= r.Order() {
			nrs[idx] = ele
			idx++
			continue
		}

		// 将待增加的rule加入到Rules中
		nrs[idx] = r
		idx++
		nrs[idx] = ele
		idx++
		inserted = true
	}
	if idx == len(rs) {
		// 待增加的rule位于slice末尾的情况
		nrs[idx] = r
	}
	return nrs
}

// deleteRule 从rules中删除特定的rule, 返回新的rules
func deleteRule(rs Rules, name string) Rules {
	idx := 0
	nrs := make(Rules, len(rs))

	for _, ele := range rs {
		if ele.Name() != name {
			nrs[idx] = ele
			idx++
		}
	}

	return nrs[0:idx]
}
