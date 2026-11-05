// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package auth

import (
	"net/http"
	"net/url"
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/templates"
	"forgejo.org/modules/test"
	"forgejo.org/services/auth/source/oauth2"
	"forgejo.org/services/contexttest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLogin(t *testing.T) {
	ctx, resp := contexttest.MockContext(t, "/user/login")
	SignIn(ctx)
	assert.Equal(t, http.StatusOK, resp.Code)

	ctx, resp = contexttest.MockContext(t, "/user/login")
	ctx.IsSigned = true
	SignIn(ctx)
	assert.Equal(t, http.StatusSeeOther, resp.Code)
	assert.Equal(t, "/", test.RedirectURL(resp))

	ctx, resp = contexttest.MockContext(t, "/user/login?redirect_to=/other")
	ctx.IsSigned = true
	SignIn(ctx)
	assert.Equal(t, "/other", test.RedirectURL(resp))

	ctx, resp = contexttest.MockContext(t, "/user/login")
	ctx.Req.AddCookie(&http.Cookie{Name: "redirect_to", Value: "/other-cookie"})
	ctx.IsSigned = true
	SignIn(ctx)
	assert.Equal(t, "/other-cookie", test.RedirectURL(resp))

	ctx, resp = contexttest.MockContext(t, "/user/login?redirect_to="+url.QueryEscape("https://example.com"))
	ctx.IsSigned = true
	SignIn(ctx)
	assert.Equal(t, "/", test.RedirectURL(resp))
}

// NB: Full signup test is in tests/integration/signup_test.go
// this is to test disabled signup
func TestSignUpDefault(t *testing.T) {
	ctx, resp := contexttest.MockContext(t, "/user/sign_up",
		contexttest.MockContextOption{Render: templates.HTMLRenderer()})
	SignUp(ctx)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), ctx.Locale.Tr("username"))
}

func TestSignUpDisabled(t *testing.T) {
	ctx, resp := contexttest.MockContext(t, "/user/sign_up",
		contexttest.MockContextOption{Render: templates.HTMLRenderer()})
	defer test.MockVariableValue(&setting.Service.DisableRegistration, true)()
	SignUp(ctx)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), ctx.Locale.Tr("auth.disable_register_prompt"))
}

func TestSignUpPostDisabled(t *testing.T) {
	ctx, resp := contexttest.MockContext(t, "/user/sign_up")
	defer test.MockVariableValue(&setting.Service.DisableRegistration, true)()
	SignUpPost(ctx)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestSignUpWithUsernamePrefix(t *testing.T) {
	ctx, resp := contexttest.MockContext(t, "/user/sign_up",
		contexttest.MockContextOption{Render: templates.HTMLRenderer()})
	defer test.MockVariableValue(&setting.Service.UsernamePrefix, "fbsd-")()
	SignUp(ctx)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "fbsd-")
}

func TestSignUpWithoutUsernamePrefix(t *testing.T) {
	ctx, resp := contexttest.MockContext(t, "/user/sign_up",
		contexttest.MockContextOption{Render: templates.HTMLRenderer()})
	defer test.MockVariableValue(&setting.Service.UsernamePrefix, "")()
	SignUp(ctx)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.NotContains(t, resp.Body.String(), "ui labeled input")
}

func TestSignInOAuth2AutoRedirect(t *testing.T) {
	// zero providers, internal sign-in disabled: no redirect
	require.NoError(t, unittest.PrepareTestDatabase())
	ctx, resp := contexttest.MockContext(t, "/user/login")
	SignIn(ctx)
	assert.Equal(t, http.StatusOK, resp.Code)

	// one provider, internal sign-in enabled: no redirect
	require.NoError(t, auth_model.CreateSource(db.DefaultContext, &auth_model.Source{
		Type:     auth_model.OAuth2,
		Name:     "test-provider",
		IsActive: true,
		Cfg: &oauth2.Source{
			Provider:     "github",
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
	}))
	t.Cleanup(func() { oauth2.RemoveProviderFromGothic("test-provider") })

	restoreInternal := test.MockVariableValue(&setting.Service.EnableInternalSignIn, true)
	ctx, resp = contexttest.MockContext(t, "/user/login")
	SignIn(ctx)
	assert.Equal(t, http.StatusOK, resp.Code)

	// one provider, internal sign-in disabled: redirect
	restoreInternal()
	ctx, resp = contexttest.MockContext(t, "/user/login")
	SignIn(ctx)
	assert.Equal(t, http.StatusSeeOther, resp.Code)
	assert.Equal(t, "/user/oauth2/test-provider", test.RedirectURL(resp))

	// two providers, internal sign-in disabled: no redirect
	require.NoError(t, auth_model.CreateSource(db.DefaultContext, &auth_model.Source{
		Type:     auth_model.OAuth2,
		Name:     "test-provider-2",
		IsActive: true,
		Cfg: &oauth2.Source{
			Provider:     "gitlab",
			ClientID:     "test-client-id-2",
			ClientSecret: "test-client-secret-2",
		},
	}))
	t.Cleanup(func() { oauth2.RemoveProviderFromGothic("test-provider-2") })

	ctx, resp = contexttest.MockContext(t, "/user/login")
	SignIn(ctx)
	assert.Equal(t, http.StatusOK, resp.Code)
}
