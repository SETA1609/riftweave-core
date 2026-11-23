// The file both apps trade (APP-30, docs/backend/files.md, backend/contract/file.json): one
// envelope, { riftweave: kind, rules_version, modules, data }, around what the command contract
// already returns. The rules and the messages are the Go CLI's (backend/cmd/riftweave/files), so a
// file exported by either app imports into the other and exports back byte for byte.

import type { PrintInput, StoredActor } from './api';

export const FILE_KINDS = ['character', 'session', 'spell'] as const;
export type FileKind = (typeof FILE_KINDS)[number];
export interface FileModule { id: string; version: string }
export interface RiftweaveFile { riftweave: FileKind; rules_version: string; modules: FileModule[]; data: unknown }
export interface MissingModule { module: string; rows: string[] }

/** What a printed sheet shows (BE_S09_03, docs/backend/print.md#printing-a-character): `current`
 * is the session's state, `fresh` the same character at full resources with no condition. */
export const PRINT_STATE = { Current: 'current', Fresh: 'fresh' } as const;
export type PrintState = (typeof PRINT_STATE)[keyof typeof PRINT_STATE];
export const PRINT_KIND_SHEET = 'sheet';

/** The `print` input for a character's sheet, its stored actor as saved. */
export function sheetPrint(actor: StoredActor, state: PrintState): PrintInput {
  return { kind: PRINT_KIND_SHEET, actor, state } as PrintInput;
}

/** Opens a printed page in a new tab and its print dialog: PDF through the browser, offline with
 * the core in the page. The page is the `print` command's HTML, inline CSS and no script. */
export function printPage(html: string): void {
  const url = URL.createObjectURL(new Blob([html], { type: 'text/html' }));
  const printout = window.open(url, '_blank');
  printout?.addEventListener('load', () => printout.print(), { once: true });
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/** Every object's keys sorted, as Go's encoder writes a map. */
function sorted(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sorted);
  if (isObject(value)) return Object.fromEntries(Object.keys(value).sort().map((key) => [key, sorted(value[key])]));
  return value;
}

/** Write a file the one way both apps write it: keys sorted, two-space indent, a final newline. */
export function encodeFile(file: RiftweaveFile): string {
  return JSON.stringify(sorted(file), null, 2) + '\n';
}

/** A file of the kind, made with these rules and modules (sorted by id). */
export function makeFile(kind: FileKind, data: unknown, rulesVersion: string, modules: FileModule[] = []): RiftweaveFile {
  return { riftweave: kind, rules_version: rulesVersion, modules: [...modules].sort((a, b) => a.id.localeCompare(b.id)), data };
}

/** Read a file's envelope; the data is checked by the core next. A settings file is reported as
 * such, since it keeps its own format (APP-25). */
export function parseFile(text: string): { file: RiftweaveFile } | { settings: true } | { error: string } {
  let value: unknown;
  try {
    value = JSON.parse(text);
  } catch {
    return { error: 'not a Riftweave file: it is not a JSON object' };
  }
  if (!isObject(value)) return { error: 'not a Riftweave file: it is not a JSON object' };
  if (value['riftweave'] === 'settings') return { settings: true };
  if (!FILE_KINDS.includes(value['riftweave'] as FileKind)) {
    return { error: `not a Riftweave file: "riftweave" is ${JSON.stringify(value['riftweave'] ?? '')}, not one of ${FILE_KINDS.join(', ')}` };
  }
  if (typeof value['rules_version'] !== 'string' || !Array.isArray(value['modules']) || !('data' in value)) {
    return { error: 'not a Riftweave file: it needs rules_version, modules, and data' };
  }
  return { file: value as unknown as RiftweaveFile };
}

/** The module refs in the data, by module id, sorted: the rows a device without it would lose. */
export function moduleRows(data: unknown): Record<string, string[]> {
  const found: Record<string, string[]> = {};
  const walk = (node: unknown): void => {
    if (typeof node === 'string' && node.startsWith('module:')) {
      const id = node.slice('module:'.length).split('/')[0];
      const rows = (found[id] ??= []);
      if (!rows.includes(node)) rows.push(node);
    } else if (Array.isArray(node)) {
      node.forEach(walk);
    } else if (isObject(node)) {
      Object.values(node).forEach(walk);
    }
  };
  walk(data);
  for (const id of Object.keys(found)) found[id].sort();
  return found;
}

/** The modules the file names that are not enabled here, with the rows at stake. */
export function missingModules(file: RiftweaveFile, enabled: readonly string[]): MissingModule[] {
  const rows = moduleRows(file.data);
  return file.modules.filter((module) => !enabled.includes(module.id)).map((module) => ({ module: module.id, rows: rows[module.id] ?? [] }));
}

/** The alert for the missing modules, word for word the CLI's. */
export function moduleAlert(missing: MissingModule[]): string {
  return missing.map((entry) => `module ${entry.module} is not enabled here (${entry.rows.length ? entry.rows.join(', ') : 'no rows named'})`).join('; ') + ': enable it, or confirm the loss';
}

/** The warning for a file from other rules, or null. */
export function rulesWarning(file: RiftweaveFile, rulesVersion: string): string | null {
  return file.rules_version === rulesVersion ? null : `made with rules ${file.rules_version}; this app runs ${rulesVersion}`;
}

/** Offer a file as a download. */
export function download(name: string, text: string): void {
  const link = document.createElement('a');
  link.href = URL.createObjectURL(new Blob([text], { type: 'application/json' }));
  link.download = name;
  link.click();
  URL.revokeObjectURL(link.href);
}
