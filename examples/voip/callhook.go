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
)

// setCallAckHandler registers the callback invoked for each <ack class="call">.
func setCallAckHandler(fn func(*waBinary.Node)) {
	callAckMu.Lock()
	callAckHandler = fn
	callAckMu.Unlock()
}

// installCallAckHook injects an "ack" entry into whatsmeow's nodeHandlers map.
// Call before Connect.
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
}
