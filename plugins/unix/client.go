package unix

import (
	"net"
	"sync"

	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-errors"
)

func WithClient() plugins.Plugin {
	return plugins.Plugin{
		Inject: func() (*cliProvider, Client) {
			s := newCliProvider()
			return s, s
		},
	}
}

type (
	cliProvider struct {
		list map[string]ClientConnect
		mux  sync.RWMutex
	}

	Client interface {
		Create(path string) (ClientConnect, error)
	}
)

func newCliProvider() *cliProvider {
	return &cliProvider{
		list: make(map[string]ClientConnect),
	}
}

func (v *cliProvider) Create(path string) (ClientConnect, error) {
	v.mux.Lock()
	defer v.mux.Unlock()
	if c, ok := v.list[path]; ok {
		return c, nil
	}
	c := newClient(path)
	v.list[path] = c
	return c, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	cli struct {
		path string
	}

	ClientConnect interface {
		Exec(name string, b []byte) ([]byte, error)
		ExecString(name string, b string) ([]byte, error)
	}
)

func newClient(path string) *cli {
	return &cli{
		path: path,
	}
}

func (v *cli) Exec(name string, b []byte) ([]byte, error) {
	conn, err := net.Dial("unix", v.path)
	if err != nil {
		return nil, errors.WrapMessage(err, "open connect [unix:%s]", v.path)
	}
	defer conn.Close() //nolint: errcheck
	if err = writeBytes(conn, append([]byte(name+cmddelimstring), b...)); err != nil {
		return nil, err
	}
	return readBytes(conn)
}

func (v *cli) ExecString(name string, b string) ([]byte, error) {
	return v.Exec(name, []byte(b))
}
