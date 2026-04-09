package bus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const eventTimeout = time.Second

func recv(t *testing.T, ch <-chan any) any {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(eventTimeout):
		t.Fatal("timed out waiting for event")
		return nil
	}
}

func TestTopic_SingleSubscriber(t *testing.T) {
	topic := newTopic()
	ch, unsub := topic.subscribe()
	defer unsub()

	topic.publish("hello")

	assert.Equal(t, "hello", recv(t, ch))
}

func TestTopic_MultipleSubscribers(t *testing.T) {
	topic := newTopic()
	ch1, unsub1 := topic.subscribe()
	ch2, unsub2 := topic.subscribe()
	defer unsub1()
	defer unsub2()

	topic.publish(42)

	assert.Equal(t, 42, recv(t, ch1))
	assert.Equal(t, 42, recv(t, ch2))
}

func TestTopic_Unsubscribe(t *testing.T) {
	topic := newTopic()
	ch, unsub := topic.subscribe()

	unsub()

	_, open := <-ch
	assert.False(t, open, "channel should be closed after unsubscribe")
}

func TestTopic_UnsubscribedDoesNotReceive(t *testing.T) {
	topic := newTopic()
	ch, unsub := topic.subscribe()
	unsub()

	<-ch

	require.NotPanics(t, func() { topic.publish("after-unsub") })
}

func TestTopic_SubscriberOnlyReceivesAfterSubscribing(t *testing.T) {
	topic := newTopic()

	topic.publish(1)

	ch, unsub := topic.subscribe()
	defer unsub()

	topic.publish(2)

	assert.Equal(t, 2, recv(t, ch))
}

func TestTopic_IndependentSubscribers(t *testing.T) {
	topic := newTopic()
	ch1, unsub1 := topic.subscribe()
	ch2, unsub2 := topic.subscribe()
	defer unsub2()

	topic.publish("first")
	assert.Equal(t, "first", recv(t, ch1))
	assert.Equal(t, "first", recv(t, ch2))

	unsub1()
	<-ch1

	topic.publish("second")

	assert.Equal(t, "second", recv(t, ch2))
}

func TestTopic_SlowSubscriberDropsMessages(t *testing.T) {
	topic := newTopic()
	ch, unsub := topic.subscribe()
	defer unsub()

	for i := range defaultSubscriberBuffer + 1 {
		topic.publish(i)
	}

	count := 0
	for range defaultSubscriberBuffer {
		select {
		case <-ch:
			count++
		default:
		}
	}
	assert.Equal(t, defaultSubscriberBuffer, count)
}

func TestTopic_PublishToNoSubscribers(t *testing.T) {
	topic := newTopic()
	require.NotPanics(t, func() { topic.publish("nothing") })
}

func TestRegisterTopic_Duplicate(t *testing.T) {
	mb := New()
	require.NoError(t, RegisterTopic(mb, "x"))
	assert.Error(t, RegisterTopic(mb, "x"))
}

func TestSubscribe_NotFound(t *testing.T) {
	mb := New()
	sub := mb.NewSubscriber()
	defer sub.Close()
	assert.Error(t, sub.Subscribe("missing"))
}

func TestPublisher_RoundTrip(t *testing.T) {
	mb := New()
	require.NoError(t, RegisterTopic(mb, "t"))

	sub := mb.NewSubscriber()
	defer sub.Close()
	require.NoError(t, sub.Subscribe("t"))

	pub := mb.NewPublisher()
	require.NoError(t, pub.Publish("t", "hello"))

	assert.Equal(t, "hello", recv(t, sub.Events()))
}

func TestPublisher_UnknownTopic(t *testing.T) {
	mb := New()
	pub := mb.NewPublisher()
	assert.Error(t, pub.Publish("missing", "x"))
}

func TestSubscriber_MultipleTopics(t *testing.T) {
	mb := New()
	require.NoError(t, RegisterTopic(mb, "a"))
	require.NoError(t, RegisterTopic(mb, "b"))

	sub := mb.NewSubscriber()
	defer sub.Close()
	require.NoError(t, sub.Subscribe("a"))
	require.NoError(t, sub.Subscribe("b"))

	pub := mb.NewPublisher()
	_ = pub.Publish("a", "from-a")
	_ = pub.Publish("b", "from-b")

	got := map[any]bool{recv(t, sub.Events()): true, recv(t, sub.Events()): true}
	assert.True(t, got["from-a"])
	assert.True(t, got["from-b"])
}

func TestSubscriber_Close(t *testing.T) {
	mb := New()
	require.NoError(t, RegisterTopic(mb, "t"))

	sub := mb.NewSubscriber()
	require.NoError(t, sub.Subscribe("t"))
	sub.Close()

	_, open := <-sub.Events()
	assert.False(t, open)
}
