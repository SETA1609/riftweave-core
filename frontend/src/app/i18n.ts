// The app's own text in the viewer's language (docs/backend/frontend.md#languages): one i18next
// instance with the locale files bundled, so a static build needs no fetch. AppContext.locale is
// the one source; `translate` reads it, so a template or a computed that calls it re-renders when
// the language changes. Ruleset content (labels, descriptions) is data and is not translated here.
import { DOCUMENT, Service, computed, effect, inject } from '@angular/core';
import { type RouterStateSnapshot, TitleStrategy } from '@angular/router';
import i18next from 'i18next';
import de from '../locales/de.json';
import en from '../locales/en.json';
import es from '../locales/es.json';
import { AppContext, FALLBACK_LANGUAGE, offeredLanguage } from './context';

const RESOURCES = { en: { translation: en }, de: { translation: de }, es: { translation: es } };

/** The values a key fills in, as `{{name}}` in the locale file. */
export type Values = Readonly<Record<string, string | number>>;

@Service()
export class I18n {
  private readonly context = inject(AppContext);
  private readonly document = inject(DOCUMENT);
  private readonly instance = i18next.createInstance();
  /** The language shown: the chosen one if it is offered, else English. */
  readonly language = computed(() => offeredLanguage(this.context.locale()) ?? FALLBACK_LANGUAGE);

  constructor() {
    // Synchronous with bundled resources; Angular escapes what a template shows, so i18next does not.
    void this.instance.init({ resources: RESOURCES, lng: this.language(), fallbackLng: FALLBACK_LANGUAGE, initAsync: false, interpolation: { escapeValue: false } });
    effect(() => {
      const language = this.language();
      void this.instance.changeLanguage(language);
      this.document.documentElement.lang = language;
    });
  }

  /** The text of a key in the shown language, with its values filled in; English when the
   * language lacks the key. */
  translate(key: string, values: Values = {}): string {
    return this.instance.t(key, { ...values, lng: this.language() }) as string;
  }
}

/** The page title: a route's `title` is a key, shown in the viewer's language and changed with it. */
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
