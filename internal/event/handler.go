package event

import "github.com/google/uuid"

type (
	HandlerID     uuid.UUID // Event handler ID
	HandlerIDList []HandlerID
	Handler       func(IEvent) error // Event handler type
)

func (hid HandlerID) String() string {
	return uuid.UUID(hid).String()
}

func (list HandlerIDList) Len() int {
	return len(list)
}

func (list HandlerIDList) Cmp(i int, id HandlerID) int {
	var (
		iID = uuid.UUID(list[i])
		jID = uuid.UUID(id)
	)
	for idx, iByte := range iID {
		if diff := int(iByte) - int(jID[idx]); diff != 0 {
			return diff
		}
	}
	return 0
}

func (list HandlerIDList) Less(i, j int) bool {
	return list.Cmp(i, list[j]) < 0
}

func (list HandlerIDList) Swap(i, j int) {
	iID := list[i]
	jID := list[j]
	list[i] = jID
	list[j] = iID
}
