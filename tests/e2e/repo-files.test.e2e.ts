// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// @watch start
// templates/repo/editor/**
// web_src/js/features/common-global.js
// routers/web/web.go
// services/repository/files/**
// web_src/js/vendor/jquery.are-you-sure.js
// @watch end

import {type Page, expect} from '@playwright/test';
import {test, dynamic_id} from './utils_e2e.ts';

test.use({user: 'user2'});

interface TestCase {
  description: string;
  files: string[];
}

async function prepareUpload(page: Page, testCase: TestCase) {
  const dropzone = page.getByRole('button', {name: 'Drop files or click here to upload.'});

  // create the virtual files
  const dataTransfer = await page.evaluateHandle((testCase: TestCase) => {
    const dt = new DataTransfer();
    for (const filename of testCase.files) {
      dt.items.add(new File([`File content of ${filename}`], filename, {type: 'text/plain'}));
    }
    return dt;
  }, testCase);
  // and drop them to the upload area
  await dropzone.dispatchEvent('drop', {dataTransfer});
}

async function doUpload(page: Page, testCase: TestCase) {
  await page.goto(`/user2/file-uploads/_upload/main/`);
  const testID = dynamic_id();
  await prepareUpload(page, testCase);

  await page.getByText('new branch').click();

  await page.getByRole('textbox', {name: 'Name the new branch for this'}).fill(testID);
  // ToDo: Potential race condition: We do not currently wait for the upload to complete.
  // See https://codeberg.org/forgejo/forgejo/pulls/6687#issuecomment-5068272 and
  // https://codeberg.org/forgejo/forgejo/issues/5893#issuecomment-5068266 for details.
  // Workaround is to wait (the uploads are just a few bytes and usually complete instantly)
  //
  // eslint-disable-next-line playwright/no-wait-for-timeout
  await page.waitForTimeout(100);

  await page.getByRole('button', {name: 'Propose file change'}).click();
}

test.describe('Drag and Drop upload', () => {
  const goodTestCases: TestCase[] = [
    {
      description: 'normal and special characters',
      files: [
        'dir1/file1.txt',
        'double/nested/file.txt',
        'special/äüöÄÜÖß.txt',
        'special/Ʉ₦ł₵ØĐɆ.txt',
      ],
    },
    {
      description: 'strange paths and spaces',
      files: [
        '..dots.txt',
        '.dots.preserved.txt',
        'special/S P  A   C   E    !.txt',
      ],
    },
  ];

  // actual good tests based on definition above
  for (const testCase of goodTestCases) {
    test(`good: ${testCase.description}`, async ({page}) => {
      await doUpload(page, testCase);

      // check that nested file structure is preserved
      for (const filename of testCase.files) {
        await expect(page.locator('#diff-file-boxes').getByRole('link', {name: filename})).toBeVisible();
      }
    });
  }

  const badTestCases: TestCase[] = [
    {
      description: 'broken path slash in front',
      files: [
        '/special/badfirstslash.txt',
      ],
    },
  ];

  // actual bad tests based on definition above
  for (const testCase of badTestCases) {
    test(`bad: ${testCase.description}`, async ({page}) => {
      await doUpload(page, testCase);
      await expect(page.getByText('Failed to upload files to')).toBeVisible();
    });
  }

  test.describe('Reload protection', () => {
    const testCase = goodTestCases[0];

    test('choose to stay', async ({page}) => {
      await page.goto(`/user2/file-uploads/_upload/main/`, {waitUntil: 'load'});

      let didShowDialog = false;
      page.on('dialog', async (dialog) => {
        expect(dialog.type()).toBe('beforeunload');
        await dialog.dismiss(); // "stay on page"
        didShowDialog = true;
      });

      let didReload = false;
      page.on('domcontentloaded', () => {
        // 'domcontentloaded' happens before 'load', so this fn only happens after a ReLoad!
        didReload = true;
      });

      const firstFilePreview = page.getByText('file1.txt');
      await expect(firstFilePreview).toBeHidden();
      await prepareUpload(page, testCase);
      await expect(page.getByRole('link', {name: 'Remove file'})).toHaveCount(testCase.files.length);
      await expect(firstFilePreview).toBeVisible();

      await page.locator('body').click(); // user interaction is required for beforeunload to fire
      await page.evaluate(() => window.location.reload()); // navigation won't actually occur
      await expect(() => expect(didShowDialog).toBe(true)).toPass();
      await expect(firstFilePreview).toBeVisible(); // files remain (we didn't close!)
      expect(didReload).toBe(false); // doing this last, to avoid race conditions
    });

    test('choose to leave', async ({page}) => {
      await page.goto(`/user2/file-uploads/_upload/main/`, {waitUntil: 'load'});

      let didShowDialog = false;
      page.on('dialog', async (dialog) => {
        expect(dialog.type()).toBe('beforeunload');
        await dialog.accept(); // "I'm sure. Reload!"
        didShowDialog = true;
      });

      const firstFilePreview = page.getByText('file1.txt');
      await expect(firstFilePreview).toBeHidden();
      await prepareUpload(page, testCase);
      await expect(page.getByRole('link', {name: 'Remove file'})).toHaveCount(testCase.files.length);
      await expect(firstFilePreview).toBeVisible();

      await page.locator('body').click(); // twiddling our thumbs...
      await page.reload({waitUntil: 'domcontentloaded'});
      await expect(() => expect(didShowDialog).toBe(true)).toPass();
      await expect(firstFilePreview).toBeHidden(); // files gone!
    });

    test('choose to stay, remove all files, then leave without bother', async ({page}) => {
      await page.goto(`/user2/file-uploads/_upload/main/`, {waitUntil: 'load'});

      let didShowDialog = false;
      page.on('dialog', async (dialog) => {
        expect(dialog.type()).toBe('beforeunload');
        await dialog.dismiss(); // "stay on page"
        didShowDialog = true;
      });

      let didReload = false;
      page.on('domcontentloaded', () => {
        didReload = true;
      });

      const firstFilePreview = page.getByText('file1.txt');
      await expect(firstFilePreview).toBeHidden();
      await prepareUpload(page, testCase);
      const removers = page.getByRole('link', {name: 'Remove file'});
      await expect(removers).toHaveCount(testCase.files.length);
      await expect(firstFilePreview).toBeVisible();

      await page.locator('body').click(); // humming a tune...
      await page.evaluate(() => window.location.reload()); // navigation won't actually occur
      await expect(() => expect(didShowDialog).toBe(true)).toPass();
      didShowDialog = false;
      await expect(firstFilePreview).toBeVisible(); // files remain (we didn't reload!)
      expect(didReload).toBe(false);

      // work through backwards, since they disappear on click and throw the locators off!
      for (const link of (await removers.all()).toReversed()) {
        await link.click();
      }

      await page.locator('body').click(); // trying this one last time...
      await page.reload({waitUntil: 'domcontentloaded'});
      expect(didShowDialog).toBe(false);
      await expect(() => expect(didReload).toBe(true)).toPass();
      await expect(firstFilePreview).toBeHidden(); // files gone! (reload didn't resurrect them)
    });

    test('no edits, leave without bother', async ({page}) => {
      await page.goto(`/user2/file-uploads/_upload/main/`, {waitUntil: 'load'});
      page.on('dialog', () => {
        // wait until timeout :)
      });

      const firstFilePreview = page.getByText('file1.txt');
      await expect(firstFilePreview).toBeHidden();

      await page.locator('body').click(); // トウットウル ー♪
      await page.reload({waitUntil: 'domcontentloaded'}); // this would freeze if the page is waiting on an "Are you sure?" dialog
      await expect(firstFilePreview).toBeHidden();
    });
  });
});
