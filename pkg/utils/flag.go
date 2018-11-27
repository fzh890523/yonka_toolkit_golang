package utils


type FlagPart interface {
	RegisterFlag()
	ParseFlag() error
}

type DummyFlagPart struct {
}

func (c *DummyFlagPart) RegisterFlag() {
}

func (c *DummyFlagPart) ParseFlag() (err error) {
	return
}
