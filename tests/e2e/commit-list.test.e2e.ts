// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// templates/repo/commits_list.tmpl
// templates/repo/commits_list_small.tmpl
// templates/repo/latest_commit.tmpl
// templates/repo/pulls/commits_list.tmpl
// web_src/css/repo.css
// web_src/css/repo/commit-list.css
// @watch end

import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';
import {screenshot} from './shared/screenshots.ts';

const runs = [
  {title: 'JS off', useJs: false},
  {title: 'JS on', useJs: true},
] as const;
for (const run of runs) {
  test.describe(run.title, () => {
    test.use({javaScriptEnabled: run.useJs});

    test.describe('Commits list', () => {
      test('Repo latest commit', async ({page, isMobile}) => {
        const response = await page.goto('/user2/multiline-commit-messages');
        expect(response?.status()).toBe(200);

        const summary = page.locator('.commit-summary', {hasText: 'This is a commit.'});
        const toggle = summary.getByLabel('Toggle full commit message');
        const body = page.locator('.commit-body', {hasText: 'The commit knows where it is because it knows where it isn\'t.'});

        await expect(summary).toBeVisible();
        await expect(toggle).toBeVisible();
        await expect(body).toBeHidden();

        await toggle.click({force: isMobile}); // open! span intercepts pointer events on mobile somehow, so force
        await expect(toggle).toBeVisible();
        await expect(summary).toBeVisible();
        await expect(body).toBeVisible();

        await toggle.click({force: isMobile}); // close!
        await expect(summary).toBeVisible();
        await expect(toggle).toBeVisible();
        await expect(body).toBeHidden();
      });

      test('Repo commits', async ({page, isMobile}) => {
        test.skip(!run.useJs); // TODO: this might work without JS after updating the layout
        const response = await page.goto('/user2/multiline-commit-messages/commits/branch/main');
        expect(response?.status()).toBe(200);

        const message = page.locator('.message', {hasText: 'Another multiline commit message'});
        const toggle = message.getByLabel('Toggle full commit message');
        const body = message.locator('.commit-body', {hasText: 'Only this time, we have a big shiny status icon 🎉'});
        const status = message.locator('a:has(> .octicon-check)');
        const otherBody = page.locator('.commit-body', {hasText: 'which spans multiple lines'});

        await expect(message).toBeVisible();
        await expect(toggle).toBeVisible();
        await expect(body).toBeHidden();
        await expect(otherBody).toBeHidden();

        await toggle.click({force: isMobile}); // open!
        await expect(toggle).toBeVisible();
        await expect(message).toBeVisible();
        await expect(body).toBeVisible();
        await expect(otherBody).toBeHidden();

        await toggle.click({force: isMobile}); // close!
        await expect(message).toBeVisible();
        await expect(toggle).toBeVisible();
        await expect(body).toBeHidden();
        await expect(otherBody).toBeHidden();

        // clicking the status (which lives inside the <summary> now) navigates, rather than opening the message body
        await status.click();
        await expect(page).toHaveURL('/user2/multiline-commit-messages/actions/runs/1/jobs/0');
      });
    });

    test.describe('PR commits', () => {
      test.use({user: 'user2'});

      test('Any layout', async ({page}) => {
        const response = await page.goto('/user2/repo1/pulls/3/commits');
        expect(response?.status()).toBe(200);

        const commitGroup = page.locator('.commit-group:first-of-type');

        // Date group visibility test
        await expect(commitGroup).toBeVisible();
        await expect(commitGroup.locator('h4')).toBeVisible();

        const commit = commitGroup.locator('.commit:first-child');

        await expect(commit).toHaveCSS('display', 'grid');
      });

      test('Mobile responsive layout checks', async ({page, isMobile}) => {
        test.skip(!isMobile);

        const response = await page.goto('/user2/repo1/pulls/3/commits');
        expect(response?.status()).toBe(200);

        // Mobile-specific visibility test
        const commit = page.locator('.commit-group:first-of-type .commit:first-child');
        await expect(commit.locator('.commit-buttons')).toBeHidden();
        await expect(commit.locator('.button-sequence button[data-clipboard-text]')).toBeVisible();

        // Mobile-specific grid positioning
        // toHaveCSS returns absolute values in px with decimals. This matcher only
        // checks if the string has two \S+px separated by one \s+
        await expect(commit).toHaveCSS('grid-template-columns', /^\S+px\s+\S+px$/);
        await expect(commit.locator('.author')).toHaveCSS('grid-column-start', '1');
        await expect(commit.locator('.date')).toHaveCSS('grid-column-start', '2');
        await expect(commit.locator('.message')).toHaveCSS('grid-column-end', 'span 2');

        // Horizontal scrolling to check for overflow
        await expect(page.locator('.commits').first()).not.toHaveCSS('overflow-x', 'scroll');

        await screenshot(page);
      });

      test('Dropdown check in mobile viewport', async ({page, isMobile}) => {
        test.skip(!isMobile);

        const response = await page.goto('/user2/repo1/pulls/3/commits');
        expect(response?.status()).toBe(200);

        const commit = page.locator('.commit-group:first-of-type .commit:first-child');

        // Click dropdown btn
        const dropdown = commit.locator('.dialog-dropdown');
        await expect(dropdown).toBeVisible();
        await dropdown.locator('.opener').click();

        // List menu items of dropdown
        const menuItem = commit.locator('.dialog-dropdown ul li a'); // repo_path; always visible
        await expect(menuItem).toHaveAttribute('href', '/user2/repo1/src/commit/5f22f7d0d95d614d25a5b68592adb345a4b5c7fd');
        await menuItem.click();
        await page.waitForURL(/.*\/user2\/repo1\/src\/commit\/5f22f7d0d95d614d25a5b68592adb345a4b5c7f/);
        await screenshot(page);
      });

      test('Desktop responsive layout checks', async ({page, isMobile}) => {
        test.skip(isMobile);

        const response = await page.goto('/user2/repo1/pulls/3/commits');
        expect(response?.status()).toBe(200);

        // Desktop-specific visibility test
        const commit = page.locator('.commit-group:first-of-type .commit:first-child');
        await expect(commit.locator('.commit-buttons')).toBeVisible();
        await expect(commit.locator('.button-sequence button[data-clipboard-text]')).toBeHidden();
        await expect(commit.locator('.dialog-dropdown')).toBeHidden();

        // Desktop layout is has specific grid-template-columns
        // toHaveCSS returns absolute values in px with decimals. This matcher only
        // checks if the string has five \S+px separated by four \s+
        await expect(commit).toHaveCSS('grid-template-columns', /^\S+px\s+\S+px\s+\S+px\s+\S+px\s+\S+px$/);

        await screenshot(page);
      });

      test('Multiline commit message in comments', async ({browser, isMobile}) => {
        // ensure multiline commits work in the PR view
        const context = await browser.newContext({javaScriptEnabled: run.useJs});
        const page = await context.newPage();
        const response = await page.goto('/user2/multiline-commit-messages/pulls/1');
        expect(response?.status()).toBe(200);

        const summary = page.locator('.message-wrapper', {hasText: 'Yet another multiline commit message'});
        const toggle = summary.getByLabel('Toggle full commit message');
        const body = page.locator('.commit-body', {hasText: 'now with a PR and status!'});
        const status = page.locator('.commits-list a:has(> .octicon-check)').first();

        await expect(summary).toBeVisible();
        await expect(toggle).toBeVisible();
        await expect(body).toBeHidden();

        await toggle.click({force: isMobile}); // open!
        await expect(toggle).toBeVisible();
        await expect(summary).toBeVisible();
        await expect(body).toBeVisible();

        await toggle.click({force: isMobile}); // close!
        await expect(summary).toBeVisible();
        await expect(toggle).toBeVisible();
        await expect(body).toBeHidden();

        // clicking the status navigates, rather than opening the message body
        await status.click();
        await expect(page).toHaveURL('/user2/multiline-commit-messages/actions/runs/3/jobs/0');
      });

      test('Multiline commit message in list', async ({page, isMobile}) => {
        test.skip(!run.useJs); // TODO: this might work without JS after updating the layout

        // ensure multiline commits work in the PR view
        const response = await page.goto('/user2/multiline-commit-messages/pulls/1/commits');
        expect(response?.status()).toBe(200);

        const summary = page.locator('.message-wrapper', {hasText: 'Yet another multiline commit message'});
        const toggle = summary.getByLabel('Toggle full commit message');
        const body = page.locator('.commit-body', {hasText: 'now with a PR and status!'});
        const status = page.locator('a:has(> .octicon-check)').first();

        await expect(summary).toBeVisible();
        await expect(toggle).toBeVisible();
        await expect(body).toBeHidden();

        await toggle.click({force: isMobile}); // open!
        await expect(toggle).toBeVisible();
        await expect(summary).toBeVisible();
        await expect(body).toBeVisible();

        await toggle.click({force: isMobile}); // close!
        await expect(summary).toBeVisible();
        await expect(toggle).toBeVisible();
        await expect(body).toBeHidden();

        // clicking the status navigates, rather than opening the message body
        await status.click();
        await expect(page).toHaveURL('/user2/multiline-commit-messages/actions/runs/5/jobs/0');
      });
    });
  });
}
