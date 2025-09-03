package event_test

import (
	"sort"
	"testing"

	"github.com/dv-net/dv-merchant/internal/event"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHandlerIdLess(t *testing.T) {
	id1 := event.HandlerID(uuid.MustParse("ef6bf928-1c1b-11ef-a710-d732d21cc62c"))
	id2 := event.HandlerID(uuid.MustParse("fc6caa3c-1c1b-11ef-8b05-1fc5d1cc0da8"))
	id3 := event.HandlerID(uuid.MustParse("ed21fc52-1c1c-11ef-898e-c7e69930f57c"))
	list := event.HandlerIDList{id1, id2}
	assert.True(t, list.Less(0, 1))
	{
		index, ok := sort.Find(list.Len(), func(i int) int {
			return -list.Cmp(i, id1)
		})
		assert.True(t, ok)
		assert.Equal(t, 0, index)
	}
	{
		index, ok := sort.Find(list.Len(), func(i int) int {
			return -list.Cmp(i, id2)
		})
		assert.True(t, ok)
		assert.Equal(t, 1, index)
	}
	{
		_, ok := sort.Find(list.Len(), func(i int) int {
			return -list.Cmp(i, id3)
		})
		assert.False(t, ok)
	}
}
