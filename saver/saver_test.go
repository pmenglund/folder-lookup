package saver

import (
	"context"
	"testing"

	"github.com/pmenglund/gcp-folders/tree"
	"github.com/stretchr/testify/assert"
)

func TestSaver(t *testing.T) {
	folders := map[string]tree.Folder{
		"0": tree.Folder{ID: "0", Name: "root", Level: 0, Parent: ""},
		"1": tree.Folder{ID: "1", Name: "keep", Level: 1, Parent: "0"},
		"3": tree.Folder{ID: "3", Name: "add", Level: 1, Parent: "0"},
		"4": tree.Folder{ID: "4", Name: "update", Level: 1, Parent: "0"},
	}
	fake := &fakeBQ{}
	s := &Saver{
		dataset: "dataset",
		table:   "table",
		bq:      fake,
	}

	saved, err := s.Save(folders)
	assert.Nil(t, err)
	assert.Equal(t, 2, saved)
	assert.Contains(t, fake.queries, `INSERT INTO dataset.table (id, name, level, parent) VALUES("3", "add", 1, "0")`)
	assert.Contains(t, fake.queries, `UPDATE dataset.table SET name = "update", level = 1, parent = "0" WHERE id = "4"`)
	assert.Contains(t, fake.queries, `DELETE dataset.table WHERE id = "2"`)
}

type fakeBQ struct {
	queries []string
}

func (f *fakeBQ) current(ctx context.Context, dataset, table string) (map[string]tree.Folder, error) {
	cur := map[string]tree.Folder{
		"0": tree.Folder{ID: "0", Name: "root", Level: 0, Parent: ""},
		"1": tree.Folder{ID: "1", Name: "keep", Level: 1, Parent: "0"},
		"2": tree.Folder{ID: "2", Name: "delete", Level: 1, Parent: "0"},
		"4": tree.Folder{ID: "4", Name: "old name", Level: 1, Parent: "0"},
		}
	return cur, nil
}

func (f *fakeBQ) exec(ctx context.Context, query string) error {
	f.queries = append(f.queries, query)
	return nil
}
