import test from 'node:test';
import assert from 'node:assert/strict';
import { dataUrlToFile, postStage } from './api.js';

test('data URL becomes a PNG file suitable for the next multipart stage', async () => {
  const file = dataUrlToFile('data:image/png;base64,iVBORw==', 'watermarked.png');
  assert.equal(file.name, 'watermarked.png');
  assert.equal(file.type, 'image/png');
  assert.deepEqual([...new Uint8Array(await file.arrayBuffer())], [137, 80, 78, 71]);
});

test('postStage sends files and fields as multipart without overriding the boundary', async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (url, options) => {
    assert.equal(url, '/api/embed');
    assert.equal(options.method, 'POST');
    assert.equal(options.headers, undefined);
    assert.equal(options.body.get('key'), 'stegno');
    assert.equal(options.body.get('image').name, 'input.png');
    return new Response(JSON.stringify({ watermarked: 'data:image/png;base64,AA==' }), { status: 200 });
  };
  try {
    assert.ok((await postStage('embed', { image: new File(['a'], 'input.png'), key: 'stegno' })).watermarked);
  } finally { globalThis.fetch = originalFetch; }
});

test('postStage exposes backend error body', async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => new Response(JSON.stringify({ error: 'image is too small' }), { status: 400 });
  try {
    await assert.rejects(postStage('embed', { key: 'stegno' }), /image is too small/);
  } finally { globalThis.fetch = originalFetch; }
});

test('postStage describes network errors', async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => { throw new TypeError('Failed to fetch'); };
  try {
    await assert.rejects(postStage('embed', { key: 'stegno' }), /Network error.*Failed to fetch/);
  } finally { globalThis.fetch = originalFetch; }
});
