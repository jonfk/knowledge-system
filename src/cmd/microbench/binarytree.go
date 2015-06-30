package main

type Tree struct {
	Left  *Tree
	Value Data
	Right *Tree
}

func sortedArrayToBST(a []Data) *Tree {
	if len(a) == 0 {
		return nil
	}
	if len(a) < 2 {
		return &Tree{Left: nil, Value: a[0], Right: nil}
	}
	mid := len(a) / 2
	left := sortedArrayToBST(a[:mid])
	right := sortedArrayToBST(a[mid+1:])
	return &Tree{Left: left, Value: a[mid], Right: right}
}

func (a *Tree) Get(x string) Data {
	root := a
	for root.Value.Label != x && root != nil {
		if x > root.Value.Label {
			root = root.Right
		} else {
			root = root.Left
		}
	}
	if root != nil {
		return root.Value
	} else {
		return Data{}
	}

}
