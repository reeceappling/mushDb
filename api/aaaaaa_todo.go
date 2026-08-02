package api

import "errors"

//go:generate ./goGenerator/mygenerator

// TODO: add orgs? Owners of each item? Want it to be extensible to allow multiple users on the internet to utilize the system...

// Rollback is a singly-linked list (LIFO) to perform rollbacks in a transaction (SAGA)
type Rollback struct { // TODO: CONSIDER USING ROLLBACKS!!!!
	first *CompensatingAction
}

func (rb *Rollback) AddHandler(onRollback func() error) {
	rb.first = &CompensatingAction{ // TODO: validate works ok
		next:       rb.first,
		onRollback: onRollback,
	}
}

func (rb *Rollback) Do() error {
	headAction := rb.first
	if headAction == nil {
		return nil
	}
	return headAction.Do()
}

type CompensatingAction struct {
	next       *CompensatingAction
	onRollback func() error // MUST BE IDEMPOTENT!
}

func (ca *CompensatingAction) Do() error {
	thisErr := ca.onRollback()
	if ca.next == nil {
		return thisErr
	}
	return errors.Join(thisErr, ca.next.Do())
}
