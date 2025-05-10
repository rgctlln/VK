package inmemory

import (
	"VK/internal/subpub"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestSubPub(t *testing.T) {
	t.Run("empty subpub", func(t *testing.T) {
		sp := NewSubPub()
		err := sp.Publish("some theme", "message")
		assert.ErrorIs(t, err, subpub.ErrNoSubscribers)
	})

	t.Run("invalid subject", func(t *testing.T) {
		sp := NewSubPub()
		_, err := sp.Subscribe("", func(msg interface{}) {
			fmt.Println("received post:", msg)
		})
		assert.Equal(t, err, subpub.ErrEmptySubject)
	})

	t.Run("valid subpub, ok publish", func(t *testing.T) {
		sp := NewSubPub()
		ch := make(chan interface{}, 1)

		_, err := sp.Subscribe("some theme", func(msg interface{}) {
			ch <- msg
		})
		assert.NoError(t, err)

		err = sp.Publish("some theme", "message")
		assert.NoError(t, err)

		select {
		case m := <-ch:
			assert.Equal(t, "message", m)
		case <-time.After(time.Second):
			t.Fatal("did not receive message")
		}
	})

	t.Run("ok unsubscribe", func(t *testing.T) {
		sp := NewSubPub()
		sub, err := sp.Subscribe("some theme", func(msg interface{}) {
			fmt.Println("received post:", msg)
		})
		assert.NoError(t, err)

		sub2, _ := sp.Subscribe("some theme", func(msg interface{}) {
			fmt.Println("received post:", msg)
		})
		assert.NoError(t, err)

		sub.Unsubscribe()
		assert.Len(t, sp.Subscriptions["some theme"], 1)

		sub2.Unsubscribe()
		assert.Len(t, sp.Subscriptions["some theme"], 0)

	})

	t.Run("ok close", func(t *testing.T) {
		sp := NewSubPub()
		_, err := sp.Subscribe("some theme", func(msg interface{}) {
			fmt.Println("received post:", msg)
		})
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = sp.Close(ctx)
		assert.NoError(t, err)

		_, err = sp.Subscribe("some theme", func(msg interface{}) {
			fmt.Println("received post:", msg)
		})
		assert.Equal(t, err, subpub.ErrClosed)
	})

	t.Run("ok concurrent subscribe and publish", func(t *testing.T) {
		sp := NewSubPub()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		defer sp.Close(ctx)

		goroutines := 10
		messagesPerGoroutine := 5
		done := make(chan struct{})
		count := 0
		mu := sync.Mutex{}

		_, err := sp.Subscribe("some theme", func(msg interface{}) {
			mu.Lock()
			count++
			mu.Unlock()
			if count == goroutines*messagesPerGoroutine {
				close(done)
			}
		})
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					err := sp.Publish("some theme", func(msg interface{}) {
						_ = fmt.Sprintf("msg %d from %d", j, id)
					})
					assert.NoError(t, err)
				}
			}(i)
		}

		wg.Wait()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("did not receive all messages in time")
		}

		mu.Lock()
		assert.Equal(t, goroutines*messagesPerGoroutine, count)
		mu.Unlock()
	})

	t.Run("ok two subscriptions with slow one", func(t *testing.T) {
		sb := NewSubPub()
		first := make([]int, 0)
		second := make([]int, 0)
		wg := new(sync.WaitGroup)
		mu := new(sync.Mutex)
		wg.Add(10)
		_, err := sb.Subscribe("some theme", func(msg any) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			first = append(first, msg.(int))
		})
		assert.NoError(t, err)
		_, err = sb.Subscribe("some theme", func(msg any) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			time.Sleep(1 * time.Second)
			second = append(second, msg.(int))
		})
		assert.NoError(t, err)

		for i := range 5 {
			err = sb.Publish("some theme", i)
			assert.NoError(t, err)
		}
		wg.Wait()

		resArr := []int{0, 1, 2, 3, 4}
		assert.Equal(t, resArr, first, "wrong fast subscriber")
		assert.Equal(t, resArr, second, "wrong slow subscriber")
	})

}
