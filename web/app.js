import { fetchContent } from './fetch.js';
import { mount as mountEditor } from './editor.js';
import { mount as mountExport } from './export.js';
import * as storage from './storage.js';

const app = document.getElementById('app');

const session = {
  id: null,
  content: '',
  sourceUrl: '',
  filename: '',
  annotations: [],
};

let saveTimer = null;

function scheduleSave() {
  if (!session.id) return;
  clearTimeout(saveTimer);
  saveTimer = setTimeout(async () => {
    await storage.saveSession(session.id, session);
    const indicator = document.getElementById('save-indicator');
    if (indicator) indicator.textContent = 'Saved locally';
  }, 500);
}

async function startSession() {
  session.id = await storage.createSession(session);
  renderEditor();
}

async function renderLanding() {
  const recent = await storage.listSessions();

  app.innerHTML = `
    <div class="landing">
      <h1>fabbro</h1>
      <p class="subtitle">Review any text. Annotate with comments. Export a structured summary.</p>
      <input type="text" id="url-input" placeholder="Paste a GitHub file URL or any web page URL">
      <button id="url-btn">Start review</button>
      <div class="divider">— or —</div>
      <textarea id="paste-input" placeholder="Paste text directly"></textarea>
      <button id="paste-btn">Start review</button>
      <div id="error" class="error"></div>
      ${recent.length > 0 ? `
        <div class="divider">— recent sessions —</div>
        <div class="recent-sessions" id="recent-sessions"></div>
      ` : ''}
    </div>
  `;

  if (recent.length > 0) {
    const container = document.getElementById('recent-sessions');
    for (const s of recent) {
      const row = document.createElement('div');
      row.className = 'recent-row';

      const info = document.createElement('button');
      info.className = 'recent-btn';
      const date = new Date(s.updatedAt).toLocaleDateString();
      info.textContent = `${s.filename} — ${s.annotationCount} note${s.annotationCount !== 1 ? 's' : ''} — ${date}`;
      info.addEventListener('click', async () => {
        const record = await storage.loadSession(s.id);
        if (record) {
          session.id = record.id;
          session.content = record.content;
          session.sourceUrl = record.sourceUrl;
          session.filename = record.filename;
          session.annotations = record.annotations;
          renderEditor();
        }
      });

      const del = document.createElement('button');
      del.className = 'recent-delete';
      del.textContent = '✕';
      del.title = 'Delete session';
      del.addEventListener('click', async (e) => {
        e.stopPropagation();
        await storage.deleteSession(s.id);
        renderLanding();
      });

      row.appendChild(info);
      row.appendChild(del);
      container.appendChild(row);
    }
  }

  const urlInput = document.getElementById('url-input');
  const pasteInput = document.getElementById('paste-input');
  const urlBtn = document.getElementById('url-btn');
  const pasteBtn = document.getElementById('paste-btn');
  const error = document.getElementById('error');

  urlBtn.addEventListener('click', async () => {
    const url = urlInput.value.trim();
    if (!url) { error.textContent = 'Enter a URL.'; return; }
    error.textContent = '';
    urlBtn.disabled = true;
    urlBtn.textContent = 'Fetching…';
    try {
      const result = await fetchContent(url);
      session.content = result.content.replace(/\r\n/g, '\n');
      session.sourceUrl = result.source;
      session.filename = result.filename;
      session.annotations = [];
      await startSession();
    } catch (err) {
      error.textContent = err.message;
      urlBtn.disabled = false;
      urlBtn.textContent = 'Start review';
    }
  });

  pasteBtn.addEventListener('click', async () => {
    const text = pasteInput.value;
    if (!text) { error.textContent = 'Paste some text first.'; return; }
    session.content = text.replace(/\r\n/g, '\n');
    session.sourceUrl = '';
    session.filename = 'pasted-text';
    session.annotations = [];
    await startSession();
  });
}

function renderEditor() {
  mountEditor(app, session, { onFinish: renderExport, onChanged: scheduleSave });
}

function renderExport() {
  mountExport(app, session, { onBack: renderEditor });
}

// Boot
storage.init().then(renderLanding);
