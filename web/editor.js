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
      onEdit: (index, newText) => {
        session.annotations[index].text = newText;
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

  let currentLine = -1;
  let viewerFocused = false;
  let lastGTime = 0;

  const viewerEl = container.querySelector('.viewer');

  function updateCurrentLine() {
    linesEl.querySelectorAll('.line--current').forEach(el => el.classList.remove('line--current'));
    if (currentLine < 0) return;
    const lines = linesEl.querySelectorAll('.line');
    if (currentLine >= lines.length) currentLine = lines.length - 1;
    if (lines[currentLine]) {
      lines[currentLine].classList.add('line--current');
      lines[currentLine].scrollIntoView({ block: 'nearest' });
    }
  }

  viewerEl.addEventListener('click', () => {
    viewerFocused = true;
    viewerEl.classList.add('viewer--focused');
  });

  document.addEventListener('mousedown', (e) => {
    if (!viewerEl.contains(e.target)) {
      viewerFocused = false;
      viewerEl.classList.remove('viewer--focused');
    }
  });

  document.addEventListener('keydown', (e) => {
    if (!viewerFocused) return;
    if (e.target.tagName === 'TEXTAREA' || e.target.tagName === 'INPUT') return;

    const lines = linesEl.querySelectorAll('.line');
    const totalLines = lines.length;
    if (totalLines === 0) return;

    const visibleLines = Math.floor(viewerEl.clientHeight / (lines[0]?.offsetHeight || 20));

    if (e.key === 'j' || e.key === 'ArrowDown') {
      e.preventDefault();
      currentLine = Math.min(currentLine + 1, totalLines - 1);
      if (currentLine < 0) currentLine = 0;
      updateCurrentLine();
    } else if (e.key === 'k' || e.key === 'ArrowUp') {
      e.preventDefault();
      if (currentLine < 0) currentLine = 0;
      else currentLine = Math.max(currentLine - 1, 0);
      updateCurrentLine();
    } else if (e.key === 'd' && e.ctrlKey) {
      e.preventDefault();
      currentLine = Math.min(currentLine + Math.floor(visibleLines / 2), totalLines - 1);
      if (currentLine < 0) currentLine = 0;
      updateCurrentLine();
    } else if (e.key === 'u' && e.ctrlKey) {
      e.preventDefault();
      if (currentLine < 0) currentLine = 0;
      else currentLine = Math.max(currentLine - Math.floor(visibleLines / 2), 0);
      updateCurrentLine();
    } else if (e.key === 'G') {
      e.preventDefault();
      currentLine = totalLines - 1;
      updateCurrentLine();
    } else if (e.key === 'g') {
      const now = Date.now();
      if (now - lastGTime < 500) {
        e.preventDefault();
        currentLine = 0;
        updateCurrentLine();
        lastGTime = 0;
      } else {
        lastGTime = now;
      }
    }
  });

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

  toolbar.show(range.getBoundingClientRect(), {
    onAnnotate: (type) => showAnnotationInput(session, sOff, eOff, type, refresh, onChanged),
  });
}

function showAnnotationInput(session, startOffset, endOffset, type, refresh, onChanged) {
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

  if (type === 'suggest') {
    const label = document.createElement('label');
    label.textContent = 'Replacement:';
    inputDiv.appendChild(label);
  }

  const textarea = document.createElement('textarea');
  textarea.className = 'annotation-input';
  textarea.placeholder = type === 'suggest'
    ? 'Enter replacement text… (Enter to save, Esc to cancel)'
    : 'Add your comment… (Enter to save, Esc to cancel)';
  inputDiv.appendChild(textarea);
  targetLine.after(inputDiv);
  textarea.focus();

  textarea.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      const text = textarea.value.trim();
      if (text) {
        session.annotations.push({
          type,
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
