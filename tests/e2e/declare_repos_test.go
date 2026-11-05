// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package e2e

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/models/db"
	issues_model "forgejo.org/models/issues"
	repo_model "forgejo.org/models/repo"
	unit_model "forgejo.org/models/unit"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/git"
	"forgejo.org/modules/gitrepo"
	"forgejo.org/modules/indexer/stats"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/timeutil"
	actions_service "forgejo.org/services/actions"
	issue_service "forgejo.org/services/issue"
	pull_service "forgejo.org/services/pull"
	release_service "forgejo.org/services/release"
	files_service "forgejo.org/services/repository/files"
	"forgejo.org/services/wiki"
	"forgejo.org/tests/forgery"

	"code.forgejo.org/forgejo/runner/v13/act/jobparser"
	"code.forgejo.org/xorm/xorm/convert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// first entry represents filename
// the following entries define the full file content over time
type FileChanges struct {
	Filename  string
	CommitMsg string
	Versions  []string
}

// performs additional repo setup as needed
type SetupRepo func(*user_model.User, *repo_model.Repository)

// put your Git repo declarations in here
// feel free to amend the helper function below or use the raw variant directly
func DeclareGitRepos(t *testing.T) {
	now := timeutil.TimeStampNow()
	postIssue := func(repo *repo_model.Repository, user *user_model.User, age int64, title, content string) {
		issue := &issues_model.Issue{
			RepoID:      repo.ID,
			PosterID:    user.ID,
			Title:       title,
			Content:     content,
			CreatedUnix: now.Add(-age),
		}
		require.NoError(t, issue_service.NewIssue(db.DefaultContext, repo, issue, nil, nil, nil))
	}
	postPullRequest := func(repo *repo_model.Repository, branchName string, user *user_model.User, age int64, title, content string) {
		issue := &issues_model.Issue{
			RepoID:      repo.ID,
			PosterID:    user.ID,
			Poster:      user,
			Title:       title,
			Content:     content,
			CreatedUnix: now.Add(-age),
			IsPull:      true,
		}
		pr := &issues_model.PullRequest{
			Issue:      issue,
			HeadRepoID: issue.RepoID,
			BaseRepoID: issue.RepoID,
			HeadBranch: branchName,
			BaseBranch: repo.DefaultBranch,
			Status:     issues_model.PullRequestStatusMergeable,
		}
		require.NoError(t, pull_service.NewPullRequest(db.DefaultContext, repo, issue, []int64{}, []string{}, pr, []int64{}))
	}

	newRepo(t, 2, "diff-test", nil, []FileChanges{{
		Filename: "testfile",
		Versions: []string{"hello", "hallo", "hola", "native", "ubuntu-latest", "- runs-on: ubuntu-latest", "- runs-on: debian-latest"},
	}}, nil)
	newRepo(t, 2, "language-stats-test", nil, []FileChanges{{
		Filename: "main.rs",
		Versions: []string{"fn main() {", "println!(\"Hello World!\");", "}"},
	}}, nil)
	newRepo(t, 2, "mentions-highlighted", nil, []FileChanges{
		{
			Filename:  "history1.md",
			Versions:  []string{""},
			CommitMsg: "A commit message which mentions @user2 in the title\nand has some additional text which mentions @user1",
		},
		{
			Filename:  "history2.md",
			Versions:  []string{""},
			CommitMsg: "Another commit which mentions @user1 in the title\nand @user2 in the text",
		},
	}, nil)
	newRepo(t, 2, "multiline-commit-messages", nil, []FileChanges{
		{
			Filename:  "file1.md",
			Versions:  []string{"file"},
			CommitMsg: "A commit message\nwhich spans multiple lines",
		},
		{
			Filename:  "file3.md",
			Versions:  []string{"another file"},
			CommitMsg: "a\nb",
		},
		{
			Filename:  "file4.md",
			Versions:  []string{"yet another file"},
			CommitMsg: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed\n\ndo eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.\n\nDolorem cupiditate deleniti illo quo vitae culpa totam blanditiis. Architecto molestias eveniet quibusdam voluptas saepe modi reprehenderit quos. Nobis qui ipsam id et delectus. Corrupti cupiditate occaecati eius. Voluptas voluptatibus culpa nostrum. Id temporibus minima quis voluptate. Et sit quos autem est natus saepe. Velit vitae eos sint magnam et magnam dolore. Aspernatur suscipit dolorem sint fugiat repudiandae provident dolorem voluptatem. Ullam ut unde aperiam. Aut occaecati sit placeat adipisci non. Animi tempore autem molestias numquam ut qui iste. Pariatur recusandae ipsam maxime nihil quia veniam. Doloremque voluptatibus voluptatum consequatur illum iure aperiam deleniti non. Quo quisquam eveniet nihil animi. Et unde in sint eligendi aut autem. Veniam voluptates debitis ullam doloremque. Debitis provident tempore ab fugiat aut distinctio omnis. Vel sapiente nulla id. Quia aliquam ab est. Consequatur voluptatem id blanditiis distinctio. Qui expedita quibusdam qui earum quis culpa. Iste laudantium fuga vero provident voluptatem laboriosam ullam et. Alias consequuntur earum dolor nemo molestiae non neque.",
		},
	}, func(user *user_model.User, repo *repo_model.Repository) {
		// status+tag on main branch
		commitMainSha1 := commitNewFile(t, user, repo, "Another multiline commit message\nOnly this time, we have a big shiny status icon 🎉", "file1.5.md", "also a file")
		commitMainSha2 := commitNewFile(t, user, repo, "This is a commit.\nThe commit knows where it is because it knows where it isn't.", "file2.md", "also a file")
		addCommitStatus(t, user, repo, repo.DefaultBranch, commitMainSha1)
		addCommitStatus(t, user, repo, repo.DefaultBranch, commitMainSha2)
		tagCommitWithRelease(t, user, repo, commitMainSha2, "v1.4.2")

		// status on PR
		commitPrSha1 := addCommitWithMessageToBranch(t, user, repo, "main", "test-branch", "Yet another multiline commit message\nnow with a PR and status!", "file2.md", "", "still a file")
		commitPrSha2 := addCommitWithMessageToBranch(t, user, repo, "test-branch", "test-branch", "One more message\nThis time, the details are longer than the summary. Imagine that!", "file2.md", commitPrSha1, "still a file")
		commitPrSha3 := addCommitWithMessageToBranch(t, user, repo, "test-branch", "test-branch", "Normal commit message", "file2.md", commitPrSha2, "yep, still a file")
		commitPrSha4 := addCommitWithMessageToBranch(t, user, repo, "test-branch", "test-branch", "Normal commit message with status", "file2.md", commitPrSha3, "yep, still a file")
		addCommitWithMessageToBranch(t, user, repo, "test-branch", "test-branch", "Normal multiline\ncommit message", "file2.md", commitPrSha4, "yep, still a file")
		postPullRequest(repo, "test-branch", user, 455, "pullreq", "PR with multiline commits")
		addCommitStatus(t, user, repo, "test-branch", commitPrSha1)
		addCommitStatus(t, user, repo, "test-branch", commitPrSha2)
		addCommitStatus(t, user, repo, "test-branch", commitPrSha4)
	})
	newRepo(t, 2, "file-uploads", nil, []FileChanges{{
		Filename: "UPLOAD_TEST.md",
		Versions: []string{"# File upload test\nUse this repo to test various file upload features in new branches."},
	}}, nil)
	newRepo(t, 2, "unicode-escaping", map[unit_model.Type]convert.Conversion{
		unit_model.TypeCode: nil,
		unit_model.TypeWiki: nil,
	}, []FileChanges{{
		Filename: "a-file",
		Versions: []string{"{a}{а}"},
	}}, func(user *user_model.User, repo *repo_model.Repository) {
		wiki.InitWiki(db.DefaultContext, repo)
		wiki.AddWikiPage(db.DefaultContext, user, repo, "Home", "{a}{а}", "{a}{а}")
		wiki.AddWikiPage(db.DefaultContext, user, repo, "_Sidebar", "{a}{а}", "{a}{а}")
		wiki.AddWikiPage(db.DefaultContext, user, repo, "_Footer", "{a}{а}", "{a}{а}")
	})
	newRepo(t, 2, "multiple-combo-boxes", nil, []FileChanges{{
		Filename: ".forgejo/issue_template/multi-combo-boxes.yaml",
		Versions: []string{`
name: "Multiple combo-boxes"
description: "To show something"
body:
- type: textarea
  id: textarea-one
  attributes:
    label: one
- type: textarea
  id: textarea-two
  attributes:
    label: two
`},
	}}, nil)
	newRepo(t, 2, "markup-attention", nil, []FileChanges{{
		Filename: "github-modern.md",
		Versions: []string{`
> [!note]
> This text is a note

> [!tip]
> This text is a tip

> [!important]
> This text is important

> [!warning]
> This text is a warning

> [!caution]
> This text is to make someone cautious
`},
	}}, nil)
	newRepo(t, 11, "dependency-test", map[unit_model.Type]convert.Conversion{
		unit_model.TypeIssues: &repo_model.IssuesConfig{
			EnableDependencies: true,
		},
		unit_model.TypeCode: nil,
	}, []FileChanges{}, func(user *user_model.User, repo *repo_model.Repository) {
		postIssue(repo, user, 500, "first issue here", "an issue created earlier")
		postIssue(repo, user, 400, "second issue here", "not the right issue, but in the right repo")
		postIssue(repo, user, 300, "third issue here", "depends on things")
		postIssue(repo, user, 200, "unrelated issue", "shrug emoji")
		postIssue(repo, user, 100, "newest issue", "very new")
	})
	newRepo(t, 11, "dependency-test-2", map[unit_model.Type]convert.Conversion{
		unit_model.TypeIssues: &repo_model.IssuesConfig{
			EnableDependencies: true,
		},
		unit_model.TypeCode: nil,
	}, []FileChanges{}, func(user *user_model.User, repo *repo_model.Repository) {
		postIssue(repo, user, 450, "right issue", "an issue containing word right")
		postIssue(repo, user, 150, "left issue", "an issue containing word left")
	})
	newRepo(t, 2, "long-diff-test", nil, []FileChanges{{
		Filename: "test-README.md",
		Versions: []string{
			readStringFile(t, "tests/e2e/declarative-repo/long-diff-test/0-README.md"),
		},
	}}, func(user *user_model.User, repo *repo_model.Repository) {
		commit1Sha := addCommitToBranch(t, user, repo, "main", "test-branch", "test-README.md", "",
			readStringFile(t, "tests/e2e/declarative-repo/long-diff-test/1-README.md"))
		commit2Sha := addCommitToBranch(t, user, repo, "test-branch", "test-branch", "test-README.md", commit1Sha,
			readStringFile(t, "tests/e2e/declarative-repo/long-diff-test/2-README.md"))
		addCommitToBranch(t, user, repo, "test-branch", "test-branch", "test-README.md", commit2Sha,
			readStringFile(t, "tests/e2e/declarative-repo/long-diff-test/3-README.md"))
	})
	newRepo(t, 2, "huge-diff-test", nil, []FileChanges{{
		Filename: "glossary.po",
		Versions: []string{
			func() string {
				var sb strings.Builder
				sb.Write([]byte("0"))
				for i := 1; i < 2000; i++ {
					sb.WriteString(strconv.Itoa(i))
					sb.WriteByte('\n')
				}
				return sb.String()
			}(),
		},
	}}, func(user *user_model.User, repo *repo_model.Repository) {
		addCommitToBranch(t, user, repo, "main", "main-2", "glossary.po", "",
			func() string {
				var sb strings.Builder
				sb.Write([]byte("0"))
				for i := 1; i < 2000; i++ {
					sb.WriteString(strconv.Itoa(i))
					if i%12 == 0 {
						sb.WriteString("Blub")
					}
					sb.WriteByte('\n')
				}
				return sb.String()
			}())
	})
	newRepo(t, 2, "funding_basic_complete", nil, []FileChanges{{
		Filename: ".forgejo/FUNDING.yml",
		Versions: []string{`
community_bridge: example
github:
  - example
  - example2
issuehunt: example
ko_fi: [example, example_2_electric_boogaloo]
liberapay: example
patreon: example
open_collective: example
buy_me_a_coffee: example
thanks_dev: u/gh/example
tidelift: npm/example
custom: ["https://example.com", 😀.com]
`},
	}}, nil)
	newRepo(t, 2, "funding_some_valid", nil, []FileChanges{{
		Filename: ".forgejo/FUNDING.yml",
		Versions: []string{`
ko_fi: 1337
custom: ["https://example.com"]
ko-fi: example
`},
	}}, nil)
	newRepo(t, 2, "funding_with_a_really_ridiculously_long_title_that_doesnt_really_happen_all_that_often_normally_but_could_really_mess_with_things_if_not_handled_properly", nil, []FileChanges{{
		Filename: ".forgejo/FUNDING.yml",
		Versions: []string{`
custom: example.com
`},
	}}, nil)
	newRepo(t, 6, ".profile", nil, []FileChanges{{
		Filename: ".forgejo/FUNDING.yml",
		Versions: []string{`
ko_fi: example
liberapay: example
custom: "http://localhost:3003/"
`},
	}}, nil)
	newRepo(t, 39, ".profile", nil, []FileChanges{{
		Filename: "FUNDING.yml",
		Versions: []string{`
ko_fi: example
liberapay: example
custom: "http://localhost:3003/"
`},
	}}, nil)
	// add your repo declarations here
}

func readStringFile(t *testing.T, fn string) string {
	c, err := os.ReadFile(filepath.Join(setting.AppWorkPath, fn))
	require.NoError(t, err)
	return string(c)
}

func newRepo(t *testing.T, userID int64, repoName string, enabledUnits map[unit_model.Type]convert.Conversion, fileChanges []FileChanges, setup SetupRepo) {
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: userID})

	somerepo := forgery.CreateRepository(t, user, &forgery.CreateRepositoryOptions{
		Name:  repoName,
		Files: forgery.FilesInit{},
	})
	if len(enabledUnits) == 0 {
		forgery.EnableRepoUnits(t, somerepo,
			unit_model.TypeCode,
			unit_model.TypeIssues,
		)
	}
	for unit, config := range enabledUnits {
		forgery.EnableRepoUnit(t, somerepo, unit, config)
	}

	var lastCommitID string
	for _, file := range fileChanges {
		for i, version := range file.Versions {
			operation := "update"
			if i == 0 {
				operation = "create"
			}

			// default to unique commit messages
			commitMsg := file.CommitMsg
			if commitMsg == "" {
				commitMsg = fmt.Sprintf("Patch: %s-%d", file.Filename, i+1)
			}

			resp, err := files_service.ChangeRepoFiles(git.DefaultContext, somerepo, user, &files_service.ChangeRepoFilesOptions{
				Files: []*files_service.ChangeRepoFile{{
					Operation:     operation,
					TreePath:      file.Filename,
					ContentReader: strings.NewReader(version),
				}},
				Message:   commitMsg,
				OldBranch: "main",
				NewBranch: "main",
				Author: &files_service.IdentityOptions{
					Name:  user.Name,
					Email: user.Email,
				},
				Committer: &files_service.IdentityOptions{
					Name:  user.Name,
					Email: user.Email,
				},
				Dates: &files_service.CommitDateOptions{
					Author:    time.Now(),
					Committer: time.Now(),
				},
				LastCommitID: lastCommitID,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, resp)

			lastCommitID = resp.Commit.SHA
		}
	}

	if setup != nil {
		setup(user, somerepo)
	}

	err := stats.UpdateRepoIndexer(somerepo)
	require.NoError(t, err)
}

func addCommitToBranch(t *testing.T, user *user_model.User, repo *repo_model.Repository, oldBranch, newBranch, filename, lastSha, content string) string {
	return addCommitWithMessageToBranch(t, user, repo, oldBranch, newBranch, "add commit to branch", filename, lastSha, content)
}

func addCommitWithMessageToBranch(t *testing.T, user *user_model.User, repo *repo_model.Repository, oldBranch, newBranch, commitMessage, filename, lastSha, content string) string {
	resp, err := files_service.ChangeRepoFiles(git.DefaultContext, repo, user, &files_service.ChangeRepoFilesOptions{
		Files: []*files_service.ChangeRepoFile{{
			Operation:     "update",
			TreePath:      filename,
			ContentReader: strings.NewReader(content),
		}},
		Message:   commitMessage,
		OldBranch: oldBranch,
		NewBranch: newBranch,
		Author: &files_service.IdentityOptions{
			Name:  user.Name,
			Email: user.Email,
		},
		Committer: &files_service.IdentityOptions{
			Name:  user.Name,
			Email: user.Email,
		},
		Dates: &files_service.CommitDateOptions{
			Author:    time.Now(),
			Committer: time.Now(),
		},
		LastCommitID: lastSha,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp)
	return resp.Commit.SHA
}

func commitNewFile(t *testing.T, user *user_model.User, repo *repo_model.Repository, commitMessage, filename, content string) string {
	resp, err := files_service.ChangeRepoFiles(git.DefaultContext, repo, user, &files_service.ChangeRepoFilesOptions{
		Files: []*files_service.ChangeRepoFile{{
			Operation:     "create",
			TreePath:      filename,
			ContentReader: strings.NewReader(content),
		}},
		Message: commitMessage,
		Author: &files_service.IdentityOptions{
			Name:  user.Name,
			Email: user.Email,
		},
		Committer: &files_service.IdentityOptions{
			Name:  user.Name,
			Email: user.Email,
		},
		Dates: &files_service.CommitDateOptions{
			Author:    time.Now(),
			Committer: time.Now(),
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp)
	return resp.Commit.SHA
}

func addCommitStatus(t *testing.T, user *user_model.User, repo *repo_model.Repository, branchName, commitSha string) {
	run := &actions_model.ActionRun{
		ID:                rand.Int63(),
		Title:             "test action",
		RepoID:            repo.ID,
		OwnerID:           user.ID,
		WorkflowID:        "",
		WorkflowDirectory: "",
		TriggerUserID:     user.ID,
		Ref:               "refs/heads/" + branchName,
		Event:             "push",
		EventPayload:      `{"head_commit": {"id": "` + commitSha + `"}}`,
		CommitSHA:         commitSha,
		Status:            actions_model.StatusSuccess,
	}
	err := actions_service.InsertRun(db.DefaultContext, run, []*jobparser.SingleWorkflow{})
	require.NoError(t, err)
	job := &actions_model.ActionRunJob{
		ID:                rand.Int63(),
		RunID:             run.ID,
		Attempt:           1,
		Status:            actions_model.StatusSuccess,
		RepoID:            run.RepoID,
		OwnerID:           run.OwnerID,
		CommitSHA:         commitSha,
		IsForkPullRequest: false,
	}
	actions_service.CreateCommitStatus(db.DefaultContext, job)
}

func tagCommitWithRelease(t *testing.T, user *user_model.User, repo *repo_model.Repository, target, tagName string) {
	gitRepo, err := gitrepo.OpenRepository(db.DefaultContext, repo)
	require.NoError(t, err)
	defer gitRepo.Close()
	rel := &repo_model.Release{
		RepoID:           repo.ID,
		PublisherID:      user.ID,
		Publisher:        user,
		TagName:          tagName,
		Target:           target,
		Title:            tagName,
		Note:             "",
		IsDraft:          false,
		IsPrerelease:     false,
		HideArchiveLinks: false,
		IsTag:            false,
		Repo:             repo,
	}
	err = release_service.CreateRelease(gitRepo, rel, "", nil)
	require.NoError(t, err)
}
