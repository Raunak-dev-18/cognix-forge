// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// templates/admin/dashboard.tmpl
// templates/base/head_script.tmpl
// web_src/js/webcomponents/relative-time.ts
// @watch end

import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';

test.use({user: 'user1'});

// The <relative-time> tooltip shows the absolute datetime, formatted with the language
// selected in Forgejo's settings (document.documentElement.lang), not the browser's language.
const browserLang = 'en-US';

for (const lang of ['en-US', 'es-ES', 'de-DE']) {
  test(`Relative time tooltip is formatted in ${lang}`, async ({browser}) => {
    const context = await browser.newContext({locale: browserLang});
    await context.addCookies([{name: 'lang', value: lang, domain: 'localhost', path: '/'}]);

    try {
      const page = await context.newPage();
      await page.goto('/user2/repo1');

      const relativeTime = page.locator('relative-time').first();
      await expect(relativeTime).toBeVisible();
      await expect(relativeTime).toHaveAttribute('data-tooltip-content', /.+?/);

      const expected = await relativeTime.evaluate(
        (el: HTMLElement, l: string) =>
          new Intl.DateTimeFormat(l, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            timeZoneName: 'short',
          }).format(new Date(el.getAttribute('datetime')!)),
        lang,
      );

      await expect(relativeTime).toHaveAttribute('data-tooltip-content', expected);

      // tippy is created with `animation: false`, so `data-state` stays "hidden"
      // only the `visibility` changes when shown
      await relativeTime.scrollIntoViewIfNeeded();
      await relativeTime.hover();
      await relativeTime.dispatchEvent('mouseenter');
      const tippyContent = page.locator('.tippy-box .tippy-content').filter({
        hasText: expected,
      });

      await expect(tippyContent).toBeVisible();
      await expect(tippyContent).toHaveText(expected);
    } finally {
      await context.close();
    }
  });
}

test('Relative time after htmx swap', async ({page}, workerInfo) => {
  test.skip(
    workerInfo.project.name !== 'firefox' && workerInfo.project.name !== 'Mobile Chrome',
    'This is a really slow test, so limit to a subset of client.',
  );
  await page.goto('/admin');

  const relativeTime = page.locator('.admin-dl-horizontal > dd:nth-child(2) > relative-time');
  // The admin dashboard uses <relative-time format="duration"> for server
  // uptime, which renders as a duration like "5 days, 3 hours" (no "ago").
  // Check that the component produced a formatted duration rather than the
  // raw datetime fallback.
  const durationPattern = /\d+ (year|month|week|day|hour|minute|second)/;
  await expect(relativeTime).toContainText(durationPattern);

  // The <relative-time> custom element renders its formatted text into an open
  // shadow root. Read that text directly: locator.textContent() does not
  // reflect the shadow-DOM content, and an htmx morph re-adds the raw-datetime
  // light-DOM fallback which would pollute toHaveText/toContainText.
  const textBefore = await relativeTime.evaluate((el) => el.shadowRoot?.textContent ?? '');

  const body = page.locator('body');
  await body.evaluate(
    (element) =>
      new Promise<void>((resolve) =>
        element.addEventListener('htmx:afterSwap', () => {
          resolve();
        }),
      ),
  );

  // The system-status panel refreshes itself via htmx every 5 seconds. A
  // previous regression (0e8d752d86) reset the rendered relative-time text on
  // swap; assert the formatted text survives the swap unchanged.
  await expect.poll(
    () => relativeTime.evaluate((el) => el.shadowRoot?.textContent ?? ''),
  ).toBe(textBefore);
});
