/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/osspkg/goppy/sdk/app"
	"github.com/osspkg/goppy/sdk/auth/oauth"
	"github.com/osspkg/goppy/sdk/log"
	"github.com/osspkg/goppy/sdk/webutil"
)

var (
	provConf = &oauth.Config{
		Provider: []oauth.ConfigItem{
			{
				Code:         "google",
				ClientID:     "****************.apps.googleusercontent.com",
				ClientSecret: "****************",
				RedirectURL:  "https://example.com/oauth/callback/google",
			},
		},
	}

	servConf = webutil.ConfigHttp{Addr: ":8080"}
)

func main() {
	ctx := app.NewContext()
	authServ := oauth.New(provConf)

	route := webutil.NewRouter()
	route.Route("/oauth/request/google", authServ.Request(oauth.CodeGoogle), http.MethodGet)
	route.Route("/oauth/callback/google", authServ.CallBack(oauth.CodeGoogle, oauthCallBackHandler), http.MethodGet)

	serv := webutil.NewServerHttp(servConf, route, log.Default())
	serv.Up(ctx) //nolint: errcheck
	<-time.After(60 * time.Minute)
	ctx.Close()
	serv.Down() //nolint: errcheck
}

const out = `
email: %s
name:  %s
ico:   %s
`

func oauthCallBackHandler(w http.ResponseWriter, _ *http.Request, u oauth.User) {
	w.WriteHeader(200)
	fmt.Fprintf(w, out, u.GetEmail(), u.GetName(), u.GetIcon())
}
