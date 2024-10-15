package tempolite

// Signals are notified on the database but store no data which is done in-memory
type SignalConsumerChannel[T any] struct {
	C chan T // when a signal is created or getted, the Producer and Consumer structs will share the same signal instance
}

type SignalProducerChannel[T any] struct {
	C chan T // when a signal is created or getted, the Producer and Consumer structs will share the same signal instance
}

func (s SignalConsumerChannel[T]) Close() {}

func ConsumeSignal[T any](ctx TempoliteContext, name string) (SignalConsumerChannel[T], error) {
	// use tempolite context to call functions from Tempolite to create/save (database) and return the signal channel
	// should use ctx to call a private function on *Tempolite to manage the signal
	return SignalConsumerChannel[T]{
		C: make(chan T, 1),
	}, nil
}

func ProduceSignal[T any](ctx TempoliteContext, name string) (SignalProducerChannel[T], error) {
	// use tempolite context to call functions from Tempolite to get (database) and return the signal channel
	// should use ctx to call a private function on *Tempolite to manage the signal
	return SignalProducerChannel[T]{
		C: make(chan T, 1),
	}, nil
}
