// Package docker is docker utils
package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/moby/go-archive"
	"github.com/moby/patternmatcher/ignorefile"
	"golang.org/x/exp/slices"
)

// BuildLocalContext builds a tarball of the local context.
func BuildLocalContext(dir string, depth int) error {
	excludes, err := getDockerIgnores(dir)
	if err != nil {
		return err
	}

	// We still need to send the dockerfile even if it is ignored
	if i := slices.Index(excludes, "Dockerfile"); i != -1 {
		excludes = slices.Delete(excludes, i, i+1)
	}

	t, err := archive.TarWithOptions(dir, &archive.TarOptions{Compression: archive.Gzip, ExcludePatterns: excludes})
	if err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	// Untar the tarball into a temporary directory to get the list of files and their size
	defer func() {
		err := t.Close()
		if err != nil {
			panic(err)
		}
	}()

	// Make the temporary directory
	tmpDir, err := os.MkdirTemp("", "docker-context-")
	if err != nil {
		return err
	}
	// Clean up the temporary directory after use
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			panic(err)
		}
	}()

	err = archive.Untar(t, tmpDir, nil)
	if err != nil {
		return err
	}

	// Print the contents of the temporary directory and their sizes (unlimited depth)
	return PrintDirectoryTree(tmpDir, depth)

}

func getDockerIgnores(dir string) ([]string, error) {
	f, err := os.Open(filepath.Join(dir, ".dockerignore"))
	// Note that a missing .dockerignore file isn't treated as an error
	switch {
	case os.IsNotExist(err):
		return nil, nil // No .dockerignore file, no excludes
	case err != nil:
		return nil, fmt.Errorf("failed to open .dockerignore: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			panic(fmt.Errorf("failed to close .dockerignore: %w", err))
		}
	}()

	excludes, err := ignorefile.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read .dockerignore: %w", err)
	}

	return excludes, nil
}

// TreeNode represents a file or directory in the tree
type TreeNode struct {
	Name     string
	Path     string
	Size     int64
	IsDir    bool
	Children []*TreeNode
}

// PrintDirectoryTree traverses a directory and prints its contents in a tree format with sizes
// depth of -1 means unlimited depth, 0 means only root, 1 means root + direct children, etc.
func PrintDirectoryTree(dir string, depth int) error {
	root, err := buildTree(dir)
	if err != nil {
		return err
	}

	// Print root
	fmt.Printf("%s (%s)\n", root.Name, formatSize(root.Size))

	// Print children with initial prefix if depth allows
	if depth != 0 {
		for i, child := range root.Children {
			printNodeWithDepth(child, "", i == len(root.Children)-1, 1, depth)
		}
	}

	return nil
}

func buildTree(path string) (*TreeNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	node := &TreeNode{
		Name:  filepath.Base(path),
		Path:  path,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		node.Size = info.Size()
		return node, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		child, err := buildTree(childPath)
		if err != nil {
			// Skip files we can't read
			continue
		}
		node.Children = append(node.Children, child)
		node.Size += child.Size
	}

	// Sort children: directories first, then alphabetically
	sort.Slice(node.Children, func(i, j int) bool {
		if node.Children[i].IsDir != node.Children[j].IsDir {
			return node.Children[i].IsDir
		}
		return node.Children[i].Name < node.Children[j].Name
	})

	return node, nil
}

func printNodeWithDepth(node *TreeNode, prefix string, isLast bool, currentDepth int, maxDepth int) {
	// Print the current node with tree connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	fmt.Printf("%s%s%s (%s)\n", prefix, connector, node.Name, formatSize(node.Size))

	// Check if we should print children based on depth
	if maxDepth != -1 && currentDepth >= maxDepth {
		return
	}

	// Prepare prefix for children
	newPrefix := prefix
	if isLast {
		newPrefix += "    "
	} else {
		newPrefix += "│   "
	}

	// Print children
	for i, child := range node.Children {
		printNodeWithDepth(child, newPrefix, i == len(node.Children)-1, currentDepth+1, maxDepth)
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
