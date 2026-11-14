// A message box, shown only when there is a message.
import { Component, input } from '@angular/core';

@Component({ selector: 'rw-alert', templateUrl: './alert.html', styleUrl: './alert.css' })
export class Alert {
  readonly message = input('');
  readonly kind = input<'danger' | 'success' | 'info'>('danger');
}
