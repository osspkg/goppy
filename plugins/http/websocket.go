package http

//go:generate easyjson

import (
	"context"
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

const (
	on  = 1
	off = 0
)

var (
	ErrServAlreadyRunning = errors.New("server already running")
	ErrServAlreadyStopped = errors.New("server already stopped")

	wsu = &websocket.Upgrader{
		EnableCompression: true,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin: func(_ *nethttp.Request) bool {
			return true
		},
	}
)

type WebsocketServerOption func(upg *websocket.Upgrader)

func WebsocketServerOptionCompression(enable bool) WebsocketServerOption {
	return func(upg *websocket.Upgrader) {
		upg.EnableCompression = enable
	}
}

func WebsocketServerOptionBuffer(read, write int) WebsocketServerOption {
	return func(upg *websocket.Upgrader) {
		upg.ReadBufferSize, upg.WriteBufferSize = read, write
	}
}

func WithWebsocketServer(options ...WebsocketServerOption) plugins.Plugin {
	return plugins.Plugin{
		Inject: func(log logger.Logger) (*websocketServerProvider, WebsocketServer) {
			for _, option := range options {
				option(wsu)
			}
			wsp := newWebsocketServerProvider(log)
			return wsp, wsp
		},
	}
}

type (
	websocketServerProvider struct {
		status  int64
		clients map[string]*conn
		events  map[uint]WebsocketServerHandler

		cm sync.RWMutex
		em sync.RWMutex

		log logger.Logger
	}

	WebsocketServerHandler func(d WebsocketEventer, c WebsocketServerProcessor) error

	WebsocketServer interface {
		Handling(ctx Ctx)
		Event(call WebsocketServerHandler, eid ...uint)
		Broadcast(t uint, m json.Marshaler)
		CloseAll()
		CountConn() int
	}
)

func newWebsocketServerProvider(log logger.Logger) *websocketServerProvider {
	return &websocketServerProvider{
		status:  off,
		clients: make(map[string]*conn),
		events:  make(map[uint]WebsocketServerHandler),
		log:     log,
	}
}

func (v *websocketServerProvider) Up() error {
	if !atomic.CompareAndSwapInt64(&v.status, off, on) {
		return ErrServAlreadyRunning
	}
	return nil
}

// Down hub
func (v *websocketServerProvider) Down() error {
	if !atomic.CompareAndSwapInt64(&v.status, on, off) {
		return ErrServAlreadyStopped
	}
	v.CloseAll()
	return nil
}

func (v *websocketServerProvider) Broadcast(t uint, m json.Marshaler) {
	eventModel(func(ev *event) {
		ev.ID = t

		b, err := m.MarshalJSON()
		if err != nil {
			v.log.WithFields(logger.Fields{
				"err": err.Error(),
			}).Errorf("[ws] Broadcast error")
			return
		}
		ev.Body(b)

		b, err = json.Marshal(ev)
		if err != nil {
			v.log.WithFields(logger.Fields{
				"err": err.Error(),
			}).Errorf("[ws] Broadcast error")
			return
		}

		v.cm.RLock()
		for _, c := range v.clients {
			c.Write(b)
		}
		v.cm.RUnlock()
	})
}

func (v *websocketServerProvider) CloseAll() {
	for _, c := range v.clients {
		c.Close()
	}
}

func (v *websocketServerProvider) Event(call WebsocketServerHandler, eid ...uint) {
	v.em.Lock()
	defer v.em.Unlock()

	for _, i := range eid {
		v.events[i] = call
	}
}

func (v *websocketServerProvider) addConn(c *conn) {
	v.cm.Lock()
	v.clients[c.CID()] = c
	v.cm.Unlock()
}

func (v *websocketServerProvider) delConn(id string) {
	v.cm.Lock()
	delete(v.clients, id)
	v.cm.Unlock()
}

func (v *websocketServerProvider) CountConn() int {
	v.cm.Lock()
	cc := len(v.clients)
	v.cm.Unlock()
	return cc
}

func (v *websocketServerProvider) getEventHandler(id uint) (WebsocketServerHandler, bool) {
	v.em.RLock()
	defer v.em.RUnlock()

	fn, ok := v.events[id]
	return fn, ok
}

func (v *websocketServerProvider) Handling(ctx Ctx) {
	c := newWebsocketServerConn(ctx.Context())

	c.OnError(func(cid string, err error, msg string, args ...interface{}) {
		if err == nil {
			return
		}
		v.log.WithFields(logger.Fields{
			"cid": cid,
			"err": err.Error(),
		}).Errorf(msg, args...)
	})

	c.GetEventCall(v.getEventHandler)

	c.OnClose(func(cid string) {
		v.delConn(cid)
	})
	c.OnOpen(func(string) {
		v.addConn(c)
	})

	if err := c.Upgrade(ctx.Response(), ctx.Request(), *wsu); err != nil {
		ctx.SetBody(nethttp.StatusBadRequest).Error(err)
		return
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	poolEvent = sync.Pool{New: func() interface{} { return &event{} }}
)

//easyjson:json
type event struct {
	ID      uint            `json:"e"`
	Data    json.RawMessage `json:"d"`
	Err     *string         `json:"err,omitempty"`
	UID     json.RawMessage `json:"u,omitempty"`
	Updated bool            `json:"-"`
}

func (v *event) EventID() uint {
	return v.ID
}

func (v *event) UniqueID() []byte {
	if v.UID == nil {
		return nil //internal.RandByte(10)
	}
	result := make([]byte, 0, len(v.UID))
	return append(result, v.UID...)
}

func (v *event) Decode(in interface{}) error {
	return json.Unmarshal(v.Data, in)
}

func (v *event) Encode(in interface{}) {
	b, err := json.Marshal(in)
	if err != nil {
		v.Error(err)
		return
	}
	v.Body(b)
}

func (v *event) Reset() *event {
	v.ID, v.Err, v.UID, v.Data, v.Updated = 0, nil, nil, v.Data[:0], false
	return v
}

func (v *event) Error(e error) {
	if e == nil {
		return
	}
	err := e.Error()
	v.Err, v.Data, v.Updated = &err, v.Data[:0], true
}

func (v *event) Body(b []byte) {
	v.Err, v.Data, v.Updated = nil, append(v.Data[:0], b...), true
}

func eventModel(call func(ev *event)) {
	m, ok := poolEvent.Get().(*event)
	if !ok {
		m = &event{}
	}
	call(m)
	poolEvent.Put(m.Reset())
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	pongWait   = 60 * time.Second
	pingPeriod = pongWait / 3
)

type (
	conn struct {
		status    int64
		cid       string
		ctx       context.Context
		cancel    context.CancelFunc
		onClose   []func(cid string)
		onOpen    []func(cid string)
		onError   func(err error, msg string, args ...interface{})
		sendC     chan []byte
		conn      *websocket.Conn
		eventCall func(id uint) (WebsocketServerHandler, bool)
		wg        sync.WaitGroup
		cm        sync.RWMutex
	}

	WebsocketServerProcessor interface {
		CID() string
		OnClose(cb func(cid string))
		OnOpen(cb func(cid string))
		Encode(eventID uint, in interface{})
		EncodeEvent(event WebsocketEventer, in interface{})
	}

	WebsocketEventer interface {
		EventID() uint
		UniqueID() []byte
		Decode(in interface{}) error
	}
)

func newWebsocketServerConn(ctx context.Context) *conn {
	c, cncl := context.WithCancel(ctx)
	return &conn{
		onClose: make([]func(string), 0),
		onOpen:  make([]func(string), 0),
		sendC:   make(chan []byte, 128),
		ctx:     c,
		cancel:  cncl,
		status:  on,
	}
}

func (v *conn) CID() string {
	return v.cid
}

func (v *conn) GetEventCall(e func(id uint) (WebsocketServerHandler, bool)) {
	v.eventCall = e
}

func (v *conn) OnClose(cb func(cid string)) {
	v.cm.Lock()
	defer v.cm.Unlock()

	v.onClose = append(v.onClose, cb)
}

func (v *conn) OnOpen(cb func(cid string)) {
	v.cm.Lock()
	defer v.cm.Unlock()

	v.onOpen = append(v.onOpen, cb)
}

func (v *conn) OnError(cb func(cid string, err error, msg string, args ...interface{})) {
	v.onError = func(err error, msg string, args ...interface{}) {
		cb(v.cid, err, msg, args...)
	}
}

func (v *conn) Close() {
	if !atomic.CompareAndSwapInt64(&v.status, on, off) {
		return
	}

	v.cm.RLock()
	for _, fn := range v.onClose {
		fn(v.CID())
	}
	v.cm.RUnlock()

	if v.conn != nil {
		v.cancel()
		v.wg.Wait()
		v.onError(v.conn.Close(), "[ws] close conn")
	}
}

func (v *conn) Encode(eventID uint, in interface{}) {
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

func (v *conn) EncodeEvent(e WebsocketEventer, in interface{}) {
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

func (v *conn) Write(b []byte) {
	if len(b) == 0 {
		return
	}

	select {
	case v.sendC <- b:
	default:
	}
}

func (v *conn) pumpWrite() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		v.wg.Done()
		v.Close()
	}()
	for {
		select {
		case <-v.ctx.Done():
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

func (v *conn) pumpRead() {
	defer func() {
		v.wg.Done()
		v.Close()
	}()
	for {
		_, message, err := v.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, 1000, 1001, 1005) {
				v.onError(err, "[ws] read message")
			}
			return
		}
		go v.processor(message)
	}
}

func (v *conn) processor(b []byte) {
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
		call, ok := v.eventCall(ev.EventID())
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

func (v *conn) Upgrade(w nethttp.ResponseWriter, r *nethttp.Request, up websocket.Upgrader) (err error) {
	v.conn, err = up.Upgrade(w, r, nil)
	if err != nil {
		v.onError(err, "[ws] upgrade")
		v.Close()
		return
	}

	v.cid = r.Header.Get("Sec-Websocket-Key")

	v.conn.SetPingHandler(func(string) error {
		return errors.Wrap(
			v.conn.SetReadDeadline(time.Now().Add(pongWait)),
			v.conn.SetWriteDeadline(time.Now().Add(pongWait)),
		)
	})
	v.conn.SetPongHandler(func(string) error {
		return errors.Wrap(
			v.conn.SetReadDeadline(time.Now().Add(pongWait)),
			v.conn.SetWriteDeadline(time.Now().Add(pongWait)),
		)
	})

	v.wg.Add(2)

	go v.pumpWrite()
	go v.pumpRead()

	v.cm.RLock()
	for _, fn := range v.onOpen {
		fn(v.CID())
	}
	v.cm.RUnlock()

	v.wg.Wait()
	return
}
