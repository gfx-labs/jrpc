package jrpctest

import (
	"context"
	"errors"
	"strings"
	"time"

	"gfx.cafe/open/jrpc/pkg/codec"

	"gfx.cafe/open/jrpc"
)

type testService struct{}

type EchoArgs struct {
	S string
}

type EchoResult struct {
	String string
	Int    int
	Args   *EchoArgs
}

type testError struct{}

func (testError) Error() string  { return "testError" }
func (testError) ErrorCode() int { return 444 }
func (testError) ErrorData() any { return "testError data" }

func (s *testService) NoArgsRets() {}

func (s *testService) EchoAny(n any) any {
	return n
}

func (s *testService) Echo(str string, i int, args *EchoArgs) EchoResult {
	return EchoResult{str, i, args}
}

func (s *testService) EchoWithCtx(ctx context.Context, str string, i int, args *EchoArgs) EchoResult {
	return EchoResult{str, i, args}
}

func (s *testService) PeerInfo(ctx context.Context) codec.PeerInfo {
	return jrpc.PeerInfoFromContext(ctx)
}

func (s *testService) Sleep(ctx context.Context, duration time.Duration) {
	time.Sleep(duration)
}

func (s *testService) Block(ctx context.Context) error {
	<-ctx.Done()
	return errors.New("context canceled in testservice_block")
}

func (s *testService) Rets() (string, error) {
	return "", nil
}

//lint:ignore ST1008 returns error first on purpose.
func (s *testService) InvalidRets1() (error, string) {
	return nil, ""
}

func (s *testService) InvalidRets2() (string, string) {
	return "", ""
}

func (s *testService) InvalidRets3() (string, string, error) {
	return "", "", nil
}

func (s *testService) ReturnError() error {
	return testError{}
}

func (s *testService) CallMeBack(ctx context.Context, method string, args []any) (any, error) {
	c, ok := jrpc.ConnFromContext(ctx)
	if !ok {
		return nil, errors.New("no client")
	}
	var result any
	err := c.Do(nil, &result, method, args)
	return result, err
}

func (s *testService) CallMeBackLater(ctx context.Context, method string, args []any) error {
	c, ok := jrpc.ConnFromContext(ctx)
	if !ok {
		return errors.New("no client")
	}
	go func() {
		<-ctx.Done()
		var result any
		c.Do(nil, &result, method, args)
	}()
	return nil
}

type notificationTestService struct {
	unsubscribed            chan string
	gotHangSubscriptionReq  chan struct{}
	unblockHangSubscription chan struct{}
}

func (s *notificationTestService) Echo(i int) int {
	return i
}

func (s *notificationTestService) Unsubscribe(subid string) {
	if s.unsubscribed != nil {
		s.unsubscribed <- subid
	}
}

//func (s *notificationTestService) SomeSubscription(ctx context.Context, n, val int) (*Subscription, error) {
//	notifier, supported := jrpc.NotifierFromContext(ctx)
//	if !supported {
//		return nil, jrpc.ErrNotificationsUnsupported
//	}
//
//	// By explicitly creating an subscription we make sure that the subscription id is send
//	// back to the client before the first subscription.Notify is called. Otherwise the
//	// events might be send before the response for the *_subscribe method.
//	subscription := notifier.CreateSubscription()
//	go func() {
//		for i := 0; i < n; i++ {
//			if err := notifier.Notify(subscription.ID, val+i); err != nil {
//				return
//			}
//		}
//		select {
//		case <-notifier.Closed():
//		case <-subscription.Err():
//		}
//		if s.unsubscribed != nil {
//			s.unsubscribed <- string(subscription.ID)
//		}
//	}()
//	return subscription, nil
//}
//
//// HangSubscription blocks on s.unblockHangSubscription before sending anything.
//func (s *notificationTestService) HangSubscription(ctx context.Context, val int) (*Subscription, error) {
//	notifier, supported := NotifierFromContext(ctx)
//	if !supported {
//		return nil, ErrNotificationsUnsupported
//	}
//	s.gotHangSubscriptionReq <- struct{}{}
//	<-s.unblockHangSubscription
//	subscription := notifier.CreateSubscription()
//
//	go func() {
//		notifier.Notify(subscription.ID, val)
//	}()
//	return subscription, nil
//}

// largeRespService generates arbitrary-size JSON responses.
type LargeRespService struct {
	Length int
}

func (x LargeRespService) LargeResp() string {
	return strings.Repeat("x", x.Length)
}
