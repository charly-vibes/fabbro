import { fetchContent } from './fetch.js';
import { parse as parseFem } from './fem.js';
import { mount as mountEditor } from './editor.js';
import { mount as mountExport } from './export.js';
import { mount as mountApply } from './apply.js';
import * as storage from './storage.js';
import * as tutorial from './tutorial.js';

function stripFrontmatter(text) {
  if (!text.startsWith('---')) return text;
  const end = text.indexOf('\n---', 3);
  if (end === -1) return text;
  return text.slice(end + 4).replace(/^\n/, '');
}

function lineToOffsets(content, startLine, endLine) {
  const lines = content.split('\n');
  let offset = 0;
  let startOffset = 0;
  let endOffset = 0;
  for (let i = 0; i < lines.length; i++) {
    if (i + 1 === startLine) startOffset = offset;
    if (i + 1 === endLine) {
      endOffset = offset + lines[i].length;
      break;
    }
    offset += lines[i].length + 1;
  }
  return { startOffset, endOffset };
}

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
  tutorial.stop();
  const recent = await storage.listSessions();

  app.innerHTML = `
    <div class="landing">
      <h1>fabbro</h1>
      <p class="subtitle">Review any text. Annotate with comments. Export a structured summary.</p>
      <button id="tutorial-btn" class="tutorial-start">New here? Try the interactive tutorial →</button>
      <input type="text" id="url-input" placeholder="Paste a GitHub file URL or any web page URL">
      <button id="url-btn">Start review</button>
      <div class="divider">— or —</div>
      <textarea id="paste-input" placeholder="Paste text directly"></textarea>
      <button id="paste-btn">Start review</button>
      <div class="divider">— or —</div>
      <div class="drop-zone" id="drop-zone">Drop a .md, .txt, .fem, or code file here</div>
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

  document.getElementById('tutorial-btn').addEventListener('click', async () => {
    session.content = tutorial.SAMPLE_CONTENT;
    session.sourceUrl = '';
    session.filename = 'greeting.js (tutorial)';
    session.annotations = [];
    await startSession();
    tutorial.start(renderLanding);
  });

  const dropZone = document.getElementById('drop-zone');
  const acceptedExts = new Set([
    '.md', '.txt', '.fem', '.go', '.py', '.js', '.ts', '.rs', '.rb', '.java',
    '.c', '.h', '.cpp', '.hpp', '.css', '.html', '.json', '.yaml', '.yml',
    '.toml', '.xml', '.sh', '.bash', '.zsh', '.fish', '.sql', '.lua',
    '.ex', '.exs', '.zig', '.nim', '.kt', '.swift', '.r',
  ]);

  dropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropZone.classList.add('drop-zone--active');
  });

  dropZone.addEventListener('dragenter', (e) => {
    e.preventDefault();
    dropZone.classList.add('drop-zone--active');
  });

  dropZone.addEventListener('dragleave', () => {
    dropZone.classList.remove('drop-zone--active');
  });

  dropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropZone.classList.remove('drop-zone--active');
    const file = e.dataTransfer.files[0];
    if (!file) return;

    const name = file.name;
    const dotIdx = name.lastIndexOf('.');
    const ext = dotIdx >= 0 ? name.slice(dotIdx).toLowerCase() : '';
    if (!acceptedExts.has(ext)) {
      error.textContent = 'Unsupported file type. Please use text or code files.';
      return;
    }

    error.textContent = '';
    const reader = new FileReader();
    reader.onload = async () => {
      const raw = reader.result.replace(/\r\n/g, '\n');
      if (ext === '.fem') {
        const stripped = stripFrontmatter(raw);
        const { annotations: femAnnotations, cleanContent } = parseFem(stripped);
        session.content = cleanContent;
        session.sourceUrl = '';
        session.filename = name;
        session.annotations = femAnnotations.map(a => {
          const { startOffset, endOffset } = lineToOffsets(cleanContent, a.startLine, a.endLine);
          return {
            type: a.type === 'change' ? 'suggest' : a.type,
            text: a.text,
            startOffset,
            endOffset,
          };
        });
        renderApply();
      } else {
        session.content = raw;
        session.sourceUrl = '';
        session.filename = name;
        session.annotations = [];
        await startSession();
      }
    };
    reader.readAsText(file);
  });
}

function renderEditor() {
  mountEditor(app, session, { onFinish: renderExport, onChanged: scheduleSave });
}

function renderExport() {
  mountExport(app, session, { onBack: renderEditor });
}

function renderApply() {
  mountApply(app, session, { onBack: renderLanding });
}

// Boot
storage.init().then(renderLanding);
