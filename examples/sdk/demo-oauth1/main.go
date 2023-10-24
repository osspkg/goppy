/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"net/http"
	"time"

	"go.osspkg.com/goppy/sdk/app"
	"go.osspkg.com/goppy/sdk/auth/oauth"
	"go.osspkg.com/goppy/sdk/log"
	"go.osspkg.com/goppy/sdk/webutil"
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
