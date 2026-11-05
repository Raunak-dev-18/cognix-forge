// Copyright 2026 The Forgejo Authors
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/models/organization"
	access_model "forgejo.org/models/perm/access"
	"forgejo.org/models/unittest"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestOrgTeams(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// not logged in user
	req := NewRequest(t, "GET", "/org/org3/teams")
	MakeRequest(t, req, http.StatusSeeOther)

	// not org member
	session := loginUser(t, "user5")
	req = NewRequest(t, "GET", "/org/org3/teams")
	session.MakeRequest(t, req, http.StatusNotFound)

	// org member, not part of the Owners team
	session = loginUser(t, "user28")
	req = NewRequest(t, "GET", "/org/org3/teams")
	doc := NewHTMLParser(t, session.MakeRequest(t, req, http.StatusOK).Body)
	doc.AssertElement(t, "a[href^='/org/org3/teams/owners'].text.black", false)
	doc.AssertElement(t, "a[href^='/org/org3/teams/team12creators'].text.black", true)
	// despite not being able to go to the page for the Owners team, the user still sees it exists:
	doc.AssertElement(t, "strong:contains('Owners')", true)

	// org owner
	session = loginUser(t, "user2")
	req = NewRequest(t, "GET", "/org/org3/teams")
	doc = NewHTMLParser(t, session.MakeRequest(t, req, http.StatusOK).Body)
	doc.AssertElement(t, "a[href^='/org/org3/teams/owners'].text.black", true)
	doc.AssertElement(t, "a[href^='/org/org3/teams/team12creators'].text.black", true)

	// site admin
	session = loginUser(t, "user1")
	req = NewRequest(t, "GET", "/org/org3/teams")
	doc = NewHTMLParser(t, session.MakeRequest(t, req, http.StatusOK).Body)
	doc.AssertElement(t, "a[href^='/org/org3/teams/owners'].text.black", true)
	doc.AssertElement(t, "a[href^='/org/org3/teams/team12creators'].text.black", true)
}

func TestOrgTeamRemoveAllRepositories(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	team := unittest.AssertExistsAndLoadBean(t, &organization.Team{OrgID: 41, LowerName: "team1"})
	unittest.AssertCount(t, &organization.TeamRepo{TeamID: team.ID, RepoID: 61}, 1)
	unittest.AssertCount(t, &access_model.Access{RepoID: 61, UserID: 39}, 1)

	session := loginUser(t, "user40")
	req := NewRequest(t, "POST", "/org/org41/teams/team1/action/repo/removeall")
	resp := session.MakeRequest(t, req, http.StatusSeeOther)
	assert.Equal(t, "/org/org41/teams/team1/repositories", resp.Header().Get("Location"))

	unittest.AssertCount(t, &organization.TeamRepo{TeamID: team.ID}, 0)
	unittest.AssertCount(t, &access_model.Access{RepoID: 61, UserID: 39}, 0) // access should be removed
}
