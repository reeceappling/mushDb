package lineage

import (
	"context"
	"errors"
	"github.com/reeceappling/mushDb/api"
	"iter"
)

type Node struct {
	Item     api.MainCollectionItem
	Children []*Node
}

func TreeFromParentItem(ctx context.Context, rootItem api.MainCollectionItem) (*Node, error) {
	root := &Node{Item: rootItem}
	children, err := rootItem.Children(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: add max depth?
	childNodes := make([]*Node, len(children))
	for i, child := range children {
		childNode, childErr := TreeFromParentItem(ctx, child)
		if childErr != nil {
			return root, err
		}
		childNodes[i] = childNode
	}
	return root, nil
}
func LineageOf(ctx context.Context, childItem api.MainCollectionItem) iter.Seq2[api.MainCollectionItem, error] {
	parentOfChild := func(thisItem api.MainCollectionItem) (api.MainCollectionItem, error) { // TODO: maybe not pointers?
		return thisItem.Parent(ctx)
	}
	return func(yield func(api.MainCollectionItem, error) bool) {
		currentItem := childItem
		var err error = nil
		for {
			currentItem, err = parentOfChild(currentItem)
			if err != nil {
				yield(nil, err)
				return
			}
			if currentItem == nil {
				yield(nil, errors.New("no parent")) // TODO: put this somewhere as a var?
				return
			}
			if !yield(currentItem, nil) {
				return
			}
		}
	}
}

//func (parent *Node) GetChildren(ctx context.Context) ([]api.MainCollectionItem, error) {
//	return parent.Item.Children(ctx)
//	// TODO: normally, check for fruit, as well as all transfersOut
//	// TODO: fruits should not check for fruits, but should also check for spore prints and swabs, and also use all transfersOut
//	// TODO: spore prints should not check for fruits, but should also check for spore swabs and MSS, and also use all transfersOut
//	// TODO: if lc check for lcSyringe (if creation does not entail a transfer), and also use all transfersOut
//	// TODO: should lcSyringe check for fruits? stasis tube should or should not check fruits?
//	// TODO: water jar should check nothing
//}
//func (parent *Node) GetParent(ctx context.Context) (api.MainCollectionItem, error) { // Can be nil
//	return parent.Item.Parent(ctx)
//	// TODO: if fruit, check for spore prints and swabs
//	// TODO: if not fruit, check for fruits, and also use all transfersOut
//}
