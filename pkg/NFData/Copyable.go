package NFData

type Copyable interface {
	Copy() Copyable
}
