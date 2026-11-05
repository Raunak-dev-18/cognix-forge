// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// initModalServiceMessage tries to retrieve session storage and fills it in the input fields,
// on blur it will set the storage again
export function initModalServiceMessageForm() {
  const smForm: HTMLElement = document.querySelector('.service-message.form');
  if (!smForm) return;

  const nav: HTMLElement = document.querySelector('#navbar');
  const modal: HTMLElement = document.querySelector('#service-message-modal');
  if (modal !== null) {
    const parent = nav.parentNode;
    parent.insertBefore(modal, nav);
  }

  const storage: Storage = sessionStorage;
  const titleVal: string = storage.getItem('smTitle');
  const textVal: string = storage.getItem('smText');

  const form: HTMLFormElement = document.querySelector('.service-message.form');
  const title = form.elements.namedItem('title') as HTMLInputElement;
  title.value = titleVal;

  const text = form.elements.namedItem('text') as HTMLInputElement;
  text.value = textVal;

  title.addEventListener('blur', () => {
    storage.setItem('smTitle', title.value);
  });

  text.addEventListener('blur', () => {
    storage.setItem('smText', text.value);
  });
}

// initPreviewServiceMessageButton replaces the action with a call to /preview so the content gets rendered properly
// it also saves the values of the title and text field in storage so its not lost
export function initPreviewServiceMessageButton() {
  const previewSMButton = document.querySelector('#preview-service-message-button');
  previewSMButton?.addEventListener('click', () => {
    const form: HTMLFormElement = document.querySelector('.service-message.form');
    if (!form.checkValidity()) {
      form.reportValidity();
      return false;
    }

    const action = form.action;
    const newAct = action.replace('/admin/service_message?sm_type=modal', '/admin/service_message/preview?sm_type=modal&status=show');
    form.action = newAct;
  });
}

export function initDeleteServiceMessageButton() {
  const deleteSMButton = document.querySelector('#delete-sm-button');
  deleteSMButton?.addEventListener('click', () => {
    const storage: Storage = sessionStorage;
    storage.setItem('smTitle', '');
    storage.setItem('smText', '');
  });
}
