package tree

type Tree struct {
	tlds map[string]*node
}

type node struct {
	childrens map[string]*node
	ip        string
}

func NewTree() *Tree {
	return &Tree{
		tlds: map[string]*node{},
	}
}
