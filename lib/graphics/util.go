package graphics

type Initializer interface {
	Initialize()
}

type Closer interface {
	Close()
}

type RefCounter interface {
	Increment()
	Decrement()
}

type Moder interface {
	Mode() Enum
	SetMode(Enum)
}

type Providable interface {
	//Providable() bool
	//SetProvidable(bool)
	Provide(Provider)
}
