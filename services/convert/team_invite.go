// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package convert

import (
	"context"

	org_model "forgejo.org/models/organization"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/optional"
	api "forgejo.org/modules/structs"
)

func ToTeamInvite(ctx context.Context, invite *org_model.TeamInvite, doer *user_model.User) *api.TeamInvite {
	serialized := api.TeamInvite{
		ID:        invite.ID,
		OrgID:     invite.OrgID,
		TeamID:    invite.TeamID,
		Inviter:   ToUser(ctx, invite.InviterUser, doer),
		Invited:   ToUser(ctx, invite.InvitedUser, doer),
		CreatedAt: invite.CreatedUnix.AsTime(),
		IsExpired: invite.IsExpired(),
	}
	// only show the email address if the invite isn't associated to a user
	if invite.InvitedUser == nil {
		serialized.Email = invite.Email
	}
	// only show the invite token if requested by the user it is intended for, or an admin
	if (invite.InvitedUser != nil && doer.ID == invite.InvitedUser.ID) || doer.IsAdmin {
		serialized.Token = invite.Token
	}
	expires, expiry := invite.ExpiryUnix.Get()
	if expires {
		serialized.ExpiresAt = optional.Some(expiry.AsTime())
	}
	return &serialized
}
