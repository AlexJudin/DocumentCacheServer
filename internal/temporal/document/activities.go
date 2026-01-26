package document

type Activities struct {
}

func (a *Activities) NewActivity() *Activities {
	return &Activities{}
}
