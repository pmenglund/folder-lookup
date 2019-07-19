package fetcher

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/pmenglund/gcp-folders/tree"
	"golang.org/x/oauth2/google"
	crm "google.golang.org/api/cloudresourcemanager/v2beta1"
)

type Config struct {
	Verbose  bool
	Root     string
	MaxDepth int
}

type Fetcher struct {
	Config
	ctx context.Context
	svc *crm.FoldersService
}

func New(ctx context.Context, conf Config) (*Fetcher, error) {
	client, err := google.DefaultClient(ctx, crm.CloudPlatformReadOnlyScope)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default client")
	}

	crmSvc, err := crm.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloud resource manager client")
	}

	svc := crm.NewFoldersService(crmSvc)

	return &Fetcher{
		ctx:    ctx,
		svc:    svc,
		Config: conf,
	}, nil
}

func (f *Fetcher) Fetch(id string) (*tree.Node, error) {
	return f.fetch(id, "root", nil, 0)
}

func (f *Fetcher) fetch(name, display string, parent *tree.Node, depth int) (*tree.Node, error) {
	if f.Verbose {
		log.Printf("fetching %s", name)
	}

	id, t, err := tree.Split(name)
	if err != nil {
		return nil, err
	}

	if t == tree.UnknownType {
		return nil, fmt.Errorf("unknown node type %s", t)
	}

	node := &tree.Node{
		ID:          id,
		Type:        t,
		DisplayName: display,
		Parent:      parent,
	}

	lc := f.svc.List()
	lc.Parent(name)

	for {
		resp, err := lc.Do()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list folders in %s", name)
		}

		if depth < f.MaxDepth {
			for _, folder := range resp.Folders {
				fldr, err := f.fetch(folder.Name, folder.DisplayName, node, depth+1)
				if err != nil {
					log.Println(err)
				} else {
					node.AddChild(fldr)
				}
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		lc.PageToken(resp.NextPageToken)
	}

	return node, nil
}
