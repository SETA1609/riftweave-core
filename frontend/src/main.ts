import { provideServiceWorker } from '@angular/service-worker';
import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter, withComponentInputBinding, withHashLocation } from '@angular/router';
import { App } from './app/app';
import { routes } from './app/app.routes';

/** True only in the Pages build (angular.json `pages`), which ships the service worker (APP-15). */
declare const RIFTWEAVE_SERVICE_WORKER: boolean;

bootstrapApplication(App, {
  providers: [provideRouter(routes, withComponentInputBinding(), withHashLocation()), provideServiceWorker('ngsw-worker.js', { enabled: RIFTWEAVE_SERVICE_WORKER, registrationStrategy: 'registerImmediately' })],
}).catch((error: unknown) => console.error(error));
