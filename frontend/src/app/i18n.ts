import { Service } from '@angular/core';

@Service()
export class I18n {
  translate(key: string, _values?: Record<string, string>): string {
    return key;
  }
}
