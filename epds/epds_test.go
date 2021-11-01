package epds

import (
	"testing"

	"github.com/blgolden/igendec/users"
)

func TestList(t *testing.T) {
	DatabasePath = "/home/joel/Code/igendec/epds"

	access := []users.AccessPath{
		{Path: "*"},
		{Path: "AHA/*", Deny: true},
	}

	_ = ListDatabases(access)
}
