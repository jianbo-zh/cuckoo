package peer

func (peer *PeerService) Apply(opts ...Option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(peer); err != nil {
			return err
		}
	}
	return nil
}

type Option func(cfg *PeerService) error

func TestOption() Option {
	return func(group *PeerService) error {
		return nil
	}
}
