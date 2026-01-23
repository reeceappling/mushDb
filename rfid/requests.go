package rfid

type Data[T any] struct {
	Data     T    `json:"data"`
	Disabled bool `json:"disabled"`
}

func DataStripped[T any](data Data[T]) T {
	return data.Data
}

type AllEntries[T any] struct {
	Existing []Data[T] `json:"existing,omitempty"`
	New      []Data[T] `json:"new,omitempty"`
}

func (ae AllEntries[T]) asEntries() []T {
	out := make([]T, 0, len(ae.Existing)+len(ae.New))
	for _, n := range ae.Existing {
		if !n.Disabled {
			out = append(out, n.Data)
		}
	}
	for _, n := range ae.New {
		out = append(out, n.Data)
	}
	return out
}

type SplitEntries[T, U any] struct {
	Existing []Data[T] `json:"existing,omitempty"`
	New      []U       `json:"new,omitempty"`
}
