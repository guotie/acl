package acl

import (
	"fmt"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/guotie/assert"
	"github.com/jinzhu/gorm"
)

var (
	dbuser = "root"
	dbpass = "guotie.action"
	dbhost = "localhost"
	dbport = "3306"
	dbname = "acl"
)

// openRedis open redis connection
func openRedis() (*redis.Pool, error) {
	var proto, addr string

	proto = "tcp"
	addr = "127.0.0.1:6379"

	rpool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 600 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(proto, addr)
			if err != nil {
				return nil, err
			}

			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	conn := rpool.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	if err != nil {
		return nil, err
	}

	return rpool, nil
}

// openDB open database
func openDB() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Asia%%2FShanghai&timeout=31536000s",
		dbuser, dbpass, dbhost, dbport, dbname))
	if err != nil {
		return nil, err
	}

	return db, nil
}

// 删除 数据库 和 redis 中的数据
func cleanAclEntry(db *gorm.DB, pool *redis.Pool) {
	db.Exec("Truncate acl_entries")
	//pool.Get().Do("flushdb")
}

//
func testAclEntry(t *testing.T) {
	db, err := openDB()
	assert.Assertf(err == nil, "open db should success: %v", err)
	rp, err := openRedis()
	assert.Assertf(err == nil, "open redis should success: %v", err)

	testGetPricipalAce(t, db, rp)
}

// testGetPrincipalAce 测试 GetPrincipalAce
func testGetPricipalAce(t *testing.T, db *gorm.DB, pool *redis.Pool) {
	var (
		now = time.Now()
		// 三个Principal 00a 00b 00c
		// 5个营业厅 001 002 003 004
		aces = []AclEntry{
			{0, "hall-001", "hall", "00a", AT, 1, 0xff, true, false, false, now}, // 1 所有权限
			{0, "hall-001", "hall", "00b", AT, 6, 0x4, true, false, false, now},  // 2 写权限
			{0, "hall-001", "hall", "00c", AT, 1, 0x3, true, false, false, now},  // 3 读写权限
			{0, "hall-001", "hall", "00a", AT, 1, 0x4, true, false, false, now},  //   合并1 不会插入到数据库中
			{0, "hall-001", "hall", "00a", AT, 1, 0x2, true, false, false, now},  //   合并1 不会插入到数据库中

			{0, "hall-002", "hall", "00a", AT, 50, 0x3, true, false, false, now}, // 4
			{0, "hall-002", "hall", "00b", AT, 10, 0x1, true, false, false, now}, // 5
			{0, "hall-003", "hall", "00a", AT, 20, 0x7, true, false, false, now}, // 6
			{0, "hall-004", "hall", "00c", AT, 30, 0xf, true, false, false, now}, // 7

			{0, "hall-002", "hall", "00b", AT, 10, 0x2, true, false, false, now}, //  写权限, 合并5
			{0, "hall-003", "hall", "00b", AT, 10, 0x2, true, false, false, now}, // 8 写权限

			{0, "hall-004", "hall", "00a", AT, 40, 0x4, true, false, false, now}, // 9

			{0, "hall-001", "hall", "00b", AT, 1, 0x1, true, false, false, now}, //  读权限
		}

		getCases = []struct {
			fn     string
			params []string
			result [][]interface{}
		}{
			{"GetAllACEs", []string{"00a", "hall-001"}, [][]interface{}{
				[]interface{}{"hall-001", "hall", "00a", AT, 1, 0xff, true, false, false},
			},
			},
			{"GetAllACEs", []string{"00b", "hall-001"}, [][]interface{}{
				[]interface{}{"hall-001", "hall", "00b", AT, 1, 0x1, true, false, false},
				[]interface{}{"hall-001", "hall", "00b", AT, 6, 0x4, true, false, false},
			}},

			{"GetACEByPrincipal", []string{"00b"}, [][]interface{}{
				[]interface{}{"hall-001", "hall", "00b", AT, 1, 0x1, true, false, false},
				[]interface{}{"hall-001", "hall", "00b", AT, 6, 0x4, true, false, false},
				[]interface{}{"hall-002", "hall", "00b", AT, 10, 0x3, true, false, false},
				[]interface{}{"hall-003", "hall", "00b", AT, 10, 0x2, true, false, false},
			}},
			{"GetACEByPrincipal", []string{"00a"}, [][]interface{}{
				[]interface{}{"hall-001", "hall", "00a", AT, 1, 0xff, true, false, false},
				[]interface{}{"hall-003", "hall", "00a", AT, 20, 0x7, true, false, false},
				[]interface{}{"hall-004", "hall", "00a", AT, 40, 0x4, true, false, false},
				[]interface{}{"hall-002", "hall", "00a", AT, 50, 0x3, true, false, false},
			}},

			{"GetACEByTarget", []string{"hall-002"}, [][]interface{}{
				[]interface{}{"hall-002", "hall", "00b", AT, 10, 0x3, true, false, false},
				[]interface{}{"hall-002", "hall", "00a", AT, 50, 0x3, true, false, false},
			}},
			{"GetACEByTarget", []string{"hall-003"}, [][]interface{}{
				[]interface{}{"hall-003", "hall", "00b", AT, 10, 0x2, true, false, false},
				[]interface{}{"hall-003", "hall", "00a", AT, 20, 0x7, true, false, false},
			}},
			{"GetACEByTarget", []string{"hall-004"}, [][]interface{}{
				[]interface{}{"hall-004", "hall", "00c", AT, 30, 0xf, true, false, false},
				[]interface{}{"hall-004", "hall", "00a", AT, 40, 0x4, true, false, false},
			}},
		}

		// update cases 7, 9
		updateCases = []struct {
			idx    int
			opt    map[string]interface{}
			fn     string
			params []string
			result [][]interface{}
		}{
			// {0, "hall-004", "hall", "00c", AT, 30, 0xf, true, false, false, now}, // 7
			{7, map[string]interface{}{
				"mask": 7,
				"pos":  1000,
			}, "GetACEByPrincipal", []string{"00c"}, [][]interface{}{
				[]interface{}{"hall-001", "hall", "00c", AT, 1, 0x3, true, false, false},
				[]interface{}{"hall-004", "hall", "00c", AT, 1000, 7, true, false, false},
			}},
			// {0, "hall-004", "hall", "00a", AT, 40, 0x4, true, false, false, now}, // 9
			{9, map[string]interface{}{
				"isGrant": false,
				"pos":     0,
			}, "GetACEByPrincipal", []string{"00a"}, [][]interface{}{
				[]interface{}{"hall-004", "hall", "00a", AT, 0, 0x4, false, false, false},
				[]interface{}{"hall-001", "hall", "00a", AT, 1, 0xff, true, false, false},
				[]interface{}{"hall-003", "hall", "00a", AT, 20, 0x7, true, false, false},
				[]interface{}{"hall-002", "hall", "00a", AT, 50, 0x3, true, false, false},
			}},
		}
	)

	// 清理数据
	defer func() {
		cleanAclEntry(db, pool)
	}()

	mgr := CreateAclManager(db, pool)

	db.LogMode(false)
	cleanAclEntry(db, pool)
	// 创建 AclEntry
	for _, ace := range aces {
		_, err := mgr.CreateACE(ace.Sid, ace.SidTyp, ace.ObjID, ace.ObjTyp,
			ace.Mask, ace.Pos,
			ace.IsGrant, ace.AuditSuccess, ace.AuditFailure)
		assert.Assertf(err == nil, "CreateACE failed: %v", err)
	}

	// GetXXX
	for i, c := range getCases {
		switch c.fn {
		case "GetAllACEs":
			res, err := mgr.GetAllACEs(c.params[0], c.params[1])
			assert.Assert(err == nil, "GetAllAces failed")
			resultEqual(t, res, c.result, i, c.fn)

		case "GetACEByPrincipal":
			res, err := mgr.GetACEByPrincipal(c.params[0])
			assert.Assert(err == nil, "GetACEByPrincipal failed")
			resultEqual(t, res, c.result, i, c.fn)

		case "GetACEByTarget":
			res, err := mgr.GetACEByTarget(c.params[0])
			assert.Assert(err == nil, "GetACEByPrincipal failed")
			resultEqual(t, res, c.result, i, c.fn)
		}
	}

	// 测试 缓存:
	//  GetPricipalACE
	//  getPricipalACEFromRedis
	//  getPricipalACEFromDB
	//  setPricipalACEToRedis
	res, err := getPricipalACEFromRedis(pool.Get(), "00a")
	assert.Assert(err == ErrRedisKeyNotExist, "getPricipalACEFromRedis: first time, key not exist")
	res, err = GetPricipalACE(db, pool.Get(), "00a")
	assert.Assert(err == nil, "GetPricipalACE failed")
	resultEqual(t, res, getCases[3].result, -1, "GetPricipalACE 00a")
	// 00b
	res, err = GetPricipalACE(db, pool.Get(), "00b")
	assert.Assert(err == nil, "GetPricipalACE failed")
	resultEqual(t, res, getCases[2].result, -1, "GetPricipalACE 00b")

	// 从缓存中获取的
	res, err = getPricipalACEFromRedis(pool.Get(), "00a")
	assert.Assertf(err == nil, "get data from redis failed: %v", err)
	resultEqual(t, res, getCases[3].result, -1, "getPricipalACEFromRedis 00a: second time, redis should has data")
	// 00b
	res, err = getPricipalACEFromRedis(pool.Get(), "00b")
	assert.Assertf(err == nil, "get data from redis failed: %v", err)
	resultEqual(t, res, getCases[2].result, -1, "getPricipalACEFromRedis 00b: second time, redis should has data")

	// 删除用户00a的记录， 缓存00b应该还在
	mgr.DeleteACEByPrincipal("00a")
	_, err = getPricipalACEFromRedis(pool.Get(), "00a")
	assert.Assert(err == ErrRedisKeyNotExist, "getPricipalACEFromRedis: key should be deleted")

	_, err = getPricipalACEFromRedis(pool.Get(), "00b")
	assert.Assert(err == nil, "getPricipalACEFromRedis: key should still exist")

	res, err = getPricipalObjectACEFromRedis(pool.Get(), "00b", "hall-001")
	assert.Assert(err == nil, "getPricipalACEFromRedis: key should still exist")
	resultEqual(t, res, getCases[1].result, -1, "getPricipalObjectACEFromRedis")

	// 删除一条数据时, 缓存是否清理
	mgr.DeleteACE(1)
	_, err = getPricipalACEFromRedis(pool.Get(), "00a")
	assert.Assert(err == ErrRedisKeyNotExist, "getPricipalACEFromRedis: key should be deleted")
	_, err = getPricipalACEFromRedis(pool.Get(), "00b")
	assert.Assert(err == ErrRedisKeyNotExist, "getPricipalACEFromRedis: key should be deleted")

	res, err = GetPricipalObjectACE(db, pool.Get(), "00b", "hall-001")
	assert.Assert(err == nil, "getPricipalACEFromRedis: key should still exist")
	resultEqual(t, res, getCases[1].result, -1, "getPricipalObjectACEFromRedis")

	// DeleteACEByPrincipal DeleteACEByTarget
	mgr.DeleteACEByPrincipal("00a")
	res, err = mgr.GetACEByPrincipal("00a")
	assert.Assert(err == nil, "should success")
	assert.Assert(len(res) == 0, "00a has been deleted")
	res, err = getPricipalACEFromRedis(pool.Get(), "00a")
	assert.Assert(len(res) == 0, "00a should been deleted")

	mgr.DeleteACEByTarget("hall-001")
	res, err = mgr.GetACEByTarget("hall-001")
	assert.Assert(err == nil, "should success")
	assert.Assert(len(res) == 0, "hall-001 has been deleted")
	res, err = getPricipalACEFromRedis(pool.Get(), "00b")
	assert.Assert(len(res) == 0, "00b redis data should been deleted")

	// 创建 AclEntry
	db.Exec("truncate acl_entries")
	for _, ace := range aces {
		_, err := mgr.CreateACE(ace.Sid, ace.SidTyp, ace.ObjID, ace.ObjTyp,
			ace.Mask, ace.Pos,
			ace.IsGrant, ace.AuditSuccess, ace.AuditFailure)
		assert.Assertf(err == nil, "CreateACE failed: %v", err)
	}
	// 测试 Update 00b hall-002
	for j, uc := range updateCases {
		err = mgr.UpdateACE(int64(uc.idx), uc.opt)
		assert.Assertf(err == nil, "update should success: %v", err)

		if uc.fn == "GetACEByPrincipal" {
			res, err := mgr.GetACEByPrincipal(uc.params[0])
			assert.Assert(err == nil, "GetACEByPrincipal after update should success")
			resultEqual(t, res, uc.result, j, "UpdateAce")
		} else {
			t.Log("Not support update func")
		}
	}

}

// 对比测试结果
func resultEqual(t *testing.T, res []*AclEntry, expect [][]interface{}, i int, fn string) {
	assert.Assertf(len(res) == len(expect),
		"%d %s result length NOT equal: len(res)=%d len(expect)=%d", i, fn, len(res), len(expect))
	match := true

	var (
		ii int
		r  *AclEntry
		e  []interface{}
	)
	for ii, r = range res {
		e = expect[ii]
		if r.ObjID == e[0].(string) &&
			r.ObjTyp == e[1].(string) &&
			r.Sid == e[2].(string) &&
			r.SidTyp == e[3].(string) &&
			r.Pos == e[4].(int) &&
			r.Mask == uint32(e[5].(int)) &&
			r.IsGrant == e[6].(bool) &&
			r.AuditSuccess == e[7].(bool) &&
			r.AuditFailure == e[8].(bool) {
			continue
		}
		match = false
		break
	}
	if !match {
		s := "[\n"
		for _, ele := range res {
			s += fmt.Sprint(ele)
			s += "\n"
		}
		s += "]"
		t.Fatalf("%d:%d %s result %v NOT equal with expect: %v, result: %v", i, ii, fn, r, expect[ii], s)
	}
}

// Smember 得到的顺序是不是入库的顺序
// Args.AddFlat的 正确性
//
