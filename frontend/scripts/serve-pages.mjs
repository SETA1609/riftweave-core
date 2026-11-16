// Serve the Pages build (npm run build:pages) under the Pages base path, /riftweave-core/, as
// GitHub Pages does, for the chromium-pages end-to-end project (APP-15). Static files only.
import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { extname, join, normalize } from 'node:path';

const root = new URL('../dist/riftweave/browser/', import.meta.url).pathname;
const base = '/riftweave-core/';
const port = Number(process.env['PORT'] ?? 4400);
const types = {
  '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css', '.json': 'application/json',
  '.wasm': 'application/wasm', '.svg': 'image/svg+xml', '.webmanifest': 'application/manifest+json',
};

createServer(async (request, response) => {
  const path = decodeURIComponent(new URL(request.url ?? '/', 'http://localhost').pathname);
  if (!path.startsWith(base)) {
    response.writeHead(404).end();
    return;
  }
  const relative = normalize(path.slice(base.length) || 'index.html').replace(/^(\.\.[/\\])+/, '');
  try {
    const body = await readFile(join(root, relative));
    response.writeHead(200, { 'content-type': types[extname(relative)] ?? 'application/octet-stream' }).end(body);
  } catch {
    response.writeHead(404).end();
  }
}).listen(port, '127.0.0.1', () => console.log(`Pages preview on http://localhost:${port}${base}`));
