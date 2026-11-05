// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package a

import (
	"errors"

	"forgejo.org/routers/api/v1/permissions"
)

func directPermissionsCallFine(ctx permissions.Context) {
	ctx.Error(500, "title", nil)
}

// Directly call Context functions, then "do work", triggering a linting error:

func directPermissionsCallError(ctx permissions.Context) {
	ctx.Error(200, "tmpl", nil) // want "Invocation of (.*) / Error, and control flow continues afterwards."
	work()
}

func directPermissionsCallInternalServerError(ctx permissions.Context) {
	ctx.InternalServerError(errors.New("oops")) // want "Invocation of (.*) / InternalServerError, and control flow continues afterwards."
	work()
}

func directPermissionsCallNotFound(ctx permissions.Context) {
	ctx.NotFound("oops") // want "Invocation of (.*) / NotFound, and control flow continues afterwards."
	work()
}
