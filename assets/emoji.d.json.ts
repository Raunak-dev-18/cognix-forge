// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// Declaration file and `allowArbitraryExtensions` needed until https://github.com/microsoft/TypeScript/issues/32063 is resolved

/** A list of key-value pairs mapping emoji glyphs to one or more descriptive alias strings. */
declare const emoji: [string, string[]][];
export default emoji;
