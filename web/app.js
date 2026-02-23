import { fetchContent } from './fetch.js';

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
  app.innerHTML = `
    <div class="editor-header">
      <span>${escapeHtml(session.filename)}</span>
      <button id="finish-btn">Finish review</button>
    </div>
    <div class="viewer">
      <div class="lines" id="lines"></div>
    </div>
  `;

  renderLines();

  document.getElementById('finish-btn').addEventListener('click', renderExport);
  document.getElementById('lines').addEventListener('mouseup', handleSelection);
}

function renderLines() {
  const container = document.getElementById('lines');
  const lines = session.content.split('\n');
  let offset = 0;
  container.innerHTML = '';

  for (let i = 0; i < lines.length; i++) {
    const lineDiv = document.createElement('div');
    lineDiv.className = 'line';
    lineDiv.dataset.start = offset;

    const gutter = document.createElement('span');
    gutter.className = 'gutter';
    gutter.textContent = i + 1;

    const text = document.createElement('span');
    text.className = 'text';

    const lineText = lines[i];
    const lineAnnotations = session.annotations.filter(a =>
      a.startOffset < offset + lineText.length && a.endOffset > offset
    );

    if (lineAnnotations.length === 0) {
      text.textContent = lineText;
    } else {
      renderHighlightedLine(text, lineText, offset, lineAnnotations);
    }

    lineDiv.appendChild(gutter);
    lineDiv.appendChild(text);
    container.appendChild(lineDiv);

    offset += lineText.length + 1; // +1 for \n
  }
}

function renderHighlightedLine(textSpan, lineText, lineOffset, annotations) {
  const ranges = [];
  for (const ann of annotations) {
    const start = Math.max(0, ann.startOffset - lineOffset);
    const end = Math.min(lineText.length, ann.endOffset - lineOffset);
    if (start < end) {
      ranges.push({ start, end });
    }
  }
  ranges.sort((a, b) => a.start - b.start);

  let pos = 0;
  for (const r of ranges) {
    if (pos < r.start) {
      textSpan.appendChild(document.createTextNode(lineText.slice(pos, r.start)));
    }
    const mark = document.createElement('mark');
    mark.textContent = lineText.slice(r.start, r.end);
    textSpan.appendChild(mark);
    pos = r.end;
  }
  if (pos < lineText.length) {
    textSpan.appendChild(document.createTextNode(lineText.slice(pos)));
  }
}

function handleSelection() {
  const sel = window.getSelection();
  if (!sel || sel.isCollapsed) return;

  const range = sel.getRangeAt(0);
  const linesContainer = document.getElementById('lines');
  if (!linesContainer.contains(range.startContainer) || !linesContainer.contains(range.endContainer)) return;

  const startOffset = getCanonicalOffset(range.startContainer, range.startOffset);
  const endOffset = getCanonicalOffset(range.endContainer, range.endOffset);
  if (startOffset === null || endOffset === null || startOffset === endOffset) return;

  const [sOff, eOff] = startOffset <= endOffset ? [startOffset, endOffset] : [endOffset, startOffset];

  showToolbar(range.getBoundingClientRect(), sOff, eOff);
}

function getCanonicalOffset(node, offsetInNode) {
  const textSpan = node.nodeType === Node.TEXT_NODE
    ? node.parentElement.closest('.text')
    : node.closest('.text');
  if (!textSpan) return null;

  const lineDiv = textSpan.closest('.line');
  if (!lineDiv) return null;

  const lineStart = parseInt(lineDiv.dataset.start, 10);

  let accumulated = 0;
  const walker = document.createTreeWalker(textSpan, NodeFilter.SHOW_TEXT);
  let current = walker.nextNode();
  while (current) {
    if (current === node) {
      return lineStart + accumulated + offsetInNode;
    }
    accumulated += current.textContent.length;
    current = walker.nextNode();
  }

  // If node is the textSpan itself (e.g., selection at element boundary)
  return lineStart + accumulated;
}

let toolbarEl = null;

function showToolbar(rect, startOffset, endOffset) {
  hideToolbar();
  toolbarEl = document.createElement('div');
  toolbarEl.className = 'toolbar';
  toolbarEl.style.left = `${rect.left + rect.width / 2 - 50}px`;
  toolbarEl.style.top = `${rect.top - 40 + window.scrollY}px`;
  toolbarEl.style.position = 'absolute';

  const btn = document.createElement('button');
  btn.textContent = '💬 Comment';
  btn.addEventListener('click', () => {
    hideToolbar();
    window.getSelection().removeAllRanges();
    showAnnotationInput(startOffset, endOffset);
  });

  toolbarEl.appendChild(btn);
  document.body.appendChild(toolbarEl);

  const dismiss = (e) => {
    if (e.key === 'Escape' || (!toolbarEl.contains(e.target) && e.type === 'mousedown')) {
      hideToolbar();
      document.removeEventListener('keydown', dismiss);
      document.removeEventListener('mousedown', dismiss);
    }
  };
  setTimeout(() => {
    document.addEventListener('keydown', dismiss);
    document.addEventListener('mousedown', dismiss);
  }, 0);
}

function hideToolbar() {
  if (toolbarEl) {
    toolbarEl.remove();
    toolbarEl = null;
  }
}

function showAnnotationInput(startOffset, endOffset) {
  // Find the line div containing the endOffset and insert textarea after it
  const lines = document.querySelectorAll('#lines .line');
  let targetLine = null;
  for (const line of lines) {
    const start = parseInt(line.dataset.start, 10);
    const textSpan = line.querySelector('.text');
    const lineLen = textSpan ? textSpan.textContent.length : 0;
    if (endOffset <= start + lineLen + 1) {
      targetLine = line;
      break;
    }
  }
  if (!targetLine) targetLine = lines[lines.length - 1];

  const inputDiv = document.createElement('div');
  inputDiv.style.padding = '0 0 0 3em';
  const textarea = document.createElement('textarea');
  textarea.className = 'annotation-input';
  textarea.placeholder = 'Add your comment… (Enter to save, Esc to cancel)';
  inputDiv.appendChild(textarea);
  targetLine.after(inputDiv);
  textarea.focus();

  textarea.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      const text = textarea.value.trim();
      if (text) {
        session.annotations.push({
          type: 'comment',
          text,
          startOffset,
          endOffset,
        });
      }
      inputDiv.remove();
      renderLines();
    }
    if (e.key === 'Escape') {
      inputDiv.remove();
    }
  });
}

function renderExport() {
  const sorted = [...session.annotations].sort((a, b) => a.startOffset - b.startOffset);
  const summaryParts = [];

  for (const ann of sorted) {
    const startLine = offsetToLine(ann.startOffset);
    const endLine = offsetToLine(ann.endOffset - 1);
    const snippet = session.content.slice(ann.startOffset, ann.endOffset);
    const truncated = snippet.length > 80 ? snippet.slice(0, 80) + '…' : snippet;
    const lineRange = startLine === endLine ? `Line ${startLine}` : `Lines ${startLine}-${endLine}`;

    summaryParts.push(`### ${lineRange}\n> "${truncated}"\n**Comment:** ${ann.text}`);
  }

  const total = sorted.length;
  const footer = `---\n${total} annotation${total !== 1 ? 's' : ''} total`;
  const summaryText = `## Review Summary\n\n${summaryParts.join('\n\n')}\n\n${footer}`;

  app.innerHTML = `
    <div class="export">
      <div class="summary">${escapeHtml(summaryText)}</div>
      <button id="copy-btn">Copy to clipboard</button>
      <button id="back-btn">Back to review</button>
    </div>
  `;

  document.getElementById('copy-btn').addEventListener('click', async () => {
    try {
      await navigator.clipboard.writeText(summaryText);
      document.getElementById('copy-btn').textContent = 'Copied!';
    } catch {
      document.getElementById('copy-btn').textContent = 'Copy failed (need HTTPS)';
    }
  });

  document.getElementById('back-btn').addEventListener('click', renderEditor);
}

function offsetToLine(offset) {
  let line = 1;
  for (let i = 0; i < offset && i < session.content.length; i++) {
    if (session.content[i] === '\n') line++;
  }
  return line;
}

function escapeHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

// Boot
renderLanding();
