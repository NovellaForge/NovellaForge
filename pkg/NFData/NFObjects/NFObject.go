package NFObjects

import "go.novellaforge.dev/novellaforge/pkg/NFData"

type NFObject interface {
	GetArgs() *NFData.NFInterfaceMap
	SetArgs(*NFData.NFInterfaceMap)
	GetType() string
	SetType(string)
	AddChild(NFObject)
	DeleteChild(string) error
	FetchChildren(map[string][]NFObject) map[string][]NFObject
}
