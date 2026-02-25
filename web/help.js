let overlayEl = null;

const SHORTCUTS = [
  { section: 'Navigation' },
  { key: 'j / ↓', desc: 'Move down one line' },
  { key: 'k / ↑', desc: 'Move up one line' },
  { key: 'gg', desc: 'Go to first line' },
  { key: 'G', desc: 'Go to last line' },
  { key: 'Ctrl+d', desc: 'Half-page down' },
  { key: 'Ctrl+u', desc: 'Half-page up' },
  { section: 'Search' },
  { key: '/', desc: 'Open search' },
  { key: 'n', desc: 'Next match' },
  { key: 'N', desc: 'Previous match' },
  { key: 'Esc', desc: 'Clear search' },
  { section: 'Annotations' },
  { key: 'Select text', desc: 'Show annotation toolbar' },
  { key: 'Enter', desc: 'Save annotation' },
  { key: 'Shift+Enter', desc: 'Newline in annotation' },
  { key: 'Esc', desc: 'Cancel annotation' },
  { section: 'General' },
  { key: 'Space', desc: 'Open command palette' },
  { key: 'Ctrl+K', desc: 'Open command palette (global)' },
  { key: '?', desc: 'Toggle this help screen' },
];

export function toggle() {
  if (overlayEl) {
    hide();
  } else {
    show();
  }
}

export function show() {
  if (overlayEl) return;

  overlayEl = document.createElement('div');
  overlayEl.className = 'help-overlay';

  const panel = document.createElement('div');
  panel.className = 'help-panel';

  const title = document.createElement('h2');
  title.textContent = 'Keyboard Shortcuts';
  panel.appendChild(title);

  const dismiss = document.createElement('span');
  dismiss.className = 'help-dismiss';
  dismiss.textContent = 'Press ? or Esc to close';
  panel.appendChild(dismiss);

  const table = document.createElement('table');
  table.className = 'help-table';

  for (const item of SHORTCUTS) {
    const tr = document.createElement('tr');
    if (item.section) {
      const th = document.createElement('th');
      th.colSpan = 2;
      th.textContent = item.section;
      tr.appendChild(th);
    } else {
      const keyTd = document.createElement('td');
      keyTd.className = 'help-key';
      const kbd = document.createElement('kbd');
      kbd.textContent = item.key;
      keyTd.appendChild(kbd);
      const descTd = document.createElement('td');
      descTd.textContent = item.desc;
      tr.appendChild(keyTd);
      tr.appendChild(descTd);
    }
    table.appendChild(tr);
  }

  panel.appendChild(table);
  overlayEl.appendChild(panel);
  document.body.appendChild(overlayEl);

  overlayEl.addEventListener('click', (e) => {
    if (e.target === overlayEl) hide();
  });
}

export function hide() {
  if (overlayEl) {
    overlayEl.remove();
    overlayEl = null;
  }
}

export function isVisible() {
  return overlayEl !== null;
}
