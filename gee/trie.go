package gee

import "strings"

type node struct {
	pattern  string  // 待匹配路由，录入 /p/:lang
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool    // 是否精确匹配，part 含有 : 或 * 时为 true
}

// 匹配算法，只匹配孩子节点中第一个相同的节点；这个用于创建新的路由（如果匹配到了，直接用；如果没匹配到，创建一个新的）
// 输入一个 part（路由中的一部分，例如 /p/python 中的 python），返回一个匹配的节点
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 匹配算法，匹配孩子节点中所有可以匹配成功的节点（动态路由或者静态路由）；这个主要用于查找
// 输入一个 part（路由中的一部分，例如 /p/python 中的 python），返回匹配的节点列表
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) insert(pattern string, parts []string, height int) {
	// 如果当前结点的深度（height）恰好为 parts 的长度，那么说明 parts 已经走到最后一个路由了，直接将 pattern 赋值给当前节点的 pattern
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	// 如果 len(parts) != height，说明还没有遍历完parts，直接取出当前深度的路由 parts[height]
	part := parts[height]
	child := n.matchChild(part) // 返回 n 结点的孩子节点中和 part 匹配的节点；注意，这里的 n 的孩子结点实际上就是 height 这一层，不是 height+1 层
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	// 递归处理，直到 pattern 走到底
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	// 如果 pattern 走到最后一个 part 了，或者当前节点的part以 * 为前缀
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		// 如果当前节点的 pattern 为空，说明不是合法路由，匹配失败
		if n.pattern == "" {
			return nil
		}
		// 如果 pattern 不为空，说明匹配成功
		return n
	}

	// 如果 pattern 还没走完6
	part := parts[height]
	// 匹配所有匹配的节点并返回给 children
	children := n.matchChildren(part)
	// DFS：对每条可能路径继续往下试
	for _, child := range children {
		// 递归调用 search 来找出 child 下匹配的节点
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
