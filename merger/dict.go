package merger

import (
	"github.com/xreception/go-swagen/utils"
)

// Dict dict tree
type Dict struct {
	root *Node
}

// Node node of dict tree
type Node struct {
	word     string
	children []*Node
	origin   string
}

// 插入字符串
func (d *Dict) insertStr(str string) {
	if d == nil {
		return
	}
	if d.root == nil {
		d.root = &Node{}
	}
	splitted := utils.Split(str)
	d.root.insert(splitted, str)
}

// 压缩
func (d *Dict) compress() {
	d.root.compress()
}

// 获取原始到简写的映射表
func (d *Dict) getOrigToShortMap() map[string]string {
	m := make(map[string]string)
	d.root.walk("", m)
	return m
}

func (n *Node) insert(words []string, origin string) {
	if len(words) == 0 {
		return
	}

	word := words[len(words)-1]
	child := n.findChild(word)
	if child == nil {
		child = &Node{
			word: word,
		}
		n.children = append(n.children, child)
	}
	if len(words) == 1 {
		child.origin = origin
	}
	child.insert(words[:len(words)-1], origin)
}

func (n *Node) findChild(word string) *Node {
	for _, child := range n.children {
		if child.word == word {
			return child
		}
	}
	return nil
}

func (n *Node) compress() {
	if len(n.children) == 0 {
		return
	}
	for _, child := range n.children {
		if child.origin != "" && n.origin == "" && n.word != "" {
			n.origin = child.origin
			child.origin = ""
		}
		child.compress()
	}
}

func (n *Node) walk(path string, m map[string]string) {
	if n.origin != "" {
		m[n.origin] = path
	}
	for _, child := range n.children {
		child.walk(child.word+path, m)
	}
}
