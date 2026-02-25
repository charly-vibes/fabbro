export function renderLines(container, content, annotations, searchMatches) {
  const lines = content.split('\n');
  let offset = 0;
  container.innerHTML = '';

  for (let i = 0; i < lines.length; i++) {
    const lineDiv = document.createElement('div');
    lineDiv.className = 'line';
    lineDiv.dataset.start = offset;

    const gutter = document.createElement('span');
    gutter.className = 'gutter';

    const hasAnnotation = annotations.some(
      a => a.startOffset < offset + lines[i].length && a.endOffset > offset
    );

    if (hasAnnotation) {
      const num = document.createElement('span');
      num.textContent = i + 1;
      const dot = document.createElement('span');
      dot.className = 'gutter-dot';
      dot.textContent = '●';
      gutter.appendChild(num);
      gutter.appendChild(dot);
    } else {
      gutter.textContent = i + 1;
    }

    const text = document.createElement('span');
    text.className = 'text';

    const lineText = lines[i];
    const lineAnnotations = annotations
      .map((a, idx) => ({ ...a, _index: idx }))
      .filter(a => a.startOffset < offset + lineText.length && a.endOffset > offset);

    const lineSearchMatches = searchMatches
      ? searchMatches.filter(m => m.start < offset + lineText.length && m.end > offset)
      : [];

    if (lineAnnotations.length === 0 && lineSearchMatches.length === 0) {
      text.textContent = lineText;
    } else {
      renderHighlightedLine(text, lineText, offset, lineAnnotations, lineSearchMatches);
    }

    lineDiv.appendChild(gutter);
    lineDiv.appendChild(text);
    container.appendChild(lineDiv);

    offset += lineText.length + 1; // +1 for \n
  }
}

function renderHighlightedLine(text, lineText, lineOffset, annotations, searchMatches) {
  // Collect boundary points to split the line into non-overlapping segments
  const boundaries = new Set([0, lineText.length]);

  const annRanges = [];
  for (const ann of annotations) {
    const s = Math.max(0, ann.startOffset - lineOffset);
    const e = Math.min(lineText.length, ann.endOffset - lineOffset);
    if (s < e) {
      boundaries.add(s);
      boundaries.add(e);
      annRanges.push({ start: s, end: e, index: ann._index });
    }
  }

  const sRanges = [];
  for (const m of searchMatches) {
    const s = Math.max(0, m.start - lineOffset);
    const e = Math.min(lineText.length, m.end - lineOffset);
    if (s < e) {
      boundaries.add(s);
      boundaries.add(e);
      sRanges.push({ start: s, end: e, matchIndex: m.index, current: m.current });
    }
  }

  const points = [...boundaries].sort((a, b) => a - b);

  for (let i = 0; i < points.length - 1; i++) {
    const segStart = points[i];
    const segEnd = points[i + 1];
    const segText = lineText.slice(segStart, segEnd);

    const ann = annRanges.find(r => r.start <= segStart && r.end >= segEnd);
    const search = sRanges.find(r => r.start <= segStart && r.end >= segEnd);

    if (!ann && !search) {
      text.appendChild(document.createTextNode(segText));
    } else if (ann && !search) {
      const mark = document.createElement('mark');
      mark.dataset.annotationIndex = ann.index;
      mark.textContent = segText;
      text.appendChild(mark);
    } else if (!ann && search) {
      const mark = document.createElement('mark');
      mark.className = 'search-match' + (search.current ? ' search-match--current' : '');
      mark.dataset.matchIndex = search.matchIndex;
      mark.textContent = segText;
      text.appendChild(mark);
    } else {
      // Both annotation and search match overlap
      const outer = document.createElement('mark');
      outer.dataset.annotationIndex = ann.index;
      const inner = document.createElement('mark');
      inner.className = 'search-match' + (search.current ? ' search-match--current' : '');
      inner.dataset.matchIndex = search.matchIndex;
      inner.textContent = segText;
      outer.appendChild(inner);
      text.appendChild(outer);
    }
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
