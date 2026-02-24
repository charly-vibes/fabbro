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
  toolbarEl.style.left = `${rect.left + rect.width / 2 - 75}px`;
  toolbarEl.style.top = `${rect.top - 40 + window.scrollY}px`;
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
    if (e.key === 'Escape' || (!toolbarEl.contains(e.target) && e.type === 'mousedown')) {
      hide();
      document.removeEventListener('keydown', dismiss);
      document.removeEventListener('mousedown', dismiss);
    }
  };
  setTimeout(() => {
    document.addEventListener('keydown', dismiss);
    document.addEventListener('mousedown', dismiss);
  }, 0);
}

export function hide() {
  if (toolbarEl) {
    toolbarEl.remove();
    toolbarEl = null;
  }
}
