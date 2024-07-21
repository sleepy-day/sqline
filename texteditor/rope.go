package texteditor

type Branch struct {
	parent, left, right *Branch
	dataNode            bool
}

type Leaf struct {
	parent *Branch
	text   []rune
	length int
}
