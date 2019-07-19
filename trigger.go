package folders

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"cloud.google.com/go/bigquery"
	"github.com/pmenglund/gcp-folders/fetcher"
	"github.com/pmenglund/gcp-folders/saver"
	"github.com/pmenglund/gcp-folders/tree"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Message is the message sent to the cloud function
type Message struct {
	Data []byte `json:"data"`
}

// Dump is a Cloud Function that walks the GCP folder structure and saves it
// to a BigQuery table which can be used to lookup folder id to folder name
// for use in DataStudio.
// It can be configured via environment variables
// ROOT
// MAX_DEPTH
// DATASET
// TABLE
// PROJECT
func Dump(ctx context.Context, msg Message) error {
	id := os.Getenv("ROOT")
	if id == "" {
		return errors.New("ROOT environment variable required")
	}
	log.Printf("ROOT is %s", id)

	md := os.Getenv("MAX_DEPTH")
	if md == "" {
		md = "4"
	}
	max, err := strconv.Atoi(md)
	if err != nil {
		return fmt.Errorf("failed to convert MAX_DEPTH %s to int: %v", md, err)
	}
	log.Printf("MAX_DEPTH is %d", max)

	dataset := os.Getenv("DATASET")
	if dataset == "" {
		return errors.New("DATASET environment variable required")
	}
	log.Printf("DATASET is %s", dataset)

	project := os.Getenv("PROJECT")
	if project == "" {
		return errors.New("PROJECT environment variable required")
	}
	log.Printf("PROJECT is %s", project)

	table := os.Getenv("TABLE")
	if table == "" {
		table = "folders"
	}
	log.Printf("TABLE is %s", table)

	conf := fetcher.Config{
		Verbose:  true,
		MaxDepth: max,
	}
	f, err := fetcher.New(ctx, conf)
	if err != nil {
		return err
	}

	root, err := f.Fetch(id)
	if err != nil {
		return err
	}

	folders := tree.Flatten(root)

	creds, err := google.FindDefaultCredentials(ctx, bigquery.Scope)
	if err != nil {
		return err
	}
	bq, err := bigquery.NewClient(ctx, project, option.WithCredentials(creds))
	if err != nil {
		return err
	}

	s := saver.New(ctx, bq, dataset, table)
	_, err = s.Save(folders)
	if err != nil {
		return err
	}

	return nil
}
