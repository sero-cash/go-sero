package share

import (
	"container/heap"
	"sort"
	"sync"
	"time"

	"github.com/sero-cash/go-sero/core/types"
)

type lotteryItem struct {
	lottery  *types.Lottery
	attempts uint8
}

type Item struct {
	Key   interface{} //the unique key of item
	Value interface{} //the value of the item
	Time  time.Time   //the priority of the item in the queue

	//heap.Interface need this index and update them
	Index int //index of the item in the heap
}

type ItemSlice struct {
	items    []*Item
	itemsMap map[interface{}]*Item //find item according to the key (inteface {} type)
}

func (s ItemSlice) Len() int { return len(s.items) }
func (s ItemSlice) Less(i, j int) bool {
	return s.items[i].Time.Before(s.items[i].Time)
}

func (s ItemSlice) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
	s.items[i].Index = i
	s.items[j].Index = j
	if s.itemsMap != nil {
		s.itemsMap[s.items[i].Key] = s.items[i]
		s.itemsMap[s.items[j].Key] = s.items[j]
	}
}

func (s *ItemSlice) Push(x interface{}) {
	n := len(s.items)
	item := x.(*Item)
	item.Index = n
	s.items = append(s.items, item)
	s.itemsMap[item.Key] = item
}

func (s *ItemSlice) Pop() interface{} {
	old := s.items
	n := len(old)
	item := old[n-1]
	item.Index = -1
	delete(s.itemsMap, item.Key)
	s.items = old[0 : n-1]
	return item
}

func (s *ItemSlice) Update(key interface{}, value interface{}, time time.Time) {
	item := s.itemByKey(key)
	if item != nil {
		s.updateItem(item, value, time)
	}

}

func (s *ItemSlice) itemByKey(key interface{}) *Item {
	if item, found := s.itemsMap[key]; found {
		return item
	}
	return nil
}

func (s *ItemSlice) updateItem(item *Item,
	value interface{}, time time.Time) {
	item.Value = value
	item.Time = time
	heap.Fix(s, item.Index)
}

type PriorityQueue struct {
	slice   ItemSlice
	maxSize int
	mutex   sync.RWMutex
}

func (pq *PriorityQueue) Init(maxSize int) {
	pq.slice.items = make([]*Item, 0, pq.maxSize)
	pq.slice.itemsMap = make(map[interface{}]*Item)
	pq.maxSize = maxSize
}

func (pq PriorityQueue) Len() int {
	pq.mutex.RLock()
	size := pq.slice.Len()
	pq.mutex.RUnlock()
	return size
}

func (pq *PriorityQueue) minItem() *Item {
	len := pq.slice.Len()
	if len == 0 {
		return nil
	}
	return pq.slice.items[0]
}

func (pq *PriorityQueue) MinItem() *Item {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()
	return pq.minItem()
}

func (pq *PriorityQueue) PushItem(key, value interface{},
	time time.Time) (bPushed bool) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	size := pq.slice.Len()
	item := pq.slice.itemByKey(key)
	if size > 0 && item != nil {
		pq.slice.updateItem(item, value, time)
		return true
	}
	item = &Item{
		Value: value,
		Key:   key,
		Time:  time,
		Index: -1,
	}
	if pq.maxSize <= 0 || size < pq.maxSize {
		heap.Push(&(pq.slice), item)
		return true
	}
	heap.Pop(&(pq.slice))
	heap.Push(&(pq.slice), item)
	return true
}

func (pq *PriorityQueue) PopItem() interface{} {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	sz := pq.slice.Len()
	if sz > 0 {
		return heap.Pop(&(pq.slice)).(*Item).Value
	} else {
		return nil
	}
}

func (pq PriorityQueue) GetQueue() []interface{} {
	items := pq.GetQueueItems()
	values := make([]interface{}, len(items))
	for i := 0; i < len(items); i++ {
		values[i] = items[i].Value
	}
	return values
}

func (pq PriorityQueue) GetQueueItems() []*Item {
	size := pq.Len()
	if size == 0 {
		return []*Item{}
	}
	s := ItemSlice{}
	s.items = make([]*Item, size)
	pq.mutex.RLock()
	for i := 0; i < size; i++ {
		s.items[i] = &Item{
			Value: pq.slice.items[i].Value,
			Time:  pq.slice.items[i].Time,
		}
	}
	pq.mutex.RUnlock()
	sort.Sort(s)
	return s.items
}
