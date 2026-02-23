let toolbarEl = null;

export function show(rect, onComment) {
  hide();
  toolbarEl = document.createElement('div');
  toolbarEl.className = 'toolbar';
  toolbarEl.style.left = `${rect.left + rect.width / 2 - 50}px`;
  toolbarEl.style.top = `${rect.top - 40 + window.scrollY}px`;
  toolbarEl.style.position = 'absolute';

  const btn = document.createElement('button');
  btn.textContent = '💬 Comment';
  btn.addEventListener('click', () => {
    hide();
    window.getSelection().removeAllRanges();
    onComment();
  });

  toolbarEl.appendChild(btn);
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
