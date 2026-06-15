package gitdiff

import (
	"sort"
	"strings"

	"github.com/txie/verdiff/internal/verdiff"
)

// BuildDirTree constructs a directory tree from a list of file diffs,
// aggregating line counts at each level.
func BuildDirTree(files []verdiff.FileDiff) *verdiff.DirNode {
	root := &verdiff.DirNode{Name: ".", Path: "."}
	for _, f := range files {
		insertFile(root, f)
	}
	sortTree(root)
	return root
}

func insertFile(root *verdiff.DirNode, f verdiff.FileDiff) {
	parts := strings.Split(f.Path, "/")

	node := root
	for i := 0; i < len(parts)-1; i++ {
		dirName := parts[i]
		dirPath := strings.Join(parts[:i+1], "/")
		child := findChild(node, dirName)
		if child == nil {
			child = &verdiff.DirNode{Name: dirName, Path: dirPath}
			node.Children = append(node.Children, child)
		}
		child.LinesAdded += f.LinesAdded
		child.LinesDeleted += f.LinesDeleted
		child.FileCount++
		node = child
	}

	// Add file to the leaf directory
	node.Files = append(node.Files, f)

	// Propagate counts to root
	root.LinesAdded += f.LinesAdded
	root.LinesDeleted += f.LinesDeleted
	root.FileCount++
}

func findChild(node *verdiff.DirNode, name string) *verdiff.DirNode {
	for _, c := range node.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func sortTree(node *verdiff.DirNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].TotalChange() > node.Children[j].TotalChange()
	})
	for _, c := range node.Children {
		sortTree(c)
	}
}
