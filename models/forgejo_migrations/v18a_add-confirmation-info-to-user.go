// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"forgejo.org/modules/timeutil"

	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add confirmation info to user table",
		Upgrade:     addConfirmInfoToUser,
	})
}

func addConfirmInfoToUser(x *xorm.Engine) error {
	type SMType string
	type ConfirmTimestamps map[SMType]timeutil.TimeStamp
	type User struct {
		ID       int64 `xorm:"pk autoincr"`
		Confirms ConfirmTimestamps
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, &User{})
	return err
}
