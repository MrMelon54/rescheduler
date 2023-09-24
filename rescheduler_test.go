package rescheduler

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRescheduler_SingleRun(t *testing.T) {
	const totalRuns = 4
	for i := 0; i < totalRuns; i++ {
		a := new(int)
		r := NewRescheduler(func() {
			time.Sleep(time.Millisecond * 100)
			*a++
		})

		for j := 0; j < i+1; j++ {
			r.Run()
		}

		// wait for all calls to finish running
		r.Wait()
		// only runs twice (due to consecutive calls)
		exp := 2
		if i == 0 {
			exp = 1
		}
		assert.Equal(t, exp, *a, "Error on %d runs", i+1)
	}
}

func TestRescheduler_TimeGap(t *testing.T) {
	const totalRuns = 4
	for i := 0; i < totalRuns; i++ {
		a := new(int)
		r := NewRescheduler(func() {
			time.Sleep(time.Millisecond * 100)
			*a++
		})

		for j := 0; j < i+1; j++ {
			r.Run()
		}
		time.Sleep(time.Millisecond * 300)
		for j := 0; j < i+1; j++ {
			r.Run()
		}
		time.Sleep(time.Millisecond * 300)
		for j := 0; j < i+1; j++ {
			r.Run()
		}
		time.Sleep(time.Millisecond * 300)
		for j := 0; j < i+1; j++ {
			r.Run()
		}

		// wait for all calls to finish running
		r.Wait()
		// only runs twice (due to consecutive calls)
		exp := 8
		if i == 0 {
			exp = 4
		}
		assert.Equal(t, exp, *a, "Error on %d runs", i+1)
	}
}

func TestRescheduler_FinishedBeforeWait(t *testing.T) {
	r := NewRescheduler(func() {
		time.Sleep(time.Millisecond * 100)
	})

	r.Run()

	time.Sleep(time.Millisecond * 200)

	select {
	case <-r.done:
	default:
		t.Fatal("Should receive from done channel now")
	}
}
