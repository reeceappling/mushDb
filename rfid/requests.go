package rfid

type Data[T any] struct {
	data     T
	disabled bool
}

func DataStripped[T any](data Data[T]) T {
	return data.data
}

type AllEntries[T any] struct {
	Existing []Data[T] `json:"existing,omitempty"`
	New      []Data[T] `json:"new,omitempty"`
}

type SplitEntries[T, U any] struct {
	Existing []Data[T] `json:"existing,omitempty"`
	New      []U       `json:"new,omitempty"`
}
