package perms

type Perm int

const (
	None Perm = iota
	Read
	Write
)
