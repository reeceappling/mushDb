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

func TreeFromParentItem(ctx context.Context, rootItem api.MainCollectionItem, maxDepth *int) (*Node, error) {
	var nextMaxDepth *int = nil
	if maxDepth != nil {
		temp := *maxDepth - 1
		nextMaxDepth = &temp
	}
	out := Node{Item: rootItem}
	if maxDepth != nil && *maxDepth == 0 {
		return &out, nil
	}
	children, err := rootItem.Children(ctx)
	if err != nil {
		return nil, err
	}
	childNodes := make([]*Node, len(children))
	for i, child := range children {
		childNode, childErr := TreeFromParentItem(ctx, child, nextMaxDepth)
		if childErr != nil {
			return &out, err
		}
		childNodes[i] = childNode
	}
	return &out, nil
}

var ErrNoParentItem = errors.New("no parent") // TODO: use for error checking?

func LineageOf(ctx context.Context, childItem api.MainCollectionItem, maxNumParents *int) iter.Seq2[api.MainCollectionItem, error] {
	return func(yield func(api.MainCollectionItem, error) bool) {
		ct := 0
		var err error = nil
		currentItem := childItem
		for {
			currentItem, err = currentItem.GetParent(ctx)
			if err != nil {
				yield(nil, err)
				return
			}
			if currentItem == nil {
				yield(nil, ErrNoParentItem)
				return
			}
			if !yield(currentItem, nil) {
				return
			}
			ct++
			if maxNumParents != nil && ct == *maxNumParents {
				return
			}
		}
	}
}
