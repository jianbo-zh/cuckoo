package groupsvc

func (group *GroupService) Apply(opts ...Option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(group); err != nil {
			return err
		}
	}
	return nil
}

type Option func(cfg *GroupService) error

func TestOption() Option {
	return func(group *GroupService) error {
		return nil
	}
}
