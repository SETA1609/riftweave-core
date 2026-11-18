// The settings file both apps trade (APP-25, docs/backend/settings.md,
// backend/contract/settings.schema.json). The rules here are the Go CLI's
// (backend/internal/settings): the same checks, the same messages, and the same layout, so a file
// exported by either app imports into the other and exports back byte for byte. Keys this app
// does not use are kept.

export const SETTINGS_VERSION = 1;
export type SettingsFile = Record<string, unknown> & { riftweave: 'settings'; version: number };
export type Parsed = { settings: SettingsFile } | { error: string };

const LOCALE = /^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$/;
const MODULE = /^[a-z0-9][a-z0-9_-]*$/;
const THEMES = ['light', 'dark', 'system'];
const NETWORK_MODES = ['local', 'server', 'room'];
const TELEMETRY_MODES = ['off', 'local', 'export'];
const ENDPOINT = /^https?:\/\//;

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/** The settings of a fresh install: English, the system's colours, no optional modules, offline, no telemetry. */
export function defaultSettings(): SettingsFile {
  return { riftweave: 'settings', version: SETTINGS_VERSION, locale: 'en', theme: 'system', modules: [], network: { mode: 'local' }, telemetry: { mode: 'off' } };
}

/** Read a settings file and check it; a file that breaks a rule is refused with a message. */
export function parseSettings(text: string): Parsed {
  let file: unknown;
  try {
    file = JSON.parse(text);
  } catch {
    return { error: 'not a settings file: it is not a JSON object' };
  }
  if (!isObject(file)) return { error: 'not a settings file: it is not a JSON object' };
  if (file['riftweave'] !== 'settings') return { error: `not a settings file: "riftweave" is ${String(file['riftweave'])}, not "settings"` };
  if (file['version'] !== SETTINGS_VERSION) {
    return { error: `settings version ${String(file['version'])} is not one this app reads (it reads ${SETTINGS_VERSION})` };
  }
  if ('locale' in file && (typeof file['locale'] !== 'string' || !LOCALE.test(file['locale']))) {
    return { error: `locale ${String(file['locale'])} is not a language code` };
  }
  if ('theme' in file && !THEMES.includes(file['theme'] as string)) return { error: `theme ${String(file['theme'])} is not light, dark, or system` };
  if ('modules' in file) {
    const modules = file['modules'];
    if (!Array.isArray(modules)) return { error: 'modules is not a list' };
    const seen: string[] = [];
    for (const module of modules) {
      if (typeof module !== 'string' || !MODULE.test(module)) return { error: `module ${String(module)} is not a module id` };
      if (module === 'core') return { error: 'modules lists core, which is always loaded and cannot be turned off' };
      if (seen.includes(module)) return { error: `modules lists ${module} twice` };
      seen.push(module);
    }
  }
  if ('network' in file) {
    const network = file['network'];
    if (!isObject(network) || !NETWORK_MODES.includes(network['mode'] as string)) return { error: 'network.mode is not local, server, or room' };
    for (const field of ['server_url', 'room_code']) {
      if (field in network && typeof network[field] !== 'string') return { error: `network.${field} is not text` };
    }
  }
  if ('bestiary' in file) {
    const bestiary = file['bestiary'];
    if (!isObject(bestiary)) return { error: 'bestiary is not an object' };
    if ('players_see_only_known' in bestiary && typeof bestiary['players_see_only_known'] !== 'boolean') {
      return { error: 'bestiary.players_see_only_known is not true or false' };
    }
  }
  if ('telemetry' in file) {
    const telemetry = file['telemetry'];
    if (!isObject(telemetry) || !TELEMETRY_MODES.includes(telemetry['mode'] as string)) return { error: 'telemetry.mode is not off, local, or export' };
    if ('endpoint' in telemetry && (typeof telemetry['endpoint'] !== 'string' || !ENDPOINT.test(telemetry['endpoint']))) {
      return { error: 'telemetry.endpoint is not an http or https URL' };
    }
  }
  return { settings: file as SettingsFile };
}

/** Every object's keys sorted, as Go's encoder writes a map. */
function sorted(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sorted);
  if (isObject(value)) return Object.fromEntries(Object.keys(value).sort().map((key) => [key, sorted(value[key])]));
  return value;
}

/** Write the file the one way both apps write it: keys sorted, two-space indent, a final newline. */
export function encodeSettings(file: SettingsFile): string {
  return JSON.stringify(sorted(file), null, 2) + '\n';
}
