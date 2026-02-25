export const SAMPLE_CONTENT = `// greeting.js — A simple greeting module

function greet(name) {
  if (name == null) {
    return "Hello, stranger!";
  }

  var message = "Hello, " + name + "!";
  console.log(message);
  return message;
}

function greetAll(names) {
  var results = [];
  for (var i = 0; i <= names.length; i++) {
    results.push(greet(names[i]));
  }
  return results;
}

// TODO: Add a farewell function
module.exports = { greet, greetAll };`;

const STEPS = [
  {
    title: 'Welcome to fabbro!',
    message: 'This tutorial teaches you to review code with annotations. We\u2019ll walk through selecting text, adding comments, and exporting your review.',
  },
  {
    title: 'Step 1: Select text',
    message: 'Click and drag to select some text in the code \u2014 try selecting <code>name == null</code> on line 4.',
    target: '.viewer',
    waitFor: 'selection',
  },
  {
    title: 'Step 2: Pick an annotation type',
    message: 'A toolbar appeared! Click <strong>\uD83D\uDCAC Comment</strong> to add a review comment.',
    target: '.toolbar',
    waitFor: 'annotation-input',
  },
  {
    title: 'Step 3: Write your note',
    message: 'Type your comment (e.g. \u201CUse strict equality === instead\u201D) and press <strong>Enter</strong> to save.',
    waitFor: 'annotation-added',
  },
  {
    title: 'Step 4: Check the notes panel',
    message: 'Your annotation appears in the <strong>Notes</strong> panel on the right. You can click a note to jump to its highlighted text.',
    target: '.notes-panel',
  },
  {
    title: 'Step 5: Try a suggestion',
    message: 'Select <code>var message</code> on line 8, then click <strong>\u270F\uFE0F Suggest</strong> and type a replacement like <code>const message</code>.',
    target: '.viewer',
    waitFor: 'second-annotation',
  },
  {
    title: 'Step 6: Finish your review',
    message: 'Click <strong>Finish review</strong> in the top-right corner to see your review summary.',
    target: '#finish-btn',
    waitFor: 'export',
  },
  {
    title: 'Step 7: Export your work',
    message: 'Here\u2019s your review summary! You can <strong>copy to clipboard</strong> or <strong>download as .fem</strong> file.',
    target: '.export',
  },
  {
    title: 'Tutorial complete! \uD83C\uDF89',
    message: 'You know the basics: select text \u2192 annotate \u2192 export. Go review some real code!',
    final: true,
  },
];

let currentStep = -1;
let panelEl = null;
let active = false;
let observer = null;
let annotationCount = 0;
let selectionHandler = null;
let onCompleteFn = null;

export function start(onComplete) {
  active = true;
  currentStep = 0;
  annotationCount = 0;
  onCompleteFn = onComplete || null;
  showStep();
  setupObserver();
}

export function stop() {
  active = false;
  currentStep = -1;
  if (panelEl) {
    panelEl.remove();
    panelEl = null;
  }
  if (observer) {
    observer.disconnect();
    observer = null;
  }
  if (selectionHandler) {
    document.removeEventListener('mouseup', selectionHandler);
    selectionHandler = null;
  }
  clearHighlight();
}

export function isActive() {
  return active;
}

function setupObserver() {
  if (observer) observer.disconnect();
  observer = new MutationObserver(() => {
    if (!active || currentStep < 0 || currentStep >= STEPS.length) return;
    const step = STEPS[currentStep];
    if (!step || !step.waitFor) return;
    checkAutoAdvance(step);
  });
  observer.observe(document.body, { childList: true, subtree: true });
}

function checkAutoAdvance(step) {
  switch (step.waitFor) {
    case 'annotation-input': {
      if (document.querySelector('.annotation-input')) advance();
      break;
    }
    case 'annotation-added':
    case 'second-annotation': {
      const cards = document.querySelectorAll('.note-card');
      if (cards.length > annotationCount) {
        annotationCount = cards.length;
        advance();
      }
      break;
    }
    case 'export': {
      if (document.querySelector('.export')) advance();
      break;
    }
  }
}

function showStep() {
  if (currentStep < 0 || currentStep >= STEPS.length) {
    stop();
    return;
  }
  const step = STEPS[currentStep];
  renderPanel(step);
  highlightTarget(step.target);
  if (step.waitFor === 'selection') {
    setupSelectionListener();
  }
}

function setupSelectionListener() {
  if (selectionHandler) {
    document.removeEventListener('mouseup', selectionHandler);
  }
  selectionHandler = () => {
    if (!active) return;
    const step = STEPS[currentStep];
    if (!step || step.waitFor !== 'selection') return;
    const sel = window.getSelection();
    if (sel && !sel.isCollapsed) {
      document.removeEventListener('mouseup', selectionHandler);
      selectionHandler = null;
      setTimeout(() => advance(), 200);
    }
  };
  document.addEventListener('mouseup', selectionHandler);
}

function advance() {
  if (!active) return;
  currentStep++;
  showStep();
}

function renderPanel(step) {
  if (!panelEl) {
    panelEl = document.createElement('div');
    panelEl.className = 'tutorial-panel';
    document.body.appendChild(panelEl);
  }

  const stepNum = currentStep + 1;
  const total = STEPS.length;
  const dots = STEPS.map((_, i) => {
    const cls = i === currentStep ? 'tutorial-dot tutorial-dot--active'
      : i < currentStep ? 'tutorial-dot tutorial-dot--done'
      : 'tutorial-dot';
    return `<span class="${cls}"></span>`;
  }).join('');

  let actions = '';
  if (step.final) {
    actions = `
      <button class="tutorial-btn tutorial-btn--primary" data-action="done">Done</button>
      ${onCompleteFn ? '<button class="tutorial-btn" data-action="home">\u2190 Back to home</button>' : ''}
    `;
  } else if (!step.waitFor) {
    actions = `<button class="tutorial-btn tutorial-btn--primary" data-action="next">Next \u2192</button>`;
  } else {
    actions = `<span class="tutorial-hint">\uD83D\uDC46 Try it now</span>`;
  }

  panelEl.innerHTML = `
    <div class="tutorial-header">
      <div class="tutorial-dots">${dots}</div>
      <span class="tutorial-step-count">${stepNum}/${total}</span>
      <button class="tutorial-close" title="Exit tutorial">\u2715</button>
    </div>
    <div class="tutorial-body">
      <div class="tutorial-title">${step.title}</div>
      <div class="tutorial-message">${step.message}</div>
      <div class="tutorial-actions">${actions}</div>
    </div>
  `;

  panelEl.querySelector('.tutorial-close').addEventListener('click', stop);

  const nextBtn = panelEl.querySelector('[data-action="next"]');
  if (nextBtn) nextBtn.addEventListener('click', advance);

  const doneBtn = panelEl.querySelector('[data-action="done"]');
  if (doneBtn) doneBtn.addEventListener('click', stop);

  const homeBtn = panelEl.querySelector('[data-action="home"]');
  if (homeBtn) homeBtn.addEventListener('click', () => {
    stop();
    if (onCompleteFn) onCompleteFn();
  });
}

function highlightTarget(selector) {
  clearHighlight();
  if (!selector) return;
  requestAnimationFrame(() => {
    const target = document.querySelector(selector);
    if (target) target.classList.add('tutorial-highlight');
  });
}

function clearHighlight() {
  document.querySelectorAll('.tutorial-highlight').forEach(el =>
    el.classList.remove('tutorial-highlight')
  );
}
