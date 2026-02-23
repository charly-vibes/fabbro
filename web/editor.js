import { escapeHtml } from './util.js';
import { renderLines, getCanonicalOffset } from './viewer.js';
import * as toolbar from './toolbar.js';

export function mount(container, session, { onFinish }) {
  container.innerHTML = `
    <div class="editor-header">
      <span>${escapeHtml(session.filename)}</span>
      <button id="finish-btn">Finish review</button>
    </div>
    <div class="viewer">
      <div class="lines" id="lines"></div>
    </div>
  `;

  renderLines(document.getElementById('lines'), session.content, session.annotations);

  document.getElementById('finish-btn').addEventListener('click', onFinish);
  document.getElementById('lines').addEventListener('mouseup', () => handleSelection(session));
}

function handleSelection(session) {
  const sel = window.getSelection();
  if (!sel || sel.isCollapsed) return;

  const range = sel.getRangeAt(0);
  const linesContainer = document.getElementById('lines');
  if (!linesContainer.contains(range.startContainer) || !linesContainer.contains(range.endContainer)) return;

  const startOffset = getCanonicalOffset(range.startContainer, range.startOffset);
  const endOffset = getCanonicalOffset(range.endContainer, range.endOffset);
  if (startOffset === null || endOffset === null || startOffset === endOffset) return;

  const [sOff, eOff] = startOffset <= endOffset ? [startOffset, endOffset] : [endOffset, startOffset];

  toolbar.show(range.getBoundingClientRect(), () => {
    showAnnotationInput(session, sOff, eOff);
  });
}

function showAnnotationInput(session, startOffset, endOffset) {
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
      renderLines(document.getElementById('lines'), session.content, session.annotations);
    }
    if (e.key === 'Escape') {
      inputDiv.remove();
    }
  });
}
