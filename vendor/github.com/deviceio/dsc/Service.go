package dsc

type Service struct {
	*Relation

	Absent     bool
	AbsentFunc func(*Service) (bool, error)

	Started     bool
	StartedFunc func(*Service) (bool, error)

	Name     string
	NameFunc func(*Service) (string, error)

	Description     string
	DescriptionFunc func(*Service) (string, error)

	Path     string
	PathFunc func(*Service) (string, error)

	Args     []string
	ArgsFunc func(*Service) ([]string, error)
}

func (t *Service) absent() (bool, error) {
	if t.AbsentFunc != nil {
		return t.AbsentFunc(t)
	}

	return t.Absent, nil
}

func (t *Service) started() (bool, error) {
	if t.StartedFunc != nil {
		return t.StartedFunc(t)
	}

	return t.Started, nil
}

func (t *Service) name() (string, error) {
	if t.NameFunc != nil {
		return t.NameFunc(t)
	}

	return t.Name, nil
}

func (t *Service) description() (string, error) {
	if t.DescriptionFunc != nil {
		return t.DescriptionFunc(t)
	}

	return t.Description, nil
}

func (t *Service) path() (string, error) {
	if t.PathFunc != nil {
		return t.PathFunc(t)
	}

	return t.Path, nil
}

func (t *Service) args() ([]string, error) {
	if t.ArgsFunc != nil {
		return t.ArgsFunc(t)
	}

	return t.Args, nil
}
