export function dataUrlToFile(dataUrl, name) {
  const match = /^data:([^;,]+)?(;base64)?,([\s\S]*)$/.exec(dataUrl || '');
  if (!match) throw new Error('The server returned an invalid image.');
  const mime = match[1] || 'image/png';
  const decoded = match[2] ? atob(match[3]) : decodeURIComponent(match[3]);
  const bytes = Uint8Array.from(decoded, character => character.charCodeAt(0));
  return new File([bytes], name, { type: mime });
}

export async function postStage(stage, fields) {
  const form = new FormData();
  for (const [name, value] of Object.entries(fields)) {
    if (value !== undefined && value !== null) form.append(name, value);
  }
  let response;
  try {
    response = await fetch(`/api/${stage}`, { method: 'POST', body: form });
  } catch (error) {
    throw new Error(`Network error: ${error.message || 'Could not reach the API.'}`);
  }
  let body;
  try {
    body = await response.json();
  } catch {
    throw new Error(`API returned an invalid response (HTTP ${response.status}).`);
  }
  if (!response.ok) throw new Error(body?.error || `Request failed (HTTP ${response.status}).`);
  return body;
}
