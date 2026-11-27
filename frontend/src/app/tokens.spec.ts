// The design guide's tokens in src/styles.css (FE_S22_02, docs/design/README.md#tokens): every text
// and background pair the shared classes draw reaches WCAG AA, 4.5:1 for text and 3:1 for a
// keyline the eye must find, in both the light and the dark set. The values are read from the
// file itself, so a change to a token is checked here.
import { describe, expect, it } from 'vitest';

/** node:fs, loaded at run time where Vitest runs on Node; the app has no Node types. */
const NODE_FS = 'node:fs';
const fs = (await import(/* @vite-ignore */ NODE_FS)) as { readFileSync(path: string, encoding: 'utf8'): string };
const CSS = fs.readFileSync(`${(globalThis as unknown as { process: { cwd(): string } }).process.cwd()}/src/styles.css`, 'utf8');

/** The hex tokens of one rule block: the light set is `:root {`, the dark set `:root[data-theme="dark"] {`. */
function tokens(selector: string): Record<string, string> {
  const start = CSS.indexOf(`${selector} {`);
  const block = CSS.slice(start, CSS.indexOf('\n}', start));
  return Object.fromEntries([...block.matchAll(/--([a-z-]+):\s*(#[0-9a-f]{6})/g)].map((match) => [match[1], match[2]]));
}

function luminance(hex: string): number {
  const [r, g, b] = [1, 3, 5].map((at) => parseInt(hex.slice(at, at + 2), 16) / 255).map((c) => (c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** The WCAG 2 contrast ratio of two colors. */
export function contrast(a: string, b: string): number {
  const [light, dark] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (light + 0.05) / (dark + 0.05);
}

const LIGHT = tokens(':root');
const DARK = { ...LIGHT, ...tokens(':root[data-theme="dark"]') };
const SURFACES = ['canvas', 'surface-container-low', 'surface-container', 'surface-container-high'];
const TEXT = ['on-surface', 'muted', 'primary', 'secondary', 'tertiary', 'danger'];
/** Text drawn on its own fill: a button, an alert, a tag. */
const ON_FILL: [string, string][] = [
  ['on-primary', 'primary'], ['on-primary', 'primary-hover'], ['danger', 'danger-bg'], ['primary', 'success-bg'], ['secondary', 'info-bg'], ['tertiary', 'tertiary-bg'],
];

describe.each([['light', LIGHT], ['dark', DARK]] as const)('the %s tokens', (_, set) => {
  it.each(TEXT.flatMap((text) => SURFACES.map((surface) => [text, surface])))('%s on %s reaches 4.5:1', (text, surface) => {
    expect(contrast(set[text], set[surface])).toBeGreaterThanOrEqual(4.5);
  });

  it.each(ON_FILL)('%s on %s reaches 4.5:1', (text, fill) => {
    expect(contrast(set[text], set[fill])).toBeGreaterThanOrEqual(4.5);
  });

  it('draw the strong keyline at 3:1 on every surface', () => {
    for (const surface of SURFACES) expect(contrast(set['outline'], set[surface]), surface).toBeGreaterThanOrEqual(3);
  });
});

describe('the tokens', () => {
  it('give the dark set the same tokens as the system dark rule', () => {
    const system = CSS.slice(CSS.indexOf(':root:not([data-theme="light"]) {'));
    const dark = tokens(':root[data-theme="dark"]');
    for (const [name, value] of Object.entries(dark)) expect(system, name).toContain(`--${name}: ${value}`);
  });

  it('name the nine colors and the five phases', () => {
    for (const color of ['red', 'orange', 'yellow', 'green', 'blue', 'indigo', 'violet', 'white', 'black']) expect(LIGHT[`color-${color}`], color).toMatch(/^#/);
    for (const phase of ['wood', 'fire', 'earth', 'metal', 'water']) expect(LIGHT[`phase-${phase}`], phase).toMatch(/^#/);
  });
});
