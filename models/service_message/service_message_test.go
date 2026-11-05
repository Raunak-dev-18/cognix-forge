// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package service_message

import (
	"testing"

	"forgejo.org/models/unittest"
	"forgejo.org/modules/service_message"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUpdateGetDeleteServiceMessage(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	serviceMessage := &ServiceMessage{
		Text:        "This is a text.",
		Type:        service_message.SMType("modal"),
		CreatedUnix: timeutil.TimeStampNow(),
	}

	// Create
	err := CreateOrUpdateServiceMessage(t.Context(), serviceMessage)
	require.NoError(t, err)
	retrieved := unittest.AssertExistsAndLoadBean(t, &ServiceMessage{Type: serviceMessage.Type})
	assert.Equal(t, serviceMessage.Type, retrieved.Type)
	assert.Equal(t, serviceMessage.Text, retrieved.Text)
	assert.NotEqual(t, int64(0), serviceMessage.ID)

	sm2 := ServiceMessage{
		Type:        service_message.SMType("modal"),
		Text:        "This is another text.",
		CreatedUnix: timeutil.TimeStampNow(),
	}

	// Update
	err = CreateOrUpdateServiceMessage(t.Context(), &sm2)
	require.NoError(t, err)
	retrieved = unittest.AssertExistsAndLoadBean(t, &ServiceMessage{Type: sm2.Type})
	assert.Equal(t, sm2.Type, retrieved.Type)
	assert.Equal(t, sm2.Text, retrieved.Text)

	// Get
	get, err := GetServiceMessageByType(t.Context(), sm2.Type)
	require.NoError(t, err)
	assert.Equal(t, sm2.Type, get.Type)

	// Delete
	err = DeleteServiceMessageByType(t.Context(), &sm2.Type)
	require.NoError(t, err)
	unittest.AssertNotExistsBean(t, &ServiceMessage{Type: sm2.Type})
}
