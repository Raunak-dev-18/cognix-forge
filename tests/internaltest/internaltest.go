// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package internaltest

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"forgejo.org/modules/optional"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/modules/web"
	"forgejo.org/routers"

	"github.com/stretchr/testify/require"
)

func NewInternalTestServer(t testing.TB, handler optional.Option[http.HandlerFunc]) {
	var socketPath string
	reset := func() {}
	if setting.InternalListenerPath == "random" {
		tmpDir := t.TempDir()
		socketPath = path.Join(tmpDir, "internal.sock")
		reset = test.MockVariableValue(&setting.InternalListenerPath, socketPath)
	} else {
		socketPath = setting.InternalListenerPath
		socketDir := filepath.Dir(socketPath)
		err := os.MkdirAll(socketDir, 0o700)
		require.NoError(t, err)
	}

	err := os.Remove(socketPath)
	if !os.IsNotExist(err) {
		require.NoError(t, err)
	}

	var routes *web.Route
	if handler.Has() {
		routes = web.NewRoute()
		_, handler := handler.Get()
		routes.R.Mount("/", handler)
	} else {
		routes = routers.InternalRoutes()
	}

	addr, err := net.ResolveUnixAddr("unix", socketPath)
	require.NoError(t, err)

	listener, err := net.ListenUnix("unix", addr)
	require.NoError(t, err)

	listener.SetUnlinkOnClose(true)
	httpServer := http.Server{
		Handler: routes,
	}

	go func() {
		err := httpServer.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			require.NoError(t, err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	t.Cleanup(func() {
		reset()
		err := httpServer.Shutdown(context.Background())
		require.NoError(t, err)
	})
}
