// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package issue_test

import (
	"context"
	"testing"

	issues_model "forgejo.org/models/issues"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	issue_service "forgejo.org/services/issue"
	notify_service "forgejo.org/services/notify"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reactionNotifier struct {
	notify_service.NullNotifier
	new     []*issues_model.Reaction
	deleted []*issues_model.Reaction
}

func (o *reactionNotifier) NewReaction(ctx context.Context, reaction *issues_model.Reaction) {
	o.new = append(o.new, reaction)
}

func (o *reactionNotifier) DeleteReaction(ctx context.Context, reaction *issues_model.Reaction) {
	o.deleted = append(o.deleted, reaction)
}

func TestServicesReaction(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	reactionEqual := func(t *testing.T, a, b *issues_model.Reaction) {
		t.Helper()
		assert.Equal(t, a.Type, b.Type)
		assert.Equal(t, a.UserID, b.UserID)
		assert.Equal(t, a.IssueID, b.IssueID)
		assert.Equal(t, a.CommentID, b.CommentID)
	}

	t.Run("CommentReaction", func(t *testing.T) {
		unittest.LoadFixtures()
		notifier := &reactionNotifier{}
		notify_service.RegisterNotifier(notifier)
		defer notify_service.UnregisterNotifier(notifier)

		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		comment := unittest.AssertExistsAndLoadBean(t, &issues_model.Comment{ID: 2})
		require.NoError(t, comment.LoadIssue(t.Context()))
		issue := comment.Issue

		content := "+1"
		expectedReaction := &issues_model.Reaction{
			Type:      content,
			UserID:    user.ID,
			IssueID:   issue.ID,
			CommentID: comment.ID,
		}

		reaction, err := issue_service.CreateCommentReaction(t.Context(), user, issue, comment, content)
		require.NoError(t, err)
		reactionEqual(t, expectedReaction, reaction)
		require.Len(t, notifier.new, 1)
		reactionEqual(t, expectedReaction, notifier.new[0])
		require.Empty(t, notifier.deleted, 0)

		require.NoError(t, issue_service.DeleteCommentReaction(t.Context(), user, comment, content))
		require.Len(t, notifier.new, 1)
		require.Len(t, notifier.deleted, 1)
		reactionEqual(t, expectedReaction, notifier.deleted[0])
	})

	t.Run("IssueReaction", func(t *testing.T) {
		unittest.LoadFixtures()
		notifier := &reactionNotifier{}
		notify_service.RegisterNotifier(notifier)
		defer notify_service.UnregisterNotifier(notifier)

		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		issue := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 2})

		content := "+1"
		expectedReaction := &issues_model.Reaction{
			Type:    content,
			UserID:  user.ID,
			IssueID: issue.ID,
		}

		reaction, err := issue_service.CreateIssueReaction(t.Context(), user, issue, content)
		require.NoError(t, err)
		reactionEqual(t, expectedReaction, reaction)
		require.Len(t, notifier.new, 1)
		reactionEqual(t, expectedReaction, notifier.new[0])
		require.Empty(t, notifier.deleted, 0)

		require.NoError(t, issue_service.DeleteIssueReaction(t.Context(), user, issue, content))
		require.Len(t, notifier.new, 1)
		require.Len(t, notifier.deleted, 1)
		reactionEqual(t, expectedReaction, notifier.deleted[0])
	})
}
