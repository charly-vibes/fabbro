import { escapeHtml } from './util.js';
import { renderLines, getCanonicalOffset } from './viewer.js';
import * as toolbar from './toolbar.js';
import * as notes from './notes.js';

export function mount(container, session, { onFinish, onChanged }) {
  container.innerHTML = `
    <div class="editor-header">
      <span>${escapeHtml(session.filename)}</span>
      <span id="save-indicator" class="save-indicator"></span>
      <button id="finish-btn">Finish review</button>
    </div>
    <div class="editor-layout">
      <div class="viewer">
        <div class="lines" id="lines"></div>
      </div>
      <div class="notes-panel" id="notes-panel"></div>
    </div>
  `;

  const linesEl = document.getElementById('lines');
  const notesPanel = document.getElementById('notes-panel');

  function refresh() {
    renderLines(linesEl, session.content, session.annotations);
    notes.render(notesPanel, session, {
      onDelete: (index) => {
        session.annotations.splice(index, 1);
        refresh();
        if (onChanged) onChanged();
      },
      onNoteClick: (ann) => {
        const mark = linesEl.querySelector(`mark[data-annotation-index="${ann.index}"]`);
        if (mark) {
          mark.scrollIntoView({ behavior: 'smooth', block: 'center' });
          mark.classList.add('mark--active');
          setTimeout(() => mark.classList.remove('mark--active'), 1500);
        }
      },
    });

    linesEl.querySelectorAll('mark[data-annotation-index]').forEach(mark => {
      mark.style.cursor = 'pointer';
      mark.addEventListener('click', () => {
        const idx = parseInt(mark.dataset.annotationIndex, 10);
        notes.scrollToNote(notesPanel, idx);
      });
    });
  }

  refresh();

  document.getElementById('finish-btn').addEventListener('click', onFinish);
  linesEl.addEventListener('mouseup', () => handleSelection(session, refresh, onChanged));
}

function handleSelection(session, refresh, onChanged) {
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
    showAnnotationInput(session, sOff, eOff, refresh, onChanged);
  });
}

function showAnnotationInput(session, startOffset, endOffset, refresh, onChanged) {
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
      refresh();
      if (onChanged) onChanged();
    }
    if (e.key === 'Escape') {
      inputDiv.remove();
    }
  });
}
