// See translate-pipe.md.
import { Pipe, type PipeTransform, inject } from '@angular/core';
import { I18n, type Said, type Values } from './i18n';

@Pipe({ name: 't', pure: false })
export class TranslatePipe implements PipeTransform {
  private readonly i18n = inject(I18n);

  transform(key: string | Said | null, values?: Values): string {
    if (key !== null && typeof key === 'object') return this.i18n.translate(key.key, key.values);
    return key ? this.i18n.translate(key, values) : '';
  }
}
