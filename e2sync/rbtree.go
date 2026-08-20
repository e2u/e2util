package e2sync

type Color bool

const (
	Red   Color = true
	Black Color = false
)

type Comparer[K any] interface {
	Compare(a, b K) int
}

type Node[K any, V any] struct {
	Key    K
	Value  V
	Left   *Node[K, V]
	Right  *Node[K, V]
	Parent *Node[K, V]
	Color  Color
}

type RBTreeMap[K any, V any] struct {
	Root     *Node[K, V]
	Nil      *Node[K, V]
	comparer Comparer[K]
}

func NewRBTreeMap[K any, V any](comparer Comparer[K]) *RBTreeMap[K, V] {
	nilNode := &Node[K, V]{Color: Black}
	nilNode.Left = nilNode
	nilNode.Right = nilNode
	return &RBTreeMap[K, V]{
		Nil:      nilNode,
		Root:     nilNode,
		comparer: comparer,
	}
}

func (t *RBTreeMap[K, V]) leftRotate(x *Node[K, V]) {
	var y = x.Right
	x.Right = y.Left
	if y.Left != t.Nil {
		y.Left.Parent = x
	}
	y.Parent = x.Parent
	//nolint:gocritic
	if x.Parent == t.Nil {
		t.Root = y
	} else if x == x.Parent.Left {
		x.Parent.Left = y
	} else {
		x.Parent.Right = y
	}

	y.Left = x
	x.Parent = y
}

func (t *RBTreeMap[K, V]) rightRotate(x *Node[K, V]) {
	y := x.Left
	x.Left = y.Right
	if y.Right != t.Nil {
		y.Right.Parent = x
	}
	y.Parent = x.Parent
	//nolint:gocritic
	if x.Parent == t.Nil {
		t.Root = y
	} else if x == x.Parent.Right {
		x.Parent.Right = y
	} else {
		x.Parent.Left = y
	}
	y.Right = x
	x.Parent = y
}

func (t *RBTreeMap[K, V]) Store(key K, value V) {
	node := &Node[K, V]{Key: key, Value: value, Left: t.Nil, Right: t.Nil, Color: Red}
	current := t.Root
	parent := t.Nil

	for current != t.Nil {
		parent = current
		comp := t.comparer.Compare(key, current.Key)
		//nolint:gocritic
		if comp < 0 {
			current = current.Left
		} else if comp > 0 {
			current = current.Right
		} else {
			current.Value = value
			return
		}
	}

	node.Parent = parent
	//nolint:gocritic
	if parent == t.Nil {
		t.Root = node
	} else if t.comparer.Compare(key, parent.Key) < 0 {
		parent.Left = node
	} else {
		parent.Right = node
	}

	t.fixInsert(node)
}

func (t *RBTreeMap[K, V]) fixInsert(z *Node[K, V]) {
	for z.Parent.Color == Red {
		if z.Parent == z.Parent.Parent.Left {
			y := z.Parent.Parent.Right
			if y.Color == Red {
				z.Parent.Color = Black
				y.Color = Black
				z.Parent.Parent.Color = Red
				z = z.Parent.Parent
			} else {
				if z == z.Parent.Right {
					z = z.Parent
					t.leftRotate(z)
				}
				z.Parent.Color = Black
				z.Parent.Parent.Color = Red
				t.rightRotate(z.Parent.Parent)
			}
		} else {
			y := z.Parent.Parent.Left
			if y.Color == Red {
				z.Parent.Color = Black
				y.Color = Black
				z.Parent.Parent.Color = Red
				z = z.Parent.Parent
			} else {
				if z == z.Parent.Left {
					z = z.Parent
					t.rightRotate(z)
				}
				z.Parent.Color = Black
				z.Parent.Parent.Color = Red
				t.leftRotate(z.Parent.Parent)
			}
		}
		if z == t.Root {
			break
		}
	}
	t.Root.Color = Black
}

func (t *RBTreeMap[K, V]) InOrderTraversal(visit func(key K, value V)) {
	t.inOrder(t.Root, visit)
}

func (t *RBTreeMap[K, V]) inOrder(node *Node[K, V], visit func(key K, value V)) {
	if node != t.Nil {
		t.inOrder(node.Left, visit)
		visit(node.Key, node.Value)
		t.inOrder(node.Right, visit)
	}
}

func (t *RBTreeMap[K, V]) Load(key K) (V, bool) {
	current := t.Root
	var zero V
	for current != t.Nil {
		comp := t.comparer.Compare(key, current.Key)
		//nolint:gocritic
		if comp < 0 {
			current = current.Left
		} else if comp > 0 {
			current = current.Right
		} else {
			return current.Value, true
		}
	}
	return zero, false
}

func (t *RBTreeMap[K, V]) RangeAscending(visit func(key K, value V) bool) {
	t.rangeAscending(t.Root, visit)
}

func (t *RBTreeMap[K, V]) rangeAscending(node *Node[K, V], visit func(key K, value V) bool) bool {
	if node == t.Nil {
		return true
	}
	if !t.rangeAscending(node.Left, visit) {
		return false
	}
	if !visit(node.Key, node.Value) {
		return false
	}
	return t.rangeAscending(node.Right, visit)
}

func (t *RBTreeMap[K, V]) RangeDescending(visit func(key K, value V) bool) {
	t.rangeDescending(t.Root, visit)
}

func (t *RBTreeMap[K, V]) rangeDescending(node *Node[K, V], visit func(key K, value V) bool) bool {
	if node == t.Nil {
		return true
	}
	if !t.rangeDescending(node.Right, visit) {
		return false
	}
	if !visit(node.Key, node.Value) {
		return false
	}
	return t.rangeDescending(node.Left, visit)
}

type StringComparer struct{}

func (StringComparer) Compare(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
