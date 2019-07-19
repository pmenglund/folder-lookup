package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		s   string
		id  string
		nt  NodeType
		err bool
	}{
		{"organizations/123", "123", OrganizationType, false},
		{"folders/123", "123", FolderType, false},
		{"foobar/123", "123", UnknownType, false},
		{"foobar", "", UnknownType, true},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			id, nt, err := Split(tt.s)
			assert.Equal(t, err != nil, tt.err)
			assert.Equal(t, id, tt.id)
			assert.Equal(t, nt, tt.nt)
		})
	}
}

func TestNodeTypeFor(t *testing.T) {
	tests := []struct {
		t  string
		nt NodeType
	}{
		{"organizations", OrganizationType},
		{"folders", FolderType},
		{"foobar", UnknownType},
	}
	for _, tt := range tests {
		t.Run(tt.t, func(t *testing.T) {
			nt := NodeTypeFor(tt.t)
			assert.Equal(t, nt, tt.nt)
		})
	}
}

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		s  string
		nt NodeType
	}{
		{"organizations", OrganizationType},
		{"folders", FolderType},
		{"unknown", UnknownType},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert.Equal(t, tt.nt.String(), tt.s)
		})
	}
}
