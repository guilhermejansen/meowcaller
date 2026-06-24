package main

import (
	"context"
	"reflect"
	"sync"
	"unsafe"

	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
)

// whatsmeow has no handler for <ack> nodes — it silently drops them (client.go's
// dispatch even special-cases `node.Tag != "ack"` to suppress the unhandled-node
// log). But an outbound call's relay allocation arrives only inside
// <ack class="call" type="offer"> (the server's response to our <offer>), so
// without intercepting it the caller never learns the relay endpoint and media
// never starts. We reach into the client's unexported nodeHandlers map and
// register an "ack" handler. Installed before Connect so the map write never
// races the receive loop.

var (
	callAckMu      sync.Mutex
	callAckHandler func(*waBinary.Node)
	callRawHandler func(*waBinary.Node) // sees the raw <call> node (with its stanza id) before whatsmeow processes it
)

// setCallAckHandler registers the callback invoked for each <ack class="call">.
func setCallAckHandler(fn func(*waBinary.Node)) {
	callAckMu.Lock()
	callAckHandler = fn
	callAckMu.Unlock()
}

// setCallRawHandler registers a callback that sees each raw <call> node before
// whatsmeow handles it — needed for the <call> stanza id, which the CallOffer
// event drops (it only carries the inner <offer>).
func setCallRawHandler(fn func(*waBinary.Node)) {
	callAckMu.Lock()
	callRawHandler = fn
	callAckMu.Unlock()
}

// installCallAckHook injects an "ack" entry into whatsmeow's nodeHandlers map and
// wraps the "call" handler so we also see the raw <call> node. Call before Connect.
func installCallAckHook(cli *whatsmeow.Client) {
	field := reflect.ValueOf(cli).Elem().FieldByName("nodeHandlers")
	handlers := *(*map[string]func(context.Context, *waBinary.Node))(unsafe.Pointer(field.UnsafeAddr()))
	handlers["ack"] = func(_ context.Context, node *waBinary.Node) {
		if node.AttrGetter().String("class") != "call" {
			return
		}
		callAckMu.Lock()
		fn := callAckHandler
		callAckMu.Unlock()
		if fn != nil {
			fn(node)
		}
	}
	origCall := handlers["call"]
	handlers["call"] = func(ctx context.Context, node *waBinary.Node) {
		callAckMu.Lock()
		raw := callRawHandler
		callAckMu.Unlock()
		if raw != nil {
			raw(node)
		}
		if origCall != nil {
			origCall(ctx, node)
		}
	}
}
