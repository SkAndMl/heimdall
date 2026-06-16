package scan

import (
	"os"
	"path/filepath"
	"sort"
)

type Node struct {
	Path     string
	TotSize  int64
	Type     string
	Children []*Node
	Depth    int
}

func CreateRootNode(path string) (*Node, error) {

	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	node := &Node{Path: path}

	if !info.IsDir() {
		node.Type = "file"
		node.TotSize = info.Size()
		node.Depth = 0
		return node, nil
	}

	node.Type = "dir"
	node.Depth = 0

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsPermission(err) {
			return node, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		childNode, err := CreateRootNode(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, err
		}
		node.TotSize += childNode.TotSize
		node.Children = append(node.Children, childNode)
		node.Depth = max(node.Depth, childNode.Depth+1)
	}

	return node, nil
}

func (n *Node) GetLargestFiles(nFiles int) []*Node {
	queue := []*Node{n}
	largestFiles := make([]*Node, 0)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node == nil {
			continue
		}

		if node.Type == "file" {
			largestFiles = append(largestFiles, node)
			sort.Slice(largestFiles, func(i, j int) bool {
				return largestFiles[i].TotSize >= largestFiles[j].TotSize
			})
			if len(largestFiles) >= nFiles {
				largestFiles = largestFiles[:nFiles]
			}
			continue
		}
		queue = append(queue, node.Children...)
	}

	return largestFiles
}
