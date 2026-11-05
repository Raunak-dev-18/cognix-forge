// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin_setting

import (
	"strconv"
	"strings"

	"forgejo.org/modules/json"
	"forgejo.org/modules/setting"
)

func MarshalBool(value string) (string, error) {
	if b, _ := strconv.ParseBool(value); b {
		return "true", nil
	}
	return "false", nil
}

func MarshalOpenWithApps(value string) (string, error) {
	lines := strings.Split(value, "\n")
	var openWithEditorApps setting.OpenWithEditorAppsType
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		displayName, openURL, ok := strings.Cut(line, "=")
		displayName, openURL = strings.TrimSpace(displayName), strings.TrimSpace(openURL)
		if !ok || displayName == "" || openURL == "" {
			continue
		}
		openWithEditorApps = append(openWithEditorApps, setting.OpenWithEditorApp{
			DisplayName: strings.TrimSpace(displayName),
			OpenURL:     strings.TrimSpace(openURL),
		})
	}
	b, err := json.Marshal(openWithEditorApps)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
