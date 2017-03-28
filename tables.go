package acl

import "github.com/jinzhu/gorm"

// 创建删除tables

var entities = []interface{}{
	Role{},
	UserRole{},
	AclEntry{},
}

// CreateModels 创建models对应的表
func CreateModels(db *gorm.DB) (err error) {
	for _, en := range entities {
		db.AutoMigrate(en)
	}
	return nil
}

// DeleteModels drop tables
func DeleteModels(db *gorm.DB) {
	for _, en := range entities {
		db.DropTableIfExists(en)
	}
}

// TruncateModels 清除table中的数据
func TruncateModels(db *gorm.DB) {
	db.Exec("truncate roles")
	db.Exec("truncate user_roles")
	db.Exec("truncate acl_entries")
}
