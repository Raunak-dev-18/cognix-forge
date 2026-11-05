// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

type Context interface {
	Error(status int, title string, obj any)
	InternalServerError(err error)
	NotFound(objs ...any)
}
