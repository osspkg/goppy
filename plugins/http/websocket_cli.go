package http

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-errors"
	"github.com/deweppro/go-logger"
	"github.com/gorilla/websocket"
)

func WithWebsocketClient() plugins.Plugin {
	return plugins.Plugin{
		Inject: func(log logger.Logger) (*websocketClientProvider, WebsocketClient) {
			c := &websocketClientProvider{
				list: make(map[string]WebsocketClientConn),
				log:  log,
			}
			return c, c
		},
	}
}

type (
	websocketClientProvider struct {
		list map[string]WebsocketClientConn
		log  logger.Logger
		mux  sync.RWMutex
	}

	WebsocketClient interface {
		Create(url string, opts ...func(WebsocketClientOption)) (WebsocketClientConn, error)
	}
)

func (v *websocketClientProvider) Up() error {
	return nil
}

func (v *websocketClientProvider) Down() error {
	for _, cliConn := range v.list {
		cliConn.Close()
	}
	return nil
}

func (v *websocketClientProvider) add(cc WebsocketClientConn) {
	v.mux.Lock()
	v.list[cc.CID()] = cc
	v.mux.Unlock()
}

func (v *websocketClientProvider) del(cid string) {
	v.mux.Lock()
	delete(v.list, cid)
	v.mux.Unlock()
}

func (v *websocketClientProvider) Create(url string, opts ...func(WebsocketClientOption)) (WebsocketClientConn, error) {
	cc := &websocketClientConn{
		status:    on,
		url:       url,
		headers:   make(nethttp.Header),
		sendC:     make(chan []byte, 128),
		events:    make(map[uint]WebsocketClientHandler, 128),
		closeChan: make(chan struct{}),
		onClose:   make([]func(string), 0),
	}

	cc.OnError(func(url string, err error, msg string, args ...interface{}) {
		if err == nil {
			return
		}
		v.log.WithFields(logger.Fields{
			"url": url,
			"err": err.Error(),
		}).Errorf(msg, args...)
	})

	for _, opt := range opts {
		opt(cc)
	}

	if err := cc.exec(); err != nil {
		return nil, err
	}

	cc.OnClose(func(cid string) {
		v.del(cid)
	})
	v.add(cc)

	return cc, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	websocketClientConn struct {
		status int64
		cid    string

		url     string
		headers nethttp.Header
		conn    *websocket.Conn

		sendC  chan []byte
		events map[uint]WebsocketClientHandler

		closeChan chan struct{}
		onClose   []func(cid string)
		onError   func(err error, msg string, args ...interface{})

		wg sync.WaitGroup
		cm sync.RWMutex
		em sync.RWMutex
	}

	WebsocketClientOption interface {
		Header(key string, value string)
	}

	WebsocketClientHandler func(d WebsocketEventer, c WebsocketClientProcessor) error

	WebsocketClientProcessor interface {
		CID() string
		OnClose(cb func(cid string))
		Encode(eventID uint, in interface{})
		EncodeEvent(event WebsocketEventer, in interface{})
	}

	WebsocketClientConn interface {
		CID() string
		Event(call WebsocketClientHandler, eid ...uint)
		Encode(id uint, in interface{})
		Close()
	}
)

func (v *websocketClientConn) CID() string {
	return v.cid
}

func (v *websocketClientConn) Close() {
	if !atomic.CompareAndSwapInt64(&v.status, on, off) {
		return
	}

	v.cm.RLock()
	for _, fn := range v.onClose {
		fn(v.CID())
	}
	v.cm.RUnlock()

	if v.conn != nil {
		close(v.closeChan)
		v.wg.Wait()
		v.onError(v.conn.Close(), "close connect [%s]", v.url)
	}
}

func (v *websocketClientConn) OnClose(cb func(cid string)) {
	v.cm.Lock()
	defer v.cm.Unlock()

	v.onClose = append(v.onClose, cb)
}

func (v *websocketClientConn) OnError(cb func(string, error, string, ...interface{})) {
	v.onError = func(err error, msg string, args ...interface{}) {
		cb(v.url, err, msg, args...)
	}
}

func (v *websocketClientConn) Header(key string, value string) {
	v.headers.Set(key, value)
}

func (v *websocketClientConn) Event(call WebsocketClientHandler, eid ...uint) {
	v.em.Lock()
	defer v.em.Unlock()

	for _, i := range eid {
		v.events[i] = call
	}
}

func (v *websocketClientConn) getEventHandler(id uint) (WebsocketClientHandler, bool) {
	v.em.RLock()
	defer v.em.RUnlock()

	fn, ok := v.events[id]
	return fn, ok
}

func (v *websocketClientConn) Write(b []byte) {
	if len(b) == 0 {
		return
	}

	select {
	case v.sendC <- b:
	default:
	}
}

func (v *websocketClientConn) Encode(eventID uint, in interface{}) {
	eventModel(func(ev *event) {
		ev.ID = eventID
		ev.Encode(in)
		b, err := json.Marshal(ev)
		if err != nil {
			v.onError(err, "[ws] encode message: %d", eventID)
			return
		}
		v.Write(b)
	})
}

func (v *websocketClientConn) EncodeEvent(e WebsocketEventer, in interface{}) {
	eventModel(func(ev *event) {
		ev.ID = e.EventID()
		ev.UID = e.UniqueID()
		ev.Encode(in)
		b, err := json.Marshal(ev)
		if err != nil {
			v.onError(err, "[ws] encode message: %d", e.EventID())
			return
		}
		v.Write(b)
	})
}

func (v *websocketClientConn) exec() error {
	var err error
	var resp *nethttp.Response
	if v.conn, resp, err = websocket.DefaultDialer.Dial(v.url, v.headers); err != nil {
		atomic.CompareAndSwapInt64(&v.status, on, off)
		v.onError(err, "open connect [%s]", v.url)
		return err
	} else {
		v.cid = resp.Header.Get("Sec-WebSocket-Accept")
	}
	v.conn.SetPingHandler(func(d string) error {
		return errors.Wrap(
			v.conn.SetReadDeadline(time.Now().Add(pongWait)),
			v.conn.SetWriteDeadline(time.Now().Add(pongWait)),
		)
	})
	v.conn.SetPongHandler(func(d string) error {
		return errors.Wrap(
			v.conn.SetReadDeadline(time.Now().Add(pongWait)),
			v.conn.SetWriteDeadline(time.Now().Add(pongWait)),
		)
	})

	v.wg.Add(2)
	go v.pumpWrite()
	go v.pumpRead()

	return nil
}

func (v *websocketClientConn) pumpRead() {
	defer func() {
		v.wg.Done()
		v.Close()
	}()
	for {
		_, m, err := v.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, 1000, 1001, 1005) {
				v.onError(err, "[ws] read message")
			}
			return
		}
		go v.processor(m)
	}
}

func (v *websocketClientConn) pumpWrite() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		v.wg.Done()
		v.Close()
	}()
	for {
		select {
		case <-v.closeChan:
			m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Bye bye!")
			if err := v.conn.WriteMessage(websocket.CloseMessage, m); err != nil && err != websocket.ErrCloseSent {
				v.onError(err, "[ws] send close")
			}
			return
		case m := <-v.sendC:
			if err := v.conn.WriteMessage(websocket.TextMessage, m); err != nil {
				v.onError(err, "[ws] send message")
				return
			}
		case <-ticker.C:
			if err := v.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				v.onError(err, "[ws] send ping")
				return
			}
		}
	}
}

func (v *websocketClientConn) processor(b []byte) {
	eventModel(func(ev *event) {
		var (
			err error
			msg string
		)
		defer func() {
			if err != nil {
				v.onError(err, "[ws] "+msg)
			}
		}()
		if err = json.Unmarshal(b, ev); err != nil {
			msg = "decode message"
			return
		}
		call, ok := v.getEventHandler(ev.EventID())
		if !ok {
			return
		}
		err = call(ev, v)
		if err != nil {
			ev.Error(err)
			bb, er := json.Marshal(ev)
			if er != nil {
				msg = fmt.Sprintf("[ws] call event handler: %d", ev.EventID())
				err = errors.Wrap(err, er)
				return
			}
			err = nil
			v.Write(bb)
			return
		}
	})
}
