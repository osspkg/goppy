package broker

import (
	"fmt"
	"reflect"

	"go.osspkg.com/errors"
	"go.osspkg.com/goppy/v3/dic"
	"go.osspkg.com/logx"
	"go.osspkg.com/xc"
)

type UniversalBroker[T any] struct {
	objects []T
	index   int

	onStartCallback func(xc.Context, T) error
	onStopCallback  func(T) error
}

func WithUniversalBroker[T any](
	onStartCallback func(xc.Context, T) error,
	onStopCallback func(T) error,
) *UniversalBroker[T] {
	return &UniversalBroker[T]{
		objects:         make([]T, 0, 10),
		index:           0,
		onStartCallback: onStartCallback,
		onStopCallback:  onStopCallback,
	}
}

func (u *UniversalBroker[T]) Name() string {
	return fmt.Sprintf("UniversalBroker: %s", getTypeName[T]())
}

func (u *UniversalBroker[T]) Priority() int {
	return 0
}

func (u *UniversalBroker[T]) Apply(arg any) {
	if arg == nil {
		return
	}
	if v, ok := arg.(T); ok {
		u.objects = append(u.objects, v)
	}
}

func (u *UniversalBroker[T]) OnStart(ctx xc.Context) error {
	logx.Info("Universal Broker", "do", "start", "type", getTypeName[T](), "count", len(u.objects))

	if len(u.objects) == 0 {
		return nil
	}

	for i := 0; i < len(u.objects); i++ {
		if err := u.onStartCallback(ctx, u.objects[i]); err != nil {
			return err
		}
		u.index = i
	}

	return nil
}

func (u *UniversalBroker[T]) OnStop() error {
	logx.Info("Universal Broker", "do", "stop", "type", getTypeName[T](), "count", len(u.objects))

	if len(u.objects) == 0 {
		return nil
	}

	var errResult error
	for ; u.index >= 0; u.index-- {
		if err := u.onStopCallback(u.objects[u.index]); err != nil {
			errResult = errors.Wrap(
				errResult,
				errors.Wrapf(err, "down [%T] service error", u.objects[u.index]),
			)
		}
	}

	return errResult
}

func getTypeName[T any]() string {
	ref := reflect.ValueOf(new(T)).Elem()
	return dic.ResolveAddress(ref.Type(), ref)
}
