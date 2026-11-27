// The routes (docs/backend/app-layout.md): the menus of backend/contract/menu.json, each a menu grid,
// and one route per screen. A route's data sets the component's inputs; its title is a locale key
// (TranslatedTitle in i18n.ts), except the game's name.
import type { Routes } from '@angular/router';
import { Attack } from './attack';
import { Bestiary } from './bestiary/bestiary';
import { CharacterList } from './character-list';
import { FileImport } from './file-import';
import { Playthrough } from './playthrough';
import { Resolution } from './resolution';
import { SavedGames } from './saved-games';
import { SettingsScreens } from './settings-screens';
import { Sheet } from './sheet';
import { Spells } from './spells';
import { MenuGrid } from './ui/menu-grid';

export const routes: Routes = [
  { path: '', component: MenuGrid, data: { parent: 'main' }, title: 'Riftweave' },
  { path: 'characters', component: CharacterList, title: 'menu.characters' },
  { path: 'settings', component: SettingsScreens, data: { section: 'settings' }, title: 'menu.settings' },
  { path: 'modules', component: SettingsScreens, data: { section: 'modules' }, title: 'menu.modules' },
  { path: 'network', component: SettingsScreens, data: { section: 'network' }, title: 'menu.network' },
  { path: 'import', component: FileImport, title: 'menu.import' },
  { path: 'play', component: MenuGrid, data: { parent: 'start' }, title: 'menu.start' },
  { path: 'play/gm_tools', component: MenuGrid, data: { parent: 'gm_tools' }, title: 'menu.gm_tools' },
  { path: 'play/character_creation', component: Sheet, title: 'menu.character_creation' },
  { path: 'play/skill_resolution', component: Resolution, title: 'menu.skill_resolution' },
  { path: 'play/spell_creation', component: Spells, title: 'menu.spell_creation' },
  { path: 'play/saved_games', component: SavedGames, title: 'menu.saved_games' },
  { path: 'play/attack', component: Attack, title: 'menu.attack' },
  { path: 'play/bestiary', component: Bestiary, title: 'menu.bestiary' },
  { path: 'play/playthrough', component: Playthrough, title: 'menu.playthrough' },
  { path: '**', redirectTo: '' },
];
