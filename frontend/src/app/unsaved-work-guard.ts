// See unsaved-work-guard.md.
import { inject } from '@angular/core';
import type { CanDeactivateFn } from '@angular/router';
import { I18n } from './i18n';

/** A screen that holds work it has not saved yet. */
export interface HoldsUnsavedWork {
  hasUnsavedWork(): boolean;
}

/** The question asked before leaving; a test replaces it. */
export const ask = { confirm: (message: string): boolean => globalThis.confirm?.(message) ?? true };

export const unsavedWorkGuard: CanDeactivateFn<HoldsUnsavedWork> = (screen) => {
  if (!screen?.hasUnsavedWork?.()) return true;
  return ask.confirm(inject(I18n).translate('screen.leave_unsaved'));
};
