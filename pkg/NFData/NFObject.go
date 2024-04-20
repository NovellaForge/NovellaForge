package NFData

type NFObject interface {
	GetArgs() *NFInterfaceMap
	SetArgs(args *NFInterfaceMap)
	GetType() string
	SetType(t string)
}
