package acl

import (
	"testing"

	"github.com/guotie/assert"
)

func testCreateTables(t *testing.T) {
	db, err := openDB()
	assert.Assert(err == nil, "openDB should success")
	DeleteModels(db)
	CreateModels(db)
}
