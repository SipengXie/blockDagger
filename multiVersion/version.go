package multiversion

import (
	"sync"
)

type Status int

const (
	Pending   Status = iota // 尚未执行的版本
	Committed               // 在执行过程中被提交的版本
	Ignore                  // 在执行过程中中止的版本
)

type Version struct {
	Data   interface{} // 可以是Balance, Nonce, Code, StorageSlots等
	Tid    int
	Status Status
	Next   *Version
	Prev   *Version

	// 为Gria多线程操作准备的变量
	Plock sync.Mutex
	Nlock sync.Mutex

	// 为Gria调度准备的变量
	Readby    map[int]struct{}
	MaxReadby int
}

func NewVersion(data interface{}, tid int, status Status) *Version {
	return &Version{
		Data:      data,
		Tid:       tid,
		Status:    status,
		Readby:    make(map[int]struct{}),
		MaxReadby: -1,
		Next:      nil,
		Prev:      nil,
		Plock:     sync.Mutex{},
		Nlock:     sync.Mutex{},
	}
}

func (v *Version) InsertOrNext(iv *Version) *Version {
	v.Nlock.Lock()
	defer v.Nlock.Unlock()
	if v.Next == nil || v.updatePrev(iv) {
		iv.Next = v.Next
		v.Next = iv
		iv.Prev = v
		return nil
	} else {
		return v.Next
	}
}

func (v *Version) updatePrev(iv *Version) bool {
	v.Plock.Lock()
	defer v.Plock.Unlock()
	if iv.Tid < v.Tid {
		v.Prev = iv
		return true
	}
	return false
}

func (v *Version) GetVisible() *Version {
	if v.Status != Committed {
		// fmt.Println(v)
		return v.Prev.GetVisible()
	}
	return v
}
