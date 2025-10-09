package tokens

type Token2 interface {
	tokenMark()
}

type Number struct {
	kind  string
	value string
}

func (n *Number) tokenMark() {}
