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

func (ae AllEntries[T]) asEntries() []T {
	out := make([]T, len(ae.Existing)+len(ae.New))
	for _, n := range ae.Existing {
		if !n.disabled {
			out = append(out, n.data)
		}
	}
	for _, n := range ae.New {
		out = append(out, n.data)
	}
	return out
}

type SplitEntries[T, U any] struct {
	Existing []Data[T] `json:"existing,omitempty"`
	New      []U       `json:"new,omitempty"`
}
