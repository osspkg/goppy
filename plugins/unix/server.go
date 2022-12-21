package unix

import (
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/deweppro/go-errors"

	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-logger"
)

type (
	Config struct {
		Path string `yaml:"unix"`
	}
)

func (v *Config) Default() {
	v.Path = "./app.socket"
}

func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &Config{},
		Inject: func(conf *Config, log logger.Logger) (*srv, Server) {
			s := newServer(conf, log)
			return s, s
		},
	}
}

type (
	srv struct {
		config   *Config
		sock     net.Listener
		log      logger.Logger
		commands map[string]Handler
		mux      sync.RWMutex
	}

	//Handler unix socket command handler
	Handler func([]byte) ([]byte, error)

	Server interface {
		Command(name string, h Handler)
	}
)

func newServer(conf *Config, log logger.Logger) *srv {
	return &srv{
		config:   conf,
		log:      log,
		commands: make(map[string]Handler),
	}
}

func (v *srv) Up() (err error) {
	if err = os.Remove(v.config.Path); err != nil && !os.IsNotExist(err) {
		err = errors.WrapMessage(err, "remove unix socket [unix:%s]", v.config.Path)
		return
	}
	if v.sock, err = net.Listen("unix", v.config.Path); err != nil {
		err = errors.WrapMessage(err, "init unix socket [unix:%s]", v.config.Path)
		return
	}

	go v.accept()
	return
}

func (v *srv) Down() error {
	if v.sock != nil {
		return v.sock.Close()
	}
	return nil
}

func (v *srv) Command(name string, h Handler) {
	v.mux.Lock()
	v.commands[name] = h
	v.mux.Unlock()
}

func (v *srv) logError(err error, msg string) {
	if err == nil {
		return
	}

	v.log.WithFields(logger.Fields{
		"err": err.Error(),
	}).Errorf(msg)
}

func (v *srv) accept() {
	for {
		fd, err := v.sock.Accept()
		if err != nil {
			v.logError(err, "accept unix socket")
			return
		}
		if err = fd.SetDeadline(time.Now().Add(time.Hour)); err != nil {
			v.logError(err, "unix socket set deadline")
			return
		}
		go v.pump(fd)
	}
}

func (v *srv) pump(rw io.ReadWriteCloser) {
	defer func() {
		if err := rw.Close(); err != nil {
			v.logError(err, "close unix socket request")
		}
	}()

	b, err := readBytes(rw)
	if err != nil {
		v.logError(err, "read unix socket request")
		v.logError(writeError(rw, err), "write unix socket error")
		return
	}

	cmd, data := parse(b)

	v.mux.RLock()
	h, ok := v.commands[cmd]
	v.mux.RUnlock()
	if !ok {
		v.logError(writeError(rw, ErrInvalidCommand), "write unix socket error")
		return
	}

	out, err := h(data)
	if err != nil {
		v.logError(writeError(rw, err), "write unix socket error")
		return
	}
	v.logError(writeBytes(rw, out), "write unix socket response")
}
