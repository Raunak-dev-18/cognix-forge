// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later
import type TomSelectType from 'tom-select';
import type {TomSettings} from 'tom-select/dist/esm/types/index.ts';

export async function createTomSelect(el: HTMLInputElement, opts: Partial<TomSettings> = {}): Promise<TomSelectType> {
  const {default: TomSelect} = await import(/* webpackChunkName: "tom-select" */'tom-select');
  const ts = new TomSelect(el, {
    ...opts,
    onItemAdd(...args: unknown[]) {
      ts.setTextboxValue('');
      opts.onItemAdd?.apply(this, args);
    },
  });

  // Handle comma key to create item immediately
  ts.control_input.addEventListener('keydown', (e: KeyboardEvent) => {
    if (e.key === ',') {
      e.preventDefault();
      const value = ts.inputValue().trim();
      if (value) {
        // Use addItem if option exists, otherwise createItem
        if (ts.options[value]) {
          ts.addItem(value);
        } else {
          ts.createItem(value);
        }
        ts.setTextboxValue('');
      }
    }
  });

  return ts;
}
