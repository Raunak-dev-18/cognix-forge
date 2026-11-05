// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package service_message

import (
	"testing"

	service_message_module "forgejo.org/modules/service_message"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceMessage(t *testing.T) {
	opts := service_message_module.ServiceMessageOptions{
		Type:  "modal",
		Text:  "Some Text.",
		Title: "Title",
	}
	sm, err := NewServiceMessage(&opts)
	require.NoError(t, err)
	assert.Equal(t, sm.Type.Name(), opts.Type)
	assert.Equal(t, timeutil.TimeStamp(0), sm.CreatedUnix)

	invalidOpts := service_message_module.ServiceMessageOptions{
		Type: "",
		Text: "",
	}
	_, err = NewServiceMessage(&invalidOpts)
	assert.Error(t, err)
}
