import { provideServiceWorker } from '@angular/service-worker';
import { bootstrapApplication } from '@angular/platform-browser';
import { App } from './app/app';

/** True only in the Pages build (angular.json `pages`), which ships the service worker (APP-15). */
declare const RIFTWEAVE_SERVICE_WORKER: boolean;

bootstrapApplication(App, {
  providers: [provideServiceWorker('ngsw-worker.js', { enabled: RIFTWEAVE_SERVICE_WORKER, registrationStrategy: 'registerImmediately' })],
}).catch((error: unknown) => console.error(error));
