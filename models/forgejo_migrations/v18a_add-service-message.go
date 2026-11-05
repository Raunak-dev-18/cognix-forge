// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	service_message "forgejo.org/modules/service_message"
	"forgejo.org/modules/timeutil"

	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add service_message table",
		Upgrade:     addServiceMessageTable,
	})
}

func addServiceMessageTable(x *xorm.Engine) error {
	type ServiceMessage struct {
		ID          int64                  `xorm:"pk autoincr"`
		Title       string                 `xorm:"NOT NULL"`
		Text        string                 `xorm:"LONGTEXT"`
		Type        service_message.SMType `xorm:"INDEX UNIQUE NOT NULL"`
		CreatedUnix timeutil.TimeStamp     `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp     `xorm:"updated"`
	}

	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, &ServiceMessage{})
	return err
}
