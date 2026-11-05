// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"fmt"
	"net/http"
	"testing"

	service_message_model "forgejo.org/models/service_message"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	service_message_module "forgejo.org/modules/service_message"
	"forgejo.org/modules/test"
	"forgejo.org/tests"
	"forgejo.org/tests/forgery"

	"github.com/stretchr/testify/assert"
)

func TestServiceMessage(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	user := forgery.CreateUser(t, &forgery.CreateUserOptions{IsAdmin: true})
	session := loginUser(t, user.Name)

	smOpts := service_message_module.ServiceMessageOptions{
		Type:  "modal",
		Text:  "TestText",
		Title: "Service Message",
	}

	t.Run("Preview", func(t *testing.T) {
		// Admin clicks on preview
		req := NewRequestWithJSON(t, "POST", "/admin/service_message/preview?sm_type=modal&status=show", &smOpts)
		resp := session.MakeRequest(t, req, http.StatusOK)
		assert.Contains(t, resp.Body.String(), smOpts.Text)
	})

	req := NewRequestWithJSON(t, "POST", "/admin/service_message?sm_type=modal", &smOpts)
	session.MakeRequest(t, req, http.StatusSeeOther)
	sm := unittest.AssertExistsAndLoadBean(t, &service_message_model.ServiceMessage{Type: "modal"})
	assert.Equal(t, service_message_module.SMType(smOpts.Type), sm.Type)
	assert.Equal(t, smOpts.Text, sm.Text)
	assert.Equal(t, smOpts.Title, sm.Title)

	t.Run("TestMustShow", func(t *testing.T) {
		// ServiceMessage has been setup and is shown for the first time
		sm := unittest.AssertExistsAndLoadBean(t, &service_message_model.ServiceMessage{Type: "modal"})
		assert.True(t, user.MustShowServiceMessage(sm.Type, sm.UpdatedUnix))

		// User clicks accept, is updated and Service Message goes away
		req = NewRequest(t, "POST", fmt.Sprintf("/user/confirm?sm_type=%s&sm_created=%v&redirect_to=localhost", sm.Type, sm.CreatedUnix))
		session.MakeRequest(t, req, http.StatusSeeOther)
		updatedUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: user.ID})
		assert.False(t, updatedUser.MustShowServiceMessage(sm.Type, sm.UpdatedUnix))

		// Admin updates Service Message, so it needs to be shown again
		test.SleepTillNextSecond()
		smOpts = service_message_module.ServiceMessageOptions{
			Type:  "modal",
			Text:  "TestText2",
			Title: "Title",
		}
		req2 := NewRequestWithJSON(t, "POST", "/admin/service_message?sm_type=modal", &smOpts)
		session.MakeRequest(t, req2, http.StatusSeeOther)
		sm2 := unittest.AssertExistsAndLoadBean(t, &service_message_model.ServiceMessage{Type: "modal"})
		assert.True(t, updatedUser.MustShowServiceMessage(sm2.Type, sm2.UpdatedUnix))
	})

	t.Run("Delete", func(t *testing.T) {
		req = NewRequest(t, "POST", fmt.Sprintf("/admin/service_message/delete?sm_type=%s", sm.Type))
		session.MakeRequest(t, req, http.StatusOK)
		unittest.AssertNotExistsBean(t, &service_message_model.ServiceMessage{Type: "modal"})
	})
}
