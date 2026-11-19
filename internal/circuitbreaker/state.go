package circuitbreaker

type State uint8

const (
	Closed State = iota
	Open
	HalfOpen
)
