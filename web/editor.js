import { escapeHtml } from './util.js';
import { renderLines, getCanonicalOffset } from './viewer.js';
import * as toolbar from './toolbar.js';
import * as notes from './notes.js';
import * as search from './search.js';
import * as help from './help.js';
import * as palette from './palette.js';

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
  const viewerEl = container.querySelector('.viewer');

  let searchMatches = [];

  const searchCtrl = search.mount(viewerEl, {
    onUpdate: (state) => {
      if (state.query) {
        const raw = search.findMatches(session.content, state.query);
        state.matches = raw;
        if (state.current < 0 || state.current >= raw.length) {
          state.current = raw.length > 0 ? 0 : -1;
        }
        searchMatches = raw.map((m, i) => ({ ...m, index: i, current: i === state.current }));
        searchCtrl.updateCount();
      } else {
        searchMatches = [];
        state.matches = [];
        state.current = -1;
      }
      refresh();
      if (searchMatches.length > 0) {
        scrollToCurrentMatch();
      }
    },
  });

  function scrollToCurrentMatch() {
    const el = linesEl.querySelector('.search-match--current');
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }

  function refresh() {
    renderLines(linesEl, session.content, session.annotations, searchMatches);
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
    if (e.target.tagName === 'TEXTAREA' || e.target.tagName === 'INPUT') return;

    if (e.key === '?') {
      e.preventDefault();
      help.toggle();
      return;
    }
    if (e.key === 'Escape' && help.isVisible()) {
      e.preventDefault();
      help.hide();
      return;
    }
    if (e.key === 'Escape' && palette.isVisible()) {
      e.preventDefault();
      palette.hide();
      return;
    }
    if (e.key === 'k' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      openPalette();
      return;
    }

    if (!viewerFocused) return;

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
    } else if (e.key === ' ') {
      e.preventDefault();
      openPalette();
    } else if (e.key === '/') {
      e.preventDefault();
      searchCtrl.open();
    } else if (e.key === 'n') {
      if (searchMatches.length > 0) {
        e.preventDefault();
        searchCtrl.navigate(1);
      }
    } else if (e.key === 'N') {
      if (searchMatches.length > 0) {
        e.preventDefault();
        searchCtrl.navigate(-1);
      }
    }
  });

  function openPalette() {
    const commands = [
      { id: 'search', label: '🔍 Search', key: '/' },
      { id: 'help', label: '❓ Help', key: '?' },
      { id: 'finish', label: '🏁 Finish review', key: '' },
      { id: 'top', label: '⬆ Go to top', key: 'gg' },
      { id: 'bottom', label: '⬇ Go to bottom', key: 'G' },
      { id: 'ann:comment', label: '💬 Comment', key: 'select + toolbar' },
      { id: 'ann:suggest', label: '✏️ Suggest', key: 'select + toolbar' },
      { id: 'ann:delete', label: '🗑️ Delete', key: 'select + toolbar' },
      { id: 'ann:question', label: '❓ Question', key: 'select + toolbar' },
      { id: 'ann:expand', label: '💡 Expand', key: 'select + toolbar' },
      { id: 'ann:keep', label: '✅ Keep', key: 'select + toolbar' },
      { id: 'ann:unclear', label: '🔍 Unclear', key: 'select + toolbar' },
    ];

    palette.open(commands, {
      onSelect: (cmd) => {
        if (cmd.id === 'search') searchCtrl.open();
        else if (cmd.id === 'help') help.toggle();
        else if (cmd.id === 'finish') onFinish();
        else if (cmd.id === 'top') { currentLine = 0; updateCurrentLine(); }
        else if (cmd.id === 'bottom') {
          const lines = linesEl.querySelectorAll('.line');
          currentLine = lines.length - 1;
          updateCurrentLine();
        }
      },
    });
  }

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
