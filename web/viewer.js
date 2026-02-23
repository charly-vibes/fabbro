export function renderLines(container, content, annotations) {
  const lines = content.split('\n');
  let offset = 0;
  container.innerHTML = '';

  for (let i = 0; i < lines.length; i++) {
    const lineDiv = document.createElement('div');
    lineDiv.className = 'line';
    lineDiv.dataset.start = offset;

    const gutter = document.createElement('span');
    gutter.className = 'gutter';
    gutter.textContent = i + 1;

    const text = document.createElement('span');
    text.className = 'text';

    const lineText = lines[i];
    const lineAnnotations = annotations
      .map((a, idx) => ({ ...a, _index: idx }))
      .filter(a => a.startOffset < offset + lineText.length && a.endOffset > offset);

    if (lineAnnotations.length === 0) {
      text.textContent = lineText;
    } else {
      renderHighlightedLine(text, lineText, offset, lineAnnotations);
    }

    lineDiv.appendChild(gutter);
    lineDiv.appendChild(text);
    container.appendChild(lineDiv);

    offset += lineText.length + 1; // +1 for \n
  }
}

function renderHighlightedLine(textSpan, lineText, lineOffset, annotations) {
  const ranges = [];
  for (const ann of annotations) {
    const start = Math.max(0, ann.startOffset - lineOffset);
    const end = Math.min(lineText.length, ann.endOffset - lineOffset);
    if (start < end) {
      ranges.push({ start, end, index: ann._index });
    }
  }
  ranges.sort((a, b) => a.start - b.start);

  let pos = 0;
  for (const r of ranges) {
    if (pos < r.start) {
      textSpan.appendChild(document.createTextNode(lineText.slice(pos, r.start)));
    }
    const mark = document.createElement('mark');
    mark.dataset.annotationIndex = r.index;
    mark.textContent = lineText.slice(r.start, r.end);
    textSpan.appendChild(mark);
    pos = r.end;
  }
  if (pos < lineText.length) {
    textSpan.appendChild(document.createTextNode(lineText.slice(pos)));
  }
}

export function getCanonicalOffset(node, offsetInNode) {
  const textSpan = node.nodeType === Node.TEXT_NODE
    ? node.parentElement.closest('.text')
    : node.closest('.text');
  if (!textSpan) return null;

  const lineDiv = textSpan.closest('.line');
  if (!lineDiv) return null;

  const lineStart = parseInt(lineDiv.dataset.start, 10);

  let accumulated = 0;
  const walker = document.createTreeWalker(textSpan, NodeFilter.SHOW_TEXT);
  let current = walker.nextNode();
  while (current) {
    if (current === node) {
      return lineStart + accumulated + offsetInNode;
    }
    accumulated += current.textContent.length;
    current = walker.nextNode();
  }

  // If node is the textSpan itself (e.g., selection at element boundary)
  return lineStart + accumulated;
}
