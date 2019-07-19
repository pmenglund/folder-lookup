package saver

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/bigquery"
	"github.com/pkg/errors"
	"github.com/pmenglund/gcp-folders/tree"
	"google.golang.org/api/iterator"
)

type Saver struct {
	ctx     context.Context
	bq      wrapper
	dataset string
	table   string
}

func New(ctx context.Context, client *bigquery.Client, dataset, table string) *Saver {
	return &Saver{ctx, &bq{client}, dataset, table}
}

// Save saves the folders and returns how many of them that were saved or updated
func (s *Saver) Save(folders map[string]tree.Folder) (int, error) {
	saved := 0

	cur, err := s.bq.current(s.ctx, s.dataset, s.table)
	if err != nil {
		return saved, err
	}

	for _, f := range folders {
		var qs string
		if folder, found := cur[f.ID]; found {
			if f.Name == folder.Name && f.Parent == folder.Parent && f.Level == folder.Level {
				continue // no change
			}
			qs = fmt.Sprintf(`UPDATE %s.%s SET name = "%s", level = %d, parent = "%s" WHERE id = "%s"`,
				s.dataset, s.table, f.Name, f.Level, f.Parent, f.ID)
			log.Printf("updating folder id %s to %s", f.ID, f.Name)
		} else {
			qs = fmt.Sprintf(`INSERT INTO %s.%s (id, name, level, parent) VALUES("%s", "%s", %d, "%s")`,
				s.dataset, s.table, f.ID, f.Name, f.Level, f.Parent)
			log.Printf("inserting folder %s named %s at level %d", f.ID, f.Name, f.Level)
		}

		if err := s.bq.exec(s.ctx, qs); err != nil {
			log.Println(err)
		} else {
			saved++
		}
	}
	log.Printf("saved %d out of %d folders", saved, len(folders))

	deleted := 0
	for id, f := range cur {
		if _, ok := folders[id]; !ok {
			log.Printf("deleting folder id %s named %s", id, f.Name)
			qs := fmt.Sprintf(`DELETE %s.%s WHERE id = "%s"`,
				s.dataset, s.table, id)
			if err := s.bq.exec(s.ctx, qs); err != nil {
				log.Println(err)
			} else {
				deleted++
			}
		}
	}
	if deleted > 0 {
		log.Printf("deleted %d folders", deleted)
	}

	return saved, nil
}

type wrapper interface {
	current(context.Context, string, string) (map[string]tree.Folder, error)
	exec(context.Context, string) error
}

type bq struct {
	client *bigquery.Client
}

func (b *bq) current(ctx context.Context, dataset, table string) (map[string]tree.Folder, error) {
	cur := make(map[string]tree.Folder)

	q := b.client.Query(fmt.Sprintf("SELECT * FROM %s.%s", dataset, table))

	it, err := q.Read(ctx)
	if err != nil {
		return cur, errors.Wrap(err, "failed to read query response")
	}

	for {
		var f tree.Folder
		err := it.Next(&f)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("failed to read next folder: %+v", err)
		} else {
			cur[f.ID] = f
		}
	}
	log.Printf("loaded %d existing folder mappings", len(cur))
	return cur, nil
}

func (b *bq) exec(ctx context.Context, query string) error {
	q := b.client.Query(query)
	job, err := q.Run(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run query")
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to wait for job to complete")
	}
	if !status.Done() {
		return errors.Wrap(status.Err(), "job failed")
	}
	return nil
}
