package mail

import "context"

type noopSender struct{}

func newNoopSender(_ Config) (Sender, error) {
	return noopSender{}, nil
}

func (noopSender) Send(_ context.Context, _ Message) error {
	return nil
}
