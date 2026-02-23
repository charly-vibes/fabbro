import { fetchContent } from './fetch.js';
import { mount as mountEditor } from './editor.js';
import { mount as mountExport } from './export.js';

const app = document.getElementById('app');

const session = {
  content: '',
  sourceUrl: '',
  filename: '',
  annotations: [],
};

function renderLanding() {
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
    </div>
  `;

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
      renderEditor();
    } catch (err) {
      error.textContent = err.message;
      urlBtn.disabled = false;
      urlBtn.textContent = 'Start review';
    }
  });

  pasteBtn.addEventListener('click', () => {
    const text = pasteInput.value;
    if (!text) { error.textContent = 'Paste some text first.'; return; }
    session.content = text.replace(/\r\n/g, '\n');
    session.sourceUrl = '';
    session.filename = 'pasted-text';
    session.annotations = [];
    renderEditor();
  });
}

function renderEditor() {
  mountEditor(app, session, { onFinish: renderExport });
}

function renderExport() {
  mountExport(app, session, { onBack: renderEditor });
}

// Boot
renderLanding();
