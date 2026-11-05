// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package service_message

import (
	"context"
	"errors"

	"forgejo.org/models/db"
	"forgejo.org/modules/log"
	service_message_types "forgejo.org/modules/service_message"
	"forgejo.org/modules/timeutil"
	"forgejo.org/modules/util"
)

func init() {
	db.RegisterModel(new(ServiceMessage))
}

type ServiceMessage struct {
	ID          int64                        `xorm:"pk autoincr"`
	Title       string                       `xorm:"NOT NULL"`
	Text        string                       `xorm:"LONGTEXT"`
	Type        service_message_types.SMType `xorm:"INDEX UNIQUE NOT NULL"`
	CreatedUnix timeutil.TimeStamp           `xorm:"created"`
	UpdatedUnix timeutil.TimeStamp           `xorm:"updated"`
}

// CreateOrUpdateServiceMessage creates record of a ServiceMessage, expects a valid ServiceMessage
func CreateOrUpdateServiceMessage(ctx context.Context, sm *ServiceMessage) error {
	// Create if not exists, trim title to fit DB size requirements
	sm.Title, _ = util.SplitStringAtByteN(sm.Title, 255)
	existing, err := GetServiceMessageByType(ctx, sm.Type)
	if err != nil {
		if errors.Is(err, service_message_types.ErrServiceMessageNotExist) {
			err = db.Insert(ctx, sm)
			log.Debug("Created service message of type %q", sm.Type.Name())
		}
		return err
	}
	// Update if exists
	e := db.GetEngine(ctx)
	_, err = e.ID(existing.ID).Cols(
		"title",
		"text",
	).Update(sm)
	log.Debug("Existing Service Message %s was updated.", sm.Type.Name())
	return err
}

func GetServiceMessageByType(ctx context.Context, smType service_message_types.SMType) (*ServiceMessage, error) {
	sm := &ServiceMessage{}
	has, err := db.GetEngine(ctx).Where("type = ?", smType.Name()).Get(sm)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, service_message_types.ErrServiceMessageNotExist
	}
	log.Debug("Got service message %q", smType)
	return sm, nil
}

// Delete the service message by type, expecting there will ever only be a single message per type
func DeleteServiceMessageByType(ctx context.Context, smType *service_message_types.SMType) error {
	_, err := db.GetEngine(ctx).Where("type = ?", smType.Name()).Delete(ServiceMessage{Type: *smType})
	if err != nil {
		return err
	}
	log.Debug("Deleted service message %q", smType.Name())
	return nil
}
