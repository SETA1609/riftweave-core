// A titled card; its body is the projected content.
import { Component, input } from '@angular/core';

@Component({ selector: 'rw-card', templateUrl: './card.html', styleUrl: './card.css' })
export class Card {
  readonly title = input.required<string>();
}
