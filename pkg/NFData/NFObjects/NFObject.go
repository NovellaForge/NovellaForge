package NFObjects

import "go.novellaforge.dev/novellaforge/pkg/NFData"

type NFObject interface {
	GetArgs() *NFData.NFInterfaceMap
	SetArgs(args *NFData.NFInterfaceMap)
	GetType() string
	SetType(t string)
	AddChild(object NFObject)
	DeleteChild(name string) error
}
