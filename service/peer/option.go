package peer

func (peer *PeerSvc) Apply(opts ...Option) error {
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

type Option func(cfg *PeerSvc) error

func TestOption() Option {
	return func(group *PeerSvc) error {
		return nil
	}
}
