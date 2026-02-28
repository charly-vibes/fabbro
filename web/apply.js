import { escapeHtml, offsetToLine } from './util.js';
import { renderLines } from './viewer.js';

export function mount(container, session, { onBack }) {
  const annotations = session.annotations;
  const sorted = [...annotations]
    .map((ann, i) => ({ ...ann, index: i }))
    .sort((a, b) => a.startOffset - b.startOffset);

  container.innerHTML = `
    <div class="editor-header">
      <span>${escapeHtml(session.filename)}</span>
      <span class="apply-badge">Read-only</span>
      <button id="apply-back-btn">← Back</button>
    </div>
    <div class="editor-layout">
      <div class="viewer">
        <div class="lines" id="lines"></div>
      </div>
      <div class="apply-panel" id="apply-panel"></div>
    </div>
  `;

  const linesEl = document.getElementById('lines');
  renderLines(linesEl, session.content, annotations, []);

  const panel = document.getElementById('apply-panel');
  renderSummary(panel, session, sorted);

  // Click annotations in text to scroll to note
  linesEl.querySelectorAll('mark[data-annotation-index]').forEach(mark => {
    mark.style.cursor = 'pointer';
    mark.addEventListener('click', () => {
      const idx = parseInt(mark.dataset.annotationIndex, 10);
      scrollToCard(panel, idx);
    });
  });

  // Click note cards to scroll to text
  panel.querySelectorAll('.apply-card').forEach(card => {
    card.addEventListener('click', () => {
      const idx = parseInt(card.dataset.annotationIndex, 10);
      const mark = linesEl.querySelector(`mark[data-annotation-index="${idx}"]`);
      if (mark) {
        mark.scrollIntoView({ behavior: 'smooth', block: 'center' });
        mark.classList.add('mark--active');
        setTimeout(() => mark.classList.remove('mark--active'), 1500);
      }
    });
  });

  document.getElementById('apply-back-btn').addEventListener('click', onBack);
}

function renderSummary(panel, session, sorted) {
  const header = document.createElement('div');
  header.className = 'notes-header';
  header.textContent = `Annotations (${sorted.length})`;
  panel.appendChild(header);

  if (sorted.length === 0) {
    const empty = document.createElement('div');
    empty.className = 'notes-empty';
    empty.textContent = 'No annotations found in this file.';
    panel.appendChild(empty);
    return;
  }

  const list = document.createElement('div');
  list.className = 'notes-list';

  for (const ann of sorted) {
    const card = document.createElement('div');
    card.className = 'apply-card';
    card.dataset.annotationIndex = ann.index;

    const startLine = offsetToLine(session.content, ann.startOffset);
    const endLine = offsetToLine(session.content, ann.endOffset - 1);
    const lineRange = startLine === endLine ? `L${startLine}` : `L${startLine}–${endLine}`;

    const badge = document.createElement('span');
    badge.className = `note-badge note-badge--${ann.type}`;
    badge.textContent = ann.type.charAt(0).toUpperCase() + ann.type.slice(1);

    const lineLabel = document.createElement('span');
    lineLabel.className = 'note-line';
    lineLabel.textContent = lineRange;

    const topRow = document.createElement('div');
    topRow.className = 'note-top';
    topRow.appendChild(badge);
    topRow.appendChild(lineLabel);

    const snippet = document.createElement('div');
    snippet.className = 'note-snippet';
    const raw = session.content.slice(ann.startOffset, ann.endOffset);
    const truncated = raw.length > 80 ? raw.slice(0, 80) + '…' : raw;
    snippet.textContent = truncated;

    const text = document.createElement('div');
    text.className = 'note-text';
    text.textContent = ann.text;

    card.appendChild(topRow);
    card.appendChild(snippet);
    card.appendChild(text);
    list.appendChild(card);
  }

  panel.appendChild(list);
}

function scrollToCard(panel, annotationIndex) {
  const card = panel.querySelector(`.apply-card[data-annotation-index="${annotationIndex}"]`);
  if (card) {
    card.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    card.classList.add('note-card--active');
    setTimeout(() => card.classList.remove('note-card--active'), 1500);
  }
}
