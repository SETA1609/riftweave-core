// The routes (docs/backend/app-layout.md): the menus of backend/contract/menu.json, each a menu grid,
// and one route per screen. A route's data sets the component's inputs.
import type { Routes } from '@angular/router';
import { Attack } from './attack';
import { CharacterList } from './character-list';
import { FileImport } from './file-import';
import { Playthrough } from './playthrough';
import { Resolution } from './resolution';
import { SavedGames } from './saved-games';
import { SettingsScreens } from './settings-screens';
import { Sheet } from './sheet';
import { MenuGrid } from './ui/menu-grid';

export const routes: Routes = [
  { path: '', component: MenuGrid, data: { parent: 'main' }, title: 'Riftweave' },
  { path: 'characters', component: CharacterList, title: 'Characters' },
  { path: 'settings', component: SettingsScreens, data: { section: 'settings' }, title: 'Settings' },
  { path: 'modules', component: SettingsScreens, data: { section: 'modules' }, title: 'Modules' },
  { path: 'network', component: SettingsScreens, data: { section: 'network' }, title: 'Network' },
  { path: 'import', component: FileImport, title: 'Import' },
  { path: 'play', component: MenuGrid, data: { parent: 'start' }, title: 'Start' },
  { path: 'play/gm_tools', component: MenuGrid, data: { parent: 'gm_tools' }, title: 'GM tools' },
  { path: 'play/character_creation', component: Sheet, title: 'Character creation' },
  { path: 'play/skill_resolution', component: Resolution, title: 'Skill resolution' },
  { path: 'play/saved_games', component: SavedGames, title: 'Saved games' },
  { path: 'play/attack', component: Attack, title: 'Resolve an attack' },
  { path: 'play/playthrough', component: Playthrough, title: 'Playthrough' },
  { path: '**', redirectTo: '' },
];
