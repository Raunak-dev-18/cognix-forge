// Copyright 2017 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package private

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"forgejo.org/modules/httplib"
	"forgejo.org/modules/json"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"
)

// Response is used for internal request response (for user message and error message)
type Response struct {
	Err     string `json:"err,omitempty"`      // server-side error log message, it won't be exposed to end users
	UserMsg string `json:"user_msg,omitempty"` // meaningful error message for end users, it will be shown in git client's output.
}

func getClientIP() string {
	sshConnEnv := strings.TrimSpace(os.Getenv("SSH_CONNECTION"))
	if len(sshConnEnv) == 0 {
		return "127.0.0.1"
	}
	return strings.Fields(sshConnEnv)[0]
}

func newInternalRequest(ctx context.Context, url, method string, body ...any) *httplib.Request {
	url = fmt.Sprintf("http://forgejo-internal/%s", strings.TrimPrefix(url, "/"))
	req := httplib.NewRequest(url, method).
		SetContext(ctx).
		Header("X-Real-IP", getClientIP())

	req.SetTransport(&http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", setting.InternalListenerPath)
		},
	})

	if len(body) == 1 {
		req.Header("Content-Type", "application/json")
		jsonBytes, _ := json.Marshal(body[0])
		req.Body(jsonBytes)
	} else if len(body) > 1 {
		log.Fatal("Too many arguments for newInternalRequest")
	}

	req.SetTimeout(10*time.Second, 60*time.Second)
	return req
}
