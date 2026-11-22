// See i18n.md.
import { DOCUMENT, Service, computed, effect, inject } from '@angular/core';
import { type RouterStateSnapshot, TitleStrategy } from '@angular/router';
import i18next from 'i18next';
import de from '../locales/de.json';
import en from '../locales/en.json';
import es from '../locales/es.json';
import { AppContext, FALLBACK_LANGUAGE, offeredLanguage } from './context';

const RESOURCES = { en: { translation: en }, de: { translation: de }, es: { translation: es } };

export type Values = Readonly<Record<string, string | number>>;

@Service()
export class I18n {
  private readonly context = inject(AppContext);
  private readonly document = inject(DOCUMENT);
  private readonly instance = i18next.createInstance();
  readonly language = computed(() => offeredLanguage(this.context.locale()) ?? FALLBACK_LANGUAGE);

  constructor() {
    void this.instance.init({ resources: RESOURCES, lng: this.language(), fallbackLng: FALLBACK_LANGUAGE, initAsync: false, interpolation: { escapeValue: false } });
    effect(() => {
      const language = this.language();
      void this.instance.changeLanguage(language);
      this.document.documentElement.lang = language;
    });
  }

  translate(key: string, values: Values = {}): string {
    return this.instance.t(key, { ...values, lng: this.language() }) as string;
  }
}

@Service()
export class TranslatedTitle extends TitleStrategy {
  private readonly i18n = inject(I18n);
  private readonly document = inject(DOCUMENT);
  private key = '';

  constructor() {
    super();
    effect(() => {
      this.i18n.language();
      if (this.key) this.document.title = this.i18n.translate(this.key);
    });
  }

  override updateTitle(snapshot: RouterStateSnapshot): void {
    this.key = this.buildTitle(snapshot) ?? '';
    if (this.key) this.document.title = this.i18n.translate(this.key);
  }
}
