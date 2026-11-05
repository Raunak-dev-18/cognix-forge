// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package private

import (
	"context"
)

// Email structure holds a data for sending general emails
type Email struct {
	Subject string
	Message string
	To      []string
}

// SendEmail calls the internal SendEmail function
// It accepts a list of usernames.
// If DB contains these users it will send the email to them.
// If to list == nil, it's supposed to send emails to every user present in DB
func SendEmail(ctx context.Context, subject, message string, to []string) (*ResponseText, ResponseExtra) {
	req := newInternalRequest(ctx, "/api/internal/mail/send", "POST", Email{
		Subject: subject,
		Message: message,
		To:      to,
	})

	return requestJSONResp(req, &ResponseText{})
}
