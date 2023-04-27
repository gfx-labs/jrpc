package subscription

//// handleSubscribe processes *_subscribe method calls.
//func (h *handler) handleSubscribe(cp *callProc, r *Request) *Response {
//	mw := NewReaderResponseWriterMsg(r.WithContext(cp.ctx))
//	switch h.peer.Transport {
//	case "http", "https":
//		mw.Send(nil, ErrNotificationsUnsupported)
//		return mw.Response()
//	}
//
//	// Subscription method name is first argument.
//	name, err := parseSubscriptionName(r.Params)
//	if err != nil {
//		mw.Send(nil, &invalidParamsError{err.Error()})
//		return mw.Response()
//	}
//	namespace := r.namespace()
//	has := h.reg.Match(NewRouteContext(), r.Method)
//	if !has {
//		mw.Send(nil, &subscriptionNotFoundError{namespace, name})
//		return mw.Response()
//	}
//	// Install notifier in context so the subscription handler can find it.
//	n := &Notifier{h: h, namespace: namespace, idgen: randomIDGenerator()}
//	cp.notifiers = append(cp.notifiers, n)
//	// now actually run the handler
//	req := r.WithContext(
//		context.WithValue(r.ctx, notifierKey{}, n),
//	)
//	mw = NewReaderResponseWriterMsg(req)
//	h.reg.ServeRPC(mw, req)
//	return mw.Response()
//}
//
//// parseSubscriptionName extracts the subscription name from an encoded argument array.
//func parseSubscriptionName(rawArgs json.RawMessage) (string, error) {
//	dec := json.NewDecoder(bytes.NewReader(rawArgs))
//	if tok, _ := dec.Token(); tok != json.Delim('[') {
//		return "", errors.New("non-array args")
//	}
//	v, _ := dec.Token()
//	method, ok := v.(string)
//	if !ok {
//		return "", errors.New("expected subscription name as first argument")
//	}
//	return method, nil
//}
//
//func (h *handler) unsubscribe(ctx context.Context, id SubID) (bool, error) {
//	h.subLock.Lock()
//	defer h.subLock.Unlock()
//
//	s := h.serverSubs[id]
//	if s == nil {
//		return false, ErrSubscriptionNotFound
//	}
//	close(s.err)
//	delete(h.serverSubs, id)
//	return true, nil
//}
//
//// cancelServerSubscriptions removes all subscriptions and closes their error channels.
//func (h *handler) cancelServerSubscriptions(err error) {
//	h.subLock.Lock()
//	defer h.subLock.Unlock()
//
//	for id, s := range h.serverSubs {
//		s.err <- err
//		close(s.err)
//		delete(h.serverSubs, id)
//	}
//}
//
//func (h *handler) addSubscriptions(nn []*Notifier) {
//	h.subLock.Lock()
//	defer h.subLock.Unlock()
//
//	for _, n := range nn {
//		if sub := n.takeSubscription(); sub != nil {
//			h.serverSubs[sub.ID] = sub
//		}
//	}
//}
