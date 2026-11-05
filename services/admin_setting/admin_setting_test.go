// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalOpenWithApps(t *testing.T) {
	testStr := `VSCodium = vscode://vscode.git/clone?url={url}`

	marshalled, err := MarshalOpenWithApps(testStr)
	require.NoError(t, err)

	assert.JSONEq(t, `[{"DisplayName":"VSCodium","OpenURL":"vscode://vscode.git/clone?url={url}"}]`, marshalled)
}
