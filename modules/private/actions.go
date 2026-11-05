// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package private

import (
	"context"
)

type GenerateTokenRequest struct {
	Scope string
}

// GenerateActionsRunnerToken calls the internal GenerateActionsRunnerToken function
func GenerateActionsRunnerToken(ctx context.Context, scope string) (*ResponseText, ResponseExtra) {
	req := newInternalRequest(ctx, "/api/internal/actions/generate_actions_runner_token", "POST", GenerateTokenRequest{
		Scope: scope,
	})

	return requestJSONResp(req, &ResponseText{})
}
