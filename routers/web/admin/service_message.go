// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin

import (
	"errors"
	"net/http"
	"strings"

	service_message_model "forgejo.org/models/service_message"
	"forgejo.org/modules/base"
	"forgejo.org/modules/log"
	service_message_module "forgejo.org/modules/service_message"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/templates"
	"forgejo.org/modules/web"
	"forgejo.org/services/context"
	"forgejo.org/services/forms"
	service_message_service "forgejo.org/services/service_message"
)

const (
	tplServiceMessage base.TplName = "admin/service_message"
)

func GetServiceMessage(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.service_message")
	ctx.Data["PageIsAdminServiceMessage"] = true
	sm, err := service_message_service.GetServiceMessage(ctx, service_message_service.SMTypeModal)
	if err != nil {
		if errors.Is(err, service_message_module.ErrServiceMessageNotExist) {
			sm = &service_message_model.ServiceMessage{}
		} else {
			ctx.ServerError("GetServiceMessage", err)
			return
		}
	}
	ctx.Data["ModalServiceMessageTitle"] = sm.Title
	ctx.Data["ModalServiceMessageText"] = sm.Text
	ctx.Data["RenderedContent"] = templates.RenderMarkdownToHtml(ctx, sm.Text)
	ctx.HTML(http.StatusOK, tplServiceMessage)
}

func CreateOrUpdateServiceMessage(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.ServiceMessageForm)
	smType := strings.TrimSpace(ctx.FormString("sm_type"))
	smOpts := service_message_module.ServiceMessageOptions{
		Title: form.Title,
		Text:  form.Text,
		Type:  smType,
	}
	serviceMessage, err := service_message_service.NewServiceMessage(&smOpts)
	if err != nil {
		if strings.Contains(err.Error(), "empty") {
			ctx.RenderWithErr(ctx.Tr("admin.service_message.empty_input"), tplServiceMessage, form)
		} else if strings.Contains(err.Error(), "invalid") {
			ctx.RenderWithErr(ctx.Tr("admin.service_message.invalid_input"), tplServiceMessage, form)
		}
		return
	}
	err = service_message_service.CreateOrUpdateServiceMessage(ctx, serviceMessage)
	if err != nil {
		ctx.ServerError("Could not create service message", err)
		return
	}
	log.Debug("Done creating Service Message.")
	ctx.Redirect(setting.AppSubURL + "/admin/service_message")
}

func DeleteServiceMessage(ctx *context.Context) {
	smType := strings.TrimSpace(ctx.FormString("sm_type"))
	sm, err := service_message_service.GetServiceMessage(ctx, smType)
	if err != nil {
		if errors.Is(err, service_message_module.ErrServiceMessageNotExist) {
			form := forms.ServiceMessageForm{}
			ctx.RenderWithErr(ctx.Tr("admin.service_message.is_already_deleted"), tplServiceMessage, form)
		} else {
			ctx.ServerError("DeleteServiceMessage", err)
		}
		return
	}
	err = service_message_service.DeleteServiceMessageByType(ctx, &sm.Type)
	if err != nil {
		ctx.ServerError("DeleteServiceMessage", err)
		return
	}
	log.Debug("Deleted Service Message %s", smType)
	ctx.JSONRedirect(setting.AppSubURL + "/admin/service_message")
}

func PreviewServiceMessage(ctx *context.Context) {
	if ctx.Written() {
		return
	}
	form := web.GetForm(ctx).(*forms.ServiceMessageForm)
	status := strings.TrimSpace(ctx.FormString("status"))
	smType := strings.TrimSpace(ctx.FormString("sm_type"))
	smOpts := service_message_module.ServiceMessageOptions{
		Title: form.Title,
		Text:  form.Text,
		Type:  smType,
	}
	sm, err := service_message_service.NewServiceMessage(&smOpts)
	if err != nil {
		if strings.Contains(err.Error(), "empty") {
			ctx.RenderWithErr(ctx.Tr("admin.service_message.empty_input"), tplServiceMessage, form)
		} else if strings.Contains(err.Error(), "invalid") {
			ctx.RenderWithErr(ctx.Tr("admin.service_message.invalid_input"), tplServiceMessage, form)
		}
		return
	}
	ctx.Data["ModalServiceMessageText"] = sm.Text
	ctx.Data["ModalServiceMessageTitle"] = sm.Title
	if status == "show" {
		ctx.Data["IsAdminPreview"] = true
		ctx.Data["RenderedContent"] = templates.RenderMarkdownToHtml(ctx, sm.Text)
	}
	if status == "closed" {
		delete(ctx.Data, "RenderedContent")
		delete(ctx.Data, "IsAdminPreview")
		ctx.Redirect(setting.AppSubURL + "/admin/service_message")
		return
	}
	ctx.HTML(http.StatusOK, tplServiceMessage)
}
