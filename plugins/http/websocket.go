package http

//go:generate easyjson

import (
	"context"
	"encoding/json"
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

type Message interface {
	Decode(in interface{}) error
	Encode(out interface{}) error
}

var (
	poolEvent = sync.Pool{New: func() interface{} { return &event{} }}
)

//easyjson:json
type event struct {
	ID      uint            `json:"id"`
	Err     *string         `json:"e"`
	Data    json.RawMessage `json:"d"`
	Updated bool            `json:"-"`
}

func (v *event) Decode(in interface{}) error {
	return json.Unmarshal(v.Data, in)
}

func (v *event) Encode(out interface{}) error {
	b, err := json.Marshal(out)
	if err != nil {
		return err
	}
	v.Body(b)
	return nil
}

func (v *event) Reset() *event {
	v.ID, v.Err, v.Data, v.Updated = 0, nil, v.Data[:0], false
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	pongWait   = 60 * time.Second
	pingPeriod = pongWait / 3
)

type (
	Connection interface {
		UID() string
		GetHeader(key string) string
	}

	conn struct {
		uid     string
		ctx     context.Context
		cncl    context.CancelFunc
		onClose []func(uid string)
		onError func(uid string, msg string, err error)
		sendC   chan []byte
		conn    *websocket.Conn
		headers nethttp.Header
		handler call
		wg      sync.WaitGroup
		status  int64
	}
)

func newConn(ctx context.Context) *conn {
	c, cncl := context.WithCancel(ctx)
	return &conn{
		uid:     uuid.NewString(),
		onClose: make([]func(string), 0),
		sendC:   make(chan []byte, 128),
		ctx:     c,
		cncl:    cncl,
		status:  on,
	}
}

func (v *conn) UID() string {
	return v.uid
}

func (v *conn) GetHeader(key string) string {
	if v.headers == nil {
		return ""
	}
	return v.headers.Get(key)
}

func (v *conn) Handler(h call) {
	v.handler = h
}

func (v *conn) OnClose(cb func(string)) {
	v.onClose = append(v.onClose, cb)
}

func (v *conn) OnError(cb func(uid string, msg string, err error)) {
	v.onError = cb
}

func (v *conn) Close() {
	if !atomic.CompareAndSwapInt64(&v.status, on, off) {
		return
	}
	for _, fn := range v.onClose {
		fn(v.UID())
	}
	if v.conn != nil {
		v.cncl()
		v.wg.Wait()
		if err := v.conn.Close(); err != nil {
			v.onError(v.UID(), "[ws] close conn", err)
		}
	}
}

func (v *conn) JSON(in interface{}) {
	b, err := json.Marshal(in)
	if err != nil {
		v.onError(v.UID(), "[ws] marshal message", err)
		return
	}

	v.Write(b)
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
				v.onError(v.UID(), "[ws] send close", err)
			}
			return
		case m := <-v.sendC:
			if err := v.conn.WriteMessage(websocket.TextMessage, m); err != nil {
				v.onError(v.UID(), "[ws] send message", err)
				return
			}
		case <-ticker.C:
			if err := v.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				v.onError(v.UID(), "[ws] send ping", err)
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
			if !websocket.IsCloseError(err, 1000, 1001) {
				v.onError(v.UID(), "[ws] read message", err)
			}
			return
		}
		v.handler(message, v)
	}
}

func (v *conn) Upgrade(w nethttp.ResponseWriter, r *nethttp.Request, up websocket.Upgrader) (err error) {
	v.conn, err = up.Upgrade(w, r, nil)
	if err != nil {
		v.onError(v.UID(), "[ws] upgrade", err)
		v.Close()
		return
	}

	v.headers = r.Header
	if err = v.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		v.onError(v.UID(), "[ws] send pong", err)
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
	ErrInvalidResponse    = errors.New("invalid response")
	ErrUnknownMessageType = errors.New("unknown message type")

	wsu = websocket.Upgrader{
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
				option(&wsu)
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
		calls   map[uint]Handler

		log logger.Logger
		mux sync.RWMutex
	}

	Handler func(m Message, c Connection) error
	call    func([]byte, *conn)

	WebSocket interface {
		Handling(ctx Ctx)
		Event(t uint, call Handler)
		Broadcast(t uint, b []byte)
		CloseAll()
	}
)

func newWsProvider(log logger.Logger) *wsProvider {
	return &wsProvider{
		status:  off,
		clients: make(map[string]*conn),
		calls:   make(map[uint]Handler),
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

func (v *wsProvider) Broadcast(t uint, b []byte) {
	v.mux.RLock()
	defer v.mux.RUnlock()

	m, ok := poolEvent.Get().(*event)
	if !ok {
		return
	}
	defer poolEvent.Put(m.Reset())

	m.ID = t
	m.Body(b)

	bb, err := json.Marshal(m)
	if err != nil {
		return
	}

	for _, c := range v.clients {
		c.Write(bb)
	}
}

func (v *wsProvider) CloseAll() {
	v.mux.Lock()
	defer v.mux.Unlock()

	for _, c := range v.clients {
		c.Close()
	}
}

func (v *wsProvider) Event(t uint, call Handler) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.calls[t] = call
}

func (v *wsProvider) Handling(ctx Ctx) {
	c := newConn(ctx.Context())

	c.OnError(func(uid string, msg string, err error) {
		if err == nil {
			return
		}
		v.log.WithFields(logger.Fields{
			"uid": uid,
			"err": err.Error(),
		}).Errorf(msg)
	})

	c.Handler(v.processor)

	v.mux.Lock()
	v.clients[c.UID()] = c
	v.mux.Unlock()

	if err := c.Upgrade(ctx.Response(), ctx.Request(), wsu); err != nil {
		ctx.SetBody(nethttp.StatusBadRequest).Error(err)
		return
	}

	v.mux.Lock()
	delete(v.clients, c.UID())
	v.mux.Unlock()
}

func (v *wsProvider) processor(b []byte, c *conn) {
	m, ok := poolEvent.Get().(*event)
	if !ok {
		return
	}
	defer poolEvent.Put(m.Reset())

	if err := json.Unmarshal(b, m); err != nil {
		m.Error(err)
		c.JSON(m)
		return
	}

	v.mux.RLock()
	fn, ok := v.calls[m.ID]
	if !ok {
		v.mux.RUnlock()
		m.Error(ErrUnknownMessageType)
		c.JSON(m)
		return
	}
	v.mux.RUnlock()

	if err := fn(m, c); err != nil {
		m.Error(err)
	}

	if !m.Updated {
		m.Error(ErrInvalidResponse)
	}

	c.JSON(m)
}
