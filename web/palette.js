let overlayEl = null;
let cursor = 0;
let filtered = [];
let allCommands = [];
let onExecute = null;

export function open(commands, { onSelect }) {
  if (overlayEl) { hide(); return; }
  allCommands = commands;
  onExecute = onSelect;
  cursor = 0;

  overlayEl = document.createElement('div');
  overlayEl.className = 'palette-overlay';

  const panel = document.createElement('div');
  panel.className = 'palette-panel';

  const input = document.createElement('input');
  input.type = 'text';
  input.className = 'palette-input';
  input.placeholder = 'Type a command…';
  panel.appendChild(input);

  const list = document.createElement('div');
  list.className = 'palette-list';
  panel.appendChild(list);

  overlayEl.appendChild(panel);
  document.body.appendChild(overlayEl);

  filtered = allCommands;
  renderList(list);
  input.focus();

  overlayEl.addEventListener('click', (e) => {
    if (e.target === overlayEl) hide();
  });

  input.addEventListener('input', () => {
    const q = input.value.toLowerCase();
    filtered = allCommands.filter(c =>
      c.label.toLowerCase().includes(q) || (c.key && c.key.toLowerCase().includes(q))
    );
    cursor = 0;
    renderList(list);
  });

  input.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      e.preventDefault();
      hide();
    } else if (e.key === 'ArrowDown' || (e.key === 'j' && e.ctrlKey)) {
      e.preventDefault();
      if (cursor < filtered.length - 1) cursor++;
      renderList(list);
    } else if (e.key === 'ArrowUp' || (e.key === 'k' && e.ctrlKey)) {
      e.preventDefault();
      if (cursor > 0) cursor--;
      renderList(list);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (filtered[cursor]) {
        const cmd = filtered[cursor];
        hide();
        if (onExecute) onExecute(cmd);
      }
    }
  });
}

function renderList(list) {
  list.innerHTML = '';
  for (let i = 0; i < filtered.length; i++) {
    const cmd = filtered[i];
    const row = document.createElement('div');
    row.className = 'palette-item' + (i === cursor ? ' palette-item--active' : '');

    const label = document.createElement('span');
    label.className = 'palette-label';
    label.textContent = cmd.label;

    row.appendChild(label);

    if (cmd.key) {
      const key = document.createElement('kbd');
      key.className = 'palette-kbd';
      key.textContent = cmd.key;
      row.appendChild(key);
    }

    row.addEventListener('click', () => {
      hide();
      if (onExecute) onExecute(cmd);
    });

    row.addEventListener('mouseenter', () => {
      cursor = i;
      renderList(list);
    });

    list.appendChild(row);
  }

  const active = list.querySelector('.palette-item--active');
  if (active) active.scrollIntoView({ block: 'nearest' });
}

export function hide() {
  if (overlayEl) {
    overlayEl.remove();
    overlayEl = null;
  }
  onExecute = null;
}

export function isVisible() {
  return overlayEl !== null;
}
