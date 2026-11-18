import { provideServiceWorker } from '@angular/service-worker';
import { isDevMode } from '@angular/core';
import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter, withComponentInputBinding, withHashLocation } from '@angular/router';
import { App } from './app/app';
import { routes } from './app/app.routes';

/** True only in the Pages build (angular.json `pages`), which ships the service worker (APP-15). */
declare const RIFTWEAVE_SERVICE_WORKER: boolean;
/** True only in the development configurations (angular.json), which set `rw` in the console (APP-48). */
declare const RIFTWEAVE_DEV: boolean;

bootstrapApplication(App, {
  providers: [provideRouter(routes, withComponentInputBinding(), withHashLocation()), provideServiceWorker('ngsw-worker.js', { enabled: RIFTWEAVE_SERVICE_WORKER, registrationStrategy: 'registerImmediately' })],
})
  .then((app) => {
    // A constant false in the production and Pages builds: the branch, and the console module
    // with it, is left out of those bundles.
    if (RIFTWEAVE_DEV && isDevMode()) {
      void import('./app/dev-console').then(({ installDevConsole }) => installDevConsole(app.injector));
    }
  })
  .catch((error: unknown) => console.error(error));
