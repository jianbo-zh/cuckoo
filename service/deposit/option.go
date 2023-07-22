package deposit

func (d *DepositService) Apply(opts ...Option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(d); err != nil {
			return err
		}
	}
	return nil
}

type Option func(cfg *DepositService) error

func WithService() Option {
	return func(d *DepositService) error {
		d.isEnableService = true
		return nil
	}
}
