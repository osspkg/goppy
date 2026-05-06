package jsonrpc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/plugins/web/jsonrpc"
)

func TestUnit_Client_Call(t *testing.T) {
	t.SkipNow()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	ctx = context.WithValue(ctx, "xaid", "XYZ")

	srv := &http.Server{
		Addr: "127.0.0.1:12345",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := httputil.DumpRequest(r, true)
			fmt.Println(string(b))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":"111","result":[1,2,3]}]`))
		}),
	}
	go func() { srv.ListenAndServe() }()
	defer func() { srv.Close() }()
	defer cancel()

	time.Sleep(time.Second * 2)

	cli := jsonrpc.New(
		"http://127.0.0.1:12345/rpc",
		jsonrpc.SetContextHeader("X-Auth-Id", "xaid"),
		jsonrpc.SetHeader("User-Agent", "goppy"),
		jsonrpc.SetGenID(func() string { return "111" }),
	)

	in := jsonrpc.ModelAdapter[[]int]{Data: []int{1, 2, 3}}
	out := jsonrpc.ModelAdapter[[]int]{}

	casecheck.NoError(t, cli.Call(ctx, "app.user", in, &out))

	fmt.Println(out.Data)
}
