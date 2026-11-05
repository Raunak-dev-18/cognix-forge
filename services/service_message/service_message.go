// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package service_message

import (
	"context"
	"fmt"

	service_message_model "forgejo.org/models/service_message"
	service_message_module "forgejo.org/modules/service_message"
)

var SMTypeModal = "modal"

func NewServiceMessage(opts *service_message_module.ServiceMessageOptions) (*service_message_model.ServiceMessage, error) {
	if opts.Title == "" {
		return nil, service_message_module.ErrInputWasEmpty
	}
	if opts.Text == "" {
		return nil, service_message_module.ErrInputWasEmpty
	}
	if !service_message_module.SMType(opts.Type).Valid() {
		return nil, service_message_module.ErrInvalidServiceMessageType
	}
	sm := &service_message_model.ServiceMessage{
		Title: opts.Title,
		Text:  opts.Text,
		Type:  service_message_module.SMType(opts.Type),
	}
	return sm, nil
}

func CreateOrUpdateServiceMessage(ctx context.Context, sm *service_message_model.ServiceMessage) error {
	return service_message_model.CreateOrUpdateServiceMessage(ctx, sm)
}

func GetServiceMessage(ctx context.Context, smType string) (*service_message_model.ServiceMessage, error) {
	smt := service_message_module.SMType(smType)
	if !smt.Valid() {
		return nil, fmt.Errorf("Invalid Service Message type %s", smType)
	}
	return service_message_model.GetServiceMessageByType(ctx, smt)
}

func DeleteServiceMessageByType(ctx context.Context, smType *service_message_module.SMType) error {
	return service_message_model.DeleteServiceMessageByType(ctx, smType)
}
