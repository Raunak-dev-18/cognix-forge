// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add index to action_run.commit_sha",
		Upgrade:     addActionRunCommitShaIndex,
	})
}

func addActionRunCommitShaIndex(x *xorm.Engine) error {
	type ActionRun struct {
		RepoID           int64  `xorm:"index unique(repo_index) index(concurrency) index(sha)"`
		Index            int64  `xorm:"index unique(repo_index)"`
		ConcurrencyGroup string `xorm:"'concurrency_group' index(concurrency)"`
		CommitSHA        string `xorm:"index(sha)"` // indexed for API ?commit_sha=... filtering
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(ActionRun))
	return err
}
