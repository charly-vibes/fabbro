export function findMatches(content, query) {
  if (!query) return [];
  const lower = content.toLowerCase();
  const q = query.toLowerCase();
  const matches = [];
  let pos = 0;
  while (true) {
    const idx = lower.indexOf(q, pos);
    if (idx === -1) break;
    matches.push({ start: idx, end: idx + q.length });
    pos = idx + 1;
  }
  return matches;
}

export function mount(viewerEl, { onUpdate }) {
  const bar = document.createElement('div');
  bar.className = 'search-bar';
  bar.innerHTML = `
    <input type="text" class="search-input" placeholder="Search…">
    <span class="search-count"></span>
  `;
  bar.style.display = 'none';

  const input = bar.querySelector('.search-input');
  const countEl = bar.querySelector('.search-count');

  const state = { query: '', matches: [], current: -1, active: false };

  function updateCount() {
    if (state.matches.length === 0) {
      countEl.textContent = state.query ? '0/0' : '';
    } else {
      countEl.textContent = `${state.current + 1}/${state.matches.length}`;
    }
  }

  function open() {
    state.active = true;
    bar.style.display = '';
    input.value = state.query;
    input.focus();
    input.select();
  }

  function dismiss() {
    state.active = false;
    state.query = '';
    state.matches = [];
    state.current = -1;
    bar.style.display = 'none';
    updateCount();
    onUpdate(state);
  }

  function confirm() {
    state.active = false;
    bar.style.display = 'none';
  }

  function navigate(delta) {
    if (state.matches.length === 0) return;
    state.current = (state.current + delta + state.matches.length) % state.matches.length;
    updateCount();
    onUpdate(state);
  }

  input.addEventListener('input', () => {
    state.query = input.value;
    onUpdate(state);
    updateCount();
  });

  input.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      e.preventDefault();
      dismiss();
    } else if (e.key === 'Enter') {
      e.preventDefault();
      confirm();
    }
  });

  viewerEl.insertBefore(bar, viewerEl.firstChild);

  return { state, open, dismiss, confirm, navigate, updateCount };
}
