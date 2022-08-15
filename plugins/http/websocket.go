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
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

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
		return nil
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
		cid       string
		ctx       context.Context
		cancel    context.CancelFunc
		onClose   []func(cid string)
		onError   func(cid string, msg string, err error)
		sendC     chan []byte
		conn      *websocket.Conn
		headers   nethttp.Header
		eventCall eventFn
		wg        sync.WaitGroup
		status    int64
	}

	Processor interface {
		CID() string
		GetHeader(key string) string
		OnClose(cb func(cid string))
		Encode(eventID uint, in interface{})
		EncodeEvent(event Eventer, in interface{})
	}

	Eventer interface {
		EventID() uint
		UniqueID() []byte
		Decode(in interface{}) error
	}
)

func newConn(ctx context.Context) *conn {
	c, cncl := context.WithCancel(ctx)
	return &conn{
		cid:     uuid.NewString(),
		onClose: make([]func(string), 0),
		sendC:   make(chan []byte, 128),
		ctx:     c,
		cancel:  cncl,
		status:  on,
	}
}

func (v *conn) CID() string {
	return v.cid
}

func (v *conn) GetHeader(key string) string {
	if v.headers == nil {
		return ""
	}
	return v.headers.Get(key)
}

func (v *conn) GetEventCall(e eventFn) {
	v.eventCall = e
}

func (v *conn) OnClose(cb func(cid string)) {
	v.onClose = append(v.onClose, cb)
}

func (v *conn) OnError(cb func(cid string, msg string, err error)) {
	v.onError = cb
}

func (v *conn) Close() {
	if !atomic.CompareAndSwapInt64(&v.status, on, off) {
		return
	}
	for _, fn := range v.onClose {
		fn(v.CID())
	}
	if v.conn != nil {
		v.cancel()
		v.wg.Wait()
		if err := v.conn.Close(); err != nil {
			v.onError(v.CID(), "[ws] close conn", err)
		}
	}
}

func (v *conn) Encode(eventID uint, in interface{}) {
	eventModel(func(ev *event) {
		ev.ID = eventID
		ev.Encode(in)
		b, err := json.Marshal(ev)
		if err != nil {
			v.onError(v.CID(), fmt.Sprintf("[ws] encode message: %d", eventID), err)
			return
		}
		v.Write(b)
	})
}

func (v *conn) EncodeEvent(e Eventer, in interface{}) {
	eventModel(func(ev *event) {
		ev.ID = e.EventID()
		ev.UID = e.UniqueID()
		ev.Encode(in)
		b, err := json.Marshal(ev)
		if err != nil {
			v.onError(v.CID(), fmt.Sprintf("[ws] encode message: %d", e.EventID()), err)
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
				v.onError(v.CID(), "[ws] send close", err)
			}
			return
		case m := <-v.sendC:
			if err := v.conn.WriteMessage(websocket.TextMessage, m); err != nil {
				v.onError(v.CID(), "[ws] send message", err)
				return
			}
		case <-ticker.C:
			if err := v.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				v.onError(v.CID(), "[ws] send ping", err)
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
				v.onError(v.CID(), "[ws] read message", err)
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
				v.onError(v.CID(), "[ws] "+msg, err)
			}
		}()
		if err = json.Unmarshal(b, ev); err != nil {
			msg = "decode message"
			return
		}
		call, err := v.eventCall(ev.EventID())
		if err != nil {
			ev.Error(err)
			bb, er := json.Marshal(ev)
			if er != nil {
				msg = fmt.Sprintf("[ws] encode message: %d", ev.EventID())
				err = errors.Wrap(err, er)
				return
			}
			err = nil
			v.Write(bb)
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
		v.onError(v.CID(), "[ws] upgrade", err)
		v.Close()
		return
	}

	v.headers = r.Header
	if err = v.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		v.onError(v.CID(), "[ws] send pong", err)
		v.Close()
		return
	}
	v.conn.SetPongHandler(func(string) error {
		return v.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	v.wg.Add(2)
	go v.pumpWrite()
	go v.pumpRead()
	v.wg.Wait()
	return
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	on  = 1
	off = 0
)

var (
	ErrServAlreadyRunning = errors.New("server already running")
	ErrServAlreadyStopped = errors.New("server already stopped")
	ErrUnknownEventID     = errors.New("unknown event id")

	wsu = &websocket.Upgrader{
		EnableCompression: true,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin: func(_ *nethttp.Request) bool {
			return true
		},
	}
)

type OptionWS func(upg *websocket.Upgrader)

func OptionWSCompression(enable bool) OptionWS {
	return func(upg *websocket.Upgrader) {
		upg.EnableCompression = enable
	}
}

func OptionWSBuffer(read, write int) OptionWS {
	return func(upg *websocket.Upgrader) {
		upg.ReadBufferSize, upg.WriteBufferSize = read, write
	}
}

func OptionWSCheckOrigin(origin ...string) OptionWS {
	return func(upg *websocket.Upgrader) {
		if len(origin) == 0 {
			return
		}
		list := make(map[string]struct{}, len(origin))
		for _, s := range origin {
			list[s] = struct{}{}
		}
		upg.CheckOrigin = func(r *nethttp.Request) bool {
			o := r.Header.Get("origin")
			if len(o) == 0 {
				return false
			}
			_, ok := list[o]
			return ok
		}
	}
}

func WithWebsocket(options ...OptionWS) plugins.Plugin {
	return plugins.Plugin{
		Inject: func(log logger.Logger) (*wsProvider, WebSocket) {
			for _, option := range options {
				option(wsu)
			}
			wsp := newWsProvider(log)
			return wsp, wsp
		},
	}
}

type (
	wsProvider struct {
		status  int64
		clients map[string]*conn
		events  map[uint]Handler

		cm sync.RWMutex
		em sync.RWMutex

		log logger.Logger
	}

	Handler func(d Eventer, c Processor) error
	eventFn func(id uint) (Handler, error)

	WebSocket interface {
		Handling(ctx Ctx)
		Event(call Handler, eid ...uint)
		Broadcast(t uint, m json.Marshaler)
		CloseAll()
		CountConn() int
	}
)

func newWsProvider(log logger.Logger) *wsProvider {
	return &wsProvider{
		status:  off,
		clients: make(map[string]*conn),
		events:  make(map[uint]Handler),
		log:     log,
	}
}

func (v *wsProvider) Up() error {
	if !atomic.CompareAndSwapInt64(&v.status, off, on) {
		return ErrServAlreadyRunning
	}
	return nil
}

//Down hub
func (v *wsProvider) Down() error {
	if !atomic.CompareAndSwapInt64(&v.status, on, off) {
		return ErrServAlreadyStopped
	}
	v.CloseAll()
	return nil
}

func (v *wsProvider) Broadcast(t uint, m json.Marshaler) {
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

func (v *wsProvider) CloseAll() {
	v.cm.Lock()
	defer v.cm.Unlock()

	for _, c := range v.clients {
		c.Close()
	}
}

func (v *wsProvider) Event(call Handler, eid ...uint) {
	v.em.Lock()
	defer v.em.Unlock()

	for _, i := range eid {
		v.events[i] = call
	}
}

func (v *wsProvider) addConn(c *conn) {
	v.cm.Lock()
	v.clients[c.CID()] = c
	v.cm.Unlock()
}

func (v *wsProvider) delConn(c *conn) {
	v.cm.Lock()
	delete(v.clients, c.CID())
	v.cm.Unlock()
}

func (v *wsProvider) CountConn() int {
	v.cm.Lock()
	cc := len(v.clients)
	v.cm.Unlock()
	return cc
}

func (v *wsProvider) getEventHandler(id uint) (Handler, error) {
	v.em.RLock()
	defer v.em.RUnlock()
	fn, ok := v.events[id]
	if !ok {
		return nil, ErrUnknownEventID
	}
	return fn, nil
}

func (v *wsProvider) Handling(ctx Ctx) {
	c := newConn(ctx.Context())

	c.OnError(func(cid string, msg string, err error) {
		if err == nil {
			return
		}
		v.log.WithFields(logger.Fields{
			"cid": cid,
			"err": err.Error(),
		}).Errorf(msg)
	})

	c.GetEventCall(v.getEventHandler)

	v.addConn(c)
	defer v.delConn(c)
	if err := c.Upgrade(ctx.Response(), ctx.Request(), *wsu); err != nil {
		ctx.SetBody(nethttp.StatusBadRequest).Error(err)
		return
	}
}
