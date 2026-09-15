package torrshash

type TorrsHash struct {
	Hash   string   `json:"hash"`
	Fields []*Field `json:"fields"`
}

func New(hash string) *TorrsHash {
	th := &TorrsHash{}
	th.Hash = hash
	return th
}

func (th *TorrsHash) AddField(tag Tag, value string) {
	th.Fields = append(th.Fields, &Field{tag, value})
}

func (h *TorrsHash) Title() string {
	for _, f := range h.Fields {
		if f.Tag == TagTitle {
			return f.Value
		}
	}
	return ""
}

func (h *TorrsHash) Poster() string {
	for _, f := range h.Fields {
		if f.Tag == TagPoster {
			return f.Value
		}
	}
	return ""
}

func (h *TorrsHash) Category() string {
	for _, f := range h.Fields {
		if f.Tag == TagCategory {
			return f.Value
		}
	}
	return ""
}

func (h *TorrsHash) Trackers() []string {
	var list []string
	for _, f := range h.Fields {
		if f.Tag == TagTracker {
			list = append(list, f.Value)
		}
	}
	return list
}

func (h *TorrsHash) String() string {
	str := "Hash: " + h.Hash

	for _, f := range h.Fields {
		str += " " + f.Tag.String() + ": " + f.Value
	}
	return str
}
