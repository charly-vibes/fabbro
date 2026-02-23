import { escapeHtml, offsetToLine } from './util.js';

export function render(container, session, { onDelete, onNoteClick }) {
  const count = session.annotations.length;
  container.innerHTML = '';

  const header = document.createElement('div');
  header.className = 'notes-header';
  header.textContent = `Notes (${count})`;
  container.appendChild(header);

  if (count === 0) {
    const empty = document.createElement('div');
    empty.className = 'notes-empty';
    empty.textContent = 'No annotations yet. Select text to add a comment.';
    container.appendChild(empty);
    return;
  }

  const list = document.createElement('div');
  list.className = 'notes-list';

  const sorted = [...session.annotations]
    .map((ann, i) => ({ ...ann, index: i }))
    .sort((a, b) => a.startOffset - b.startOffset);

  for (const ann of sorted) {
    const card = document.createElement('div');
    card.className = 'note-card';
    card.dataset.annotationIndex = ann.index;

    const badge = document.createElement('span');
    badge.className = `note-badge note-badge--${ann.type}`;
    badge.textContent = ann.type === 'suggest' ? 'Suggest' : 'Comment';

    const snippet = document.createElement('div');
    snippet.className = 'note-snippet';
    const raw = session.content.slice(ann.startOffset, ann.endOffset);
    const truncated = raw.length > 60 ? raw.slice(0, 60) + '…' : raw;
    snippet.textContent = truncated;

    const text = document.createElement('div');
    text.className = 'note-text';
    text.textContent = ann.text;

    const line = offsetToLine(session.content, ann.startOffset);
    const lineLabel = document.createElement('span');
    lineLabel.className = 'note-line';
    lineLabel.textContent = `L${line}`;

    const deleteBtn = document.createElement('button');
    deleteBtn.className = 'note-delete';
    deleteBtn.textContent = '✕';
    deleteBtn.title = 'Delete annotation';
    deleteBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      onDelete(ann.index);
    });

    const topRow = document.createElement('div');
    topRow.className = 'note-top';
    topRow.appendChild(badge);
    topRow.appendChild(lineLabel);
    topRow.appendChild(deleteBtn);

    card.appendChild(topRow);
    card.appendChild(snippet);
    card.appendChild(text);

    card.addEventListener('click', () => onNoteClick(ann));

    list.appendChild(card);
  }

  container.appendChild(list);
}

export function scrollToNote(container, annotationIndex) {
  const card = container.querySelector(`.note-card[data-annotation-index="${annotationIndex}"]`);
  if (card) {
    card.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    card.classList.add('note-card--active');
    setTimeout(() => card.classList.remove('note-card--active'), 1500);
  }
}
