let toolbarEl = null;

const ANNOTATION_TYPES = [
  { type: 'comment', label: '💬 Comment', primary: true },
  { type: 'suggest', label: '✏️ Suggest', primary: true },
  { type: 'delete', label: '🗑️ Delete' },
  { type: 'question', label: '❓ Question' },
  { type: 'expand', label: '💡 Expand' },
  { type: 'keep', label: '✅ Keep' },
  { type: 'unclear', label: '🔍 Unclear' },
];

export function show(rect, { onAnnotate }) {
  hide();
  toolbarEl = document.createElement('div');
  toolbarEl.className = 'toolbar';
  const left = Math.max(8, Math.min(rect.left + rect.width / 2 - 75, window.innerWidth - 200));
  const topAbove = rect.top - 40 + window.scrollY;
  const topBelow = rect.bottom + 8 + window.scrollY;
  const top = topAbove < window.scrollY + 8 ? topBelow : topAbove;
  toolbarEl.style.left = `${left}px`;
  toolbarEl.style.top = `${top}px`;
  toolbarEl.style.position = 'absolute';

  const primaryTypes = ANNOTATION_TYPES.filter(t => t.primary);
  const moreTypes = ANNOTATION_TYPES.filter(t => !t.primary);

  for (const { type, label } of primaryTypes) {
    const btn = document.createElement('button');
    btn.textContent = label;
    btn.addEventListener('click', () => {
      hide();
      window.getSelection().removeAllRanges();
      onAnnotate(type);
    });
    toolbarEl.appendChild(btn);
  }

  const moreWrapper = document.createElement('div');
  moreWrapper.className = 'toolbar-more';

  const moreBtn = document.createElement('button');
  moreBtn.textContent = 'More ▾';
  moreBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    dropdown.classList.toggle('toolbar-dropdown--open');
  });

  const dropdown = document.createElement('div');
  dropdown.className = 'toolbar-dropdown';

  for (const { type, label } of moreTypes) {
    const item = document.createElement('button');
    item.className = 'toolbar-dropdown-item';
    item.textContent = label;
    item.addEventListener('click', () => {
      hide();
      window.getSelection().removeAllRanges();
      onAnnotate(type);
    });
    dropdown.appendChild(item);
  }

  moreWrapper.appendChild(moreBtn);
  moreWrapper.appendChild(dropdown);
  toolbarEl.appendChild(moreWrapper);
  document.body.appendChild(toolbarEl);

  const dismiss = (e) => {
    if (e.key === 'Escape' || (!toolbarEl.contains(e.target) && e.type === 'pointerdown')) {
      hide();
      document.removeEventListener('keydown', dismiss);
      document.removeEventListener('pointerdown', dismiss);
    }
  };
  setTimeout(() => {
    document.addEventListener('keydown', dismiss);
    document.addEventListener('pointerdown', dismiss);
  }, 0);
}

export function hide() {
  if (toolbarEl) {
    toolbarEl.remove();
    toolbarEl = null;
  }
}
