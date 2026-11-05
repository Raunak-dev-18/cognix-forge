// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import (
	"time"

	"forgejo.org/modules/optional"
)

// TeamInvite represents an invitation of a user to a team
type TeamInvite struct {
	// the identifier of the invitation
	ID int64 `json:"id"`
	// a token that the invited user can use to accept the invitation
	Token string `json:"token,omitempty"`
	// the identifier of the organization the invitation is issued for
	OrgID int64 `json:"org_id"`
	// the identifier of the team the invitation is issued for
	TeamID int64 `json:"team_id"`
	// the user who issued the invitation
	Inviter *User `json:"inviter"`
	// the invited user, if the invite is intended for an already-registered user
	Invited *User `json:"invited"`
	// the email of the invited person, if they don't have an account yet
	Email string `json:"email,omitempty"`
	// when this invitation was created
	CreatedAt time.Time `json:"created_at"`
	// when this invitation will expire (if at all)
	ExpiresAt optional.Option[time.Time] `json:"expires_at"`
	// whether this invitation has already expired
	IsExpired bool `json:"is_expired"`
}

// AcceptTeamInviteOptions options for accepting an invitation
// swagger:model
type AcceptTeamInviteOptions struct {
	// Whether to hide the membership to the organization containing this team.
	//
	// required: false
	HideMembership bool `json:"hide_membership"`
}
