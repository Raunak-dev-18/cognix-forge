// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package convert

import (
	"testing"
	"time"

	"forgejo.org/models/db"
	org_model "forgejo.org/models/organization"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertInvitedUserWithExpiry(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	user5 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 5})

	teamInvite := org_model.TeamInvite{
		ID:          42,
		Token:       "aBcDeFg",
		InviterID:   user2.ID,
		InvitedID:   optional.Some(user5.ID),
		TeamID:      1,
		OrgID:       3,
		Email:       user5.Email,
		CreatedUnix: timeutil.TimeStampNow(),
		ExpiryUnix:  optional.Some(timeutil.TimeStampNow().AddDuration(1000 * time.Second)),
	}
	require.NoError(t, teamInvite.LoadUsers(db.DefaultContext))

	apiInvite := ToTeamInvite(db.DefaultContext, &teamInvite, user2)
	assert.Equal(t, int64(42), apiInvite.ID)
	assert.Empty(t, apiInvite.Token) // because it is not viewed by the invited user
	assert.Equal(t, user2.Name, apiInvite.Inviter.UserName)
	assert.Equal(t, user5.Name, apiInvite.Invited.UserName)
	assert.Equal(t, int64(1), apiInvite.TeamID)
	assert.Equal(t, int64(3), apiInvite.OrgID)
	assert.Empty(t, apiInvite.Email)
	assert.NotNil(t, apiInvite.ExpiresAt)
	assert.False(t, apiInvite.IsExpired)

	apiInvite = ToTeamInvite(db.DefaultContext, &teamInvite, user5)
	assert.Equal(t, "aBcDeFg", apiInvite.Token) // because it is viewed by the invited user
}

func TestConvertEmailWithoutExpiry(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	teamInvite := org_model.TeamInvite{
		ID:          42,
		InviterID:   user2.ID,
		InvitedID:   optional.None[int64](),
		TeamID:      1,
		OrgID:       3,
		Email:       "foo@example.com",
		CreatedUnix: timeutil.TimeStampNow(),
		ExpiryUnix:  optional.None[timeutil.TimeStamp](),
	}
	require.NoError(t, teamInvite.LoadUsers(db.DefaultContext))

	apiInvite := ToTeamInvite(db.DefaultContext, &teamInvite, user2)
	assert.Equal(t, int64(42), apiInvite.ID)
	assert.Empty(t, apiInvite.Token)
	assert.Equal(t, user2.Name, apiInvite.Inviter.UserName)
	assert.Nil(t, apiInvite.Invited)
	assert.Equal(t, int64(1), apiInvite.TeamID)
	assert.Equal(t, int64(3), apiInvite.OrgID)
	assert.Equal(t, "foo@example.com", apiInvite.Email)
	assert.Nil(t, apiInvite.ExpiresAt)
	assert.False(t, apiInvite.IsExpired)
}
