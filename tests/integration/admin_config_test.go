// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/modules/test"
	app_context "forgejo.org/services/context"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestAdminConfig(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")
	req := NewRequest(t, "GET", "/admin/config")
	resp := session.MakeRequest(t, req, http.StatusOK)
	assert.True(t, test.IsNormalPageCompleted(resp.Body.String()))
}

func TestAdminConfigCacheTest(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")
	req := NewRequest(t, "POST", "/admin/config/test_cache")
	session.MakeRequest(t, req, http.StatusSeeOther)

	flashCookie := session.GetCookie(app_context.CookieNameFlash)
	assert.NotNil(t, flashCookie)
	assert.Contains(t, flashCookie.Value, "info%3DCache%2Btest%2Bsuccessful%252C%2Bgot%2Ba%2Bresponse")
}
