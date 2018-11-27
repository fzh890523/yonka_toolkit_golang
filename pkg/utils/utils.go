package utils

import (
	"container/list"
	"errors"
	"time"
	log "github.com/golang/glog"
	"math"
	"fmt"
	"unsafe"
	"strings"
)

const (
	VisitStrategyBreak    = iota
	VisitStrategyContinue
	VisitStrategyRollback
)

var ErrInvalidEnum = errors.New("invalid enum")

func VisitList1(l *list.List, visitStrategy int, rollback func(item *list.Element) error, visitor func(item *list.Element) error) (err error) {
	var item *list.Element = nil

outer:
	for item = l.Front(); item != nil; item = item.Next() {
		if curErr := visitor(item); curErr != nil {
			log.Error("visit item (%v) met error (%v)", item, err)

			if err == nil {
				err = curErr
			}

			switch visitStrategy {
			case VisitStrategyBreak, VisitStrategyRollback:
				break outer
			case VisitStrategyContinue:
				continue
			default:
				err = ErrInvalidEnum
				return
			}
		}
	}

	if visitStrategy == VisitStrategyRollback && err != nil && item != nil && rollback != nil {
		for ; item != nil; item = item.Prev() {
			if curErr := rollback(item); curErr != nil {
				log.Error("rollback item (%v) met error (%v)", item, err)
			}
		}
	}

	return err
}

func VisitList(l *list.List, visitor func(item *list.Element) error) (err error) {
	return VisitList1(l, VisitStrategyBreak, nil, visitor)
}

type ChanLock struct {
	ch chan struct{}
}

func NewChanLock() *ChanLock {
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	return &ChanLock{ch: ch}
}

func (l *ChanLock) Lock() {
	<-l.ch
}

func (l *ChanLock) LockWithTimeout(d time.Duration) bool {
	select {
	case <-l.ch:
		return true
	case <-time.NewTimer(d).C:
		return false
	}
}

func (l *ChanLock) TryLock(d time.Duration) bool {
	select {
	case <-l.ch:
		return true
	default:
		return false
	}
}

var ErrStopIteration = errors.New("stop iteration")
var errChanIteratorNoChan = errors.New("chan iterator no chan")
var errChanIteratorAlreadyPrepared = errors.New("chan interator already prepared")

type Iterator interface {
	Next() (value interface{}, err error)
	HasNext() (ok bool, err error)
}

type NonPredicableIterator interface {
	Next() (value interface{}, err error)
}

type ChanIterator struct {
	ch chan interface{}
	f  func() error
}

func NewChanIterator() ChanIterator {
	ch := make(chan interface{}) // no buf
	res := ChanIterator{ch: ch}

	return res
}

func (i ChanIterator) Prepare(f func() error) error {
	if i.ch == nil {
		return errChanIteratorNoChan
	}

	if i.f != nil {
		return errChanIteratorAlreadyPrepared
	}

	i.f = f

	go func() {
		err := f() // returned error ?
		_ = err
		close(i.ch)
	}()

	return nil
}

func (i ChanIterator) Channel() chan interface{} {
	return i.ch
}

func (i ChanIterator) Next() (value interface{}, err error) {
	if v, ok := <-i.ch; !ok {
		err = ErrStopIteration
		return
	} else {
		value = v
	}

	return
}

func Iterate1(it NonPredicableIterator, visitStrategy int, visitor func(value interface{}) error) (err error) {
	switch visitStrategy {
	case VisitStrategyBreak, VisitStrategyContinue:
	default:
		err = ErrInvalidEnum
		return
	}

	for {
		v, curErr := it.Next()
		if curErr != nil {
			if curErr == ErrStopIteration {
				break
			}

			log.Errorf("") // XXX log
			if err == nil {
				err = curErr
			}
		} else {
			err = visitor(v)
		}

		if err != nil {
			switch visitStrategy {
			case VisitStrategyBreak:
				break
			}
		}
	}

	return
}

func Iterate(it NonPredicableIterator, visitor func(value interface{}) error) (err error) {
	return Iterate1(it, VisitStrategyBreak, visitor)
}

func IntRangeInStep(from, to, step int) []int {
	diff := to - from
	count := int(math.Floor(math.Abs(float64(diff)/float64(step))) + 1)

	res := make([]int, 0, count)
	reversed := from > to

	if diff != 0 && diff*step/int(math.Abs(float64(diff))) <= 0 {
		panic(fmt.Sprintf("invalid arg: %d-%d-%d", from, to, step))
	}

	for i := from; ; {
		if reversed {
			if i <= to {
				break
			}
		} else {
			if i >= to {
				break
			}
		}

		res = append(res, i)

		i += step
	}

	return res
}

func IntRange(from, to int) []int {
	step := 1
	if from > to {
		step = -step
	}
	return IntRangeInStep(from, to, step)
}

func ModInt32Pointer(addr *int32, mod int32) int32 {
	return *((*int32)(unsafe.Pointer(addr))) % mod // unsafe
}

func CompareStrings(s1, s2 []string) int {
	l1 := len(s1)
	l2 := len(s2)

	l := l1
	if l2 < l {
		l = l2
	}

	for i := 0; i < l; i++ {
		if res := strings.Compare(s1[i], s2[i]); res != 0 {
			return res
		}
	}

	res := l1 - l2
	if res == 0 {
		return 0
	} else {
		return res / int(math.Abs(float64(res)))
	}
}

func CheckRangeOverlap(rangeLow, rangeHigh int, checkRangeLow, checkRangeHigh int) bool {
	if rangeHigh <= checkRangeLow || rangeLow >= checkRangeHigh {
		return false
	}

	return true
}

// low == high means empty-range
// TODO add unit test
func CutRangeOverlap(rangeLow, rangeHigh int, checkRangeLow, checkRangeHigh int) (low, high int) {
	if rangeLow >= rangeHigh {
		goto empty
	}

	if checkRangeLow >= checkRangeHigh {
		goto original
	}

	switch {
	case rangeLow < checkRangeLow:
		{
			switch {
			case rangeHigh <= checkRangeLow:
				goto original // totally left side
			case rangeHigh <= checkRangeHigh:
				low, high = rangeLow, checkRangeLow
			default:
				// choose a side = =
				if rangeHigh-checkRangeHigh > checkRangeLow-rangeLow {
					low, high = checkRangeHigh, rangeHigh
				} else {
					low, high = rangeLow, checkRangeLow
				}
			}
		}
	case rangeLow < checkRangeHigh:
		{
			switch {
			case rangeHigh <= checkRangeHigh:
				goto empty
			default:
				low, high = checkRangeHigh, rangeHigh
			}
		}
	default: // totally right side
		goto original
	}

	return

original:
	low, high = rangeLow, rangeHigh
	return

empty:
	low, high = rangeLow, rangeLow
	return
}

func ErrAsFalse(v bool, err error) (res bool) {
	if v && err != nil {
		res = true
	}

	return
}
