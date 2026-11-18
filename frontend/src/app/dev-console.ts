// The app's objects in the browser console, in a dev build only (APP-48, docs/backend/dev-mode.md#in-the-browser-console):
// `rw.context`, `rw.api`, `rw.saves`, `rw.router`, and `rw.screen()`, so a developer can try a
// state from the console as `--dev` lets one from the CLI. Nothing here is a rule: a change made
// through `rw` goes through the same context and API a screen uses. `globalThis.riftweave` is the
// Go core in the page; this is `globalThis.rw`.
import { type Injector } from '@angular/core';
import { Router } from '@angular/router';
import { API, type Api } from './api';
import { AppContext } from './context';
import { Saves } from './saves';

export interface DevConsole {
  /** The app context's signals: `rw.context.theme.set('dark')`. */
  readonly context: AppContext;
  /** The API the screens share: `await rw.api.derived({ attributes: { end: 6 }, level: 3 })`. */
  readonly api: Api;
  /** The saves: `rw.saves.list()`, `await rw.saves.resume(id)`. */
  readonly saves: Saves;
  /** The router: `rw.router.navigateByUrl('/settings')`. */
  readonly router: Router;
  /** The current screen's component, or null where Angular's dev tools are not on the page. */
  screen(): unknown;
  /** What `rw` holds, one line each; also printed to the console. */
  help(): string;
}

const HELP = [
  'rw — the Riftweave app in a dev build (APP-48). A change goes through the context or the API, as a screen\'s does.',
  '  rw.context   the app context: theme, locale, the settings file, as signals (rw.context.theme.set("dark"))',
  '  rw.api       the API: await rw.api.derived({ attributes: { end: 6 }, level: 3 })',
  '  rw.saves     the saves: rw.saves.list(), await rw.saves.resume(id)',
  '  rw.router    the router: rw.router.navigateByUrl("/settings")',
  '  rw.screen()  the component of the current screen',
  '  rw.help()    this text',
].join('\n');

/** The console object over an app's injector. */
export function devConsole(injector: Injector, document: Document = globalThis.document): DevConsole {
  return {
    context: injector.get(AppContext),
    api: injector.get(API),
    saves: injector.get(Saves),
    router: injector.get(Router),
    screen() {
      // Angular's dev-mode utilities (globalThis.ng) find the component on the routed element.
      const element = document?.querySelector('router-outlet')?.nextElementSibling;
      const ng = (globalThis as { ng?: { getComponent?: (element: Element) => unknown } }).ng;
      return element && ng?.getComponent ? ng.getComponent(element) : null;
    },
    help() {
      console.info(HELP);
      return HELP;
    },
  };
}

/** Sets `rw` on the host (globalThis in the app), once. */
export function installDevConsole(injector: Injector, host: { rw?: DevConsole } = globalThis as { rw?: DevConsole }): DevConsole {
  host.rw = devConsole(injector);
  return host.rw;
}
