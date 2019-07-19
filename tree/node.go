package tree

import (
	"fmt"
	"strings"
	"sync"
)

type NodeType int

const (
	UnknownType NodeType = iota
	OrganizationType
	FolderType
	ProjectType
)

func (n NodeType) String() string {
	names := [...]string{
		"unknown",
		"organizations",
		"folders",
		"projects",
	}
	return names[int(n)]
}

func NodeTypeFor(t string) NodeType {
	switch t {
	case "organizations":
		return OrganizationType
	case "folders":
		return FolderType
	case "projects":
		return ProjectType
	default:
		return UnknownType
	}
}

// Folder is a representation of the GCP folder structure suitable to be saved to a BigQuery table
type Folder struct {
	ID     string
	Name   string
	Level  int
	Parent string
}

type Node struct {
	ID          string
	Type        NodeType
	DisplayName string
	Parent      *Node
	Children    []*Node
}

// split splits a GCP id, e.g. organizations/123 or folders/123 into its components
func Split(s string) (string, NodeType, error) {
	fields := strings.Split(s, "/")
	if len(fields) != 2 {
		return "", UnknownType, fmt.Errorf("failed to split %s in two", s)
	}
	return fields[1], NodeTypeFor(fields[0]), nil
}

func (n *Node) AddChild(c *Node) {
	n.Children = append(n.Children, c)
}

// Visit recursively invokes `fn` on the node and all children
func (n *Node) Visit(fn func(level int, node *Node)) {
	n.visit(0, fn)
}

func (n *Node) visit(level int, fn func(int, *Node)) {
	fn(level, n)
	for _, node := range n.Children {
		node.visit(level+1, fn)
	}
}

// Flatten takes a node and returns a flattened view of it
func Flatten(root *Node) map[string]Folder {
	fc := make(chan Folder, 10)
	folders := make(map[string]Folder)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for f := range fc {
			folders[f.ID] = f
		}
		wg.Done()
	}()

	root.Visit(flatten(fc))
	close(fc)
	wg.Wait()

	return folders
}

// TODO: use a map instead of chan
func flatten(folders chan Folder) func(int, *Node) {
	return func(level int, node *Node) {
		var parent string
		if node.Parent != nil {
			parent = node.Parent.ID
		}
		folders <- Folder{
			ID:     node.ID,
			Name:   node.DisplayName,
			Parent: parent,
			Level:  level,
		}
	}
}
