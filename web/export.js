import { escapeHtml, offsetToLine } from './util.js';
import { serialize, ANNOTATION_TYPES } from './fem.js';

export function mount(container, session, { onBack }) {
  const sorted = [...session.annotations].sort((a, b) => a.startOffset - b.startOffset);
  const summaryParts = [];

  for (const ann of sorted) {
    const startLine = offsetToLine(session.content, ann.startOffset);
    const endLine = offsetToLine(session.content, ann.endOffset - 1);
    const snippet = session.content.slice(ann.startOffset, ann.endOffset);
    const truncated = snippet.length > 80 ? snippet.slice(0, 80) + '…' : snippet;
    const lineRange = startLine === endLine ? `Line ${startLine}` : `Lines ${startLine}-${endLine}`;

    summaryParts.push(`### ${lineRange}\n> "${truncated}"\n**Comment:** ${ann.text}`);
  }

  const total = sorted.length;
  const footer = `---\n${total} annotation${total !== 1 ? 's' : ''} total`;
  const summaryText = `## Review Summary\n\n${summaryParts.join('\n\n')}\n\n${footer}`;

  container.innerHTML = `
    <div class="export">
      <div class="summary">${escapeHtml(summaryText)}</div>
      <button id="copy-btn">Copy to clipboard</button>
      <button id="download-fem-btn">Download .fem</button>
      <button id="back-btn">Back to review</button>
    </div>
  `;

  document.getElementById('copy-btn').addEventListener('click', async () => {
    try {
      await navigator.clipboard.writeText(summaryText);
      document.getElementById('copy-btn').textContent = 'Copied!';
    } catch {
      document.getElementById('copy-btn').textContent = 'Copy failed (need HTTPS)';
    }
  });

  document.getElementById('download-fem-btn').addEventListener('click', () => {
    const femContent = buildFemContent(session.content, session.annotations);
    const output = serialize(femContent, {
      sessionId: session.id,
      createdAt: new Date().toISOString(),
      sourceFile: session.filename,
    });
    const blob = new Blob([output], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `review-${session.id}.fem`;
    a.click();
    URL.revokeObjectURL(url);
  });

  document.getElementById('back-btn').addEventListener('click', onBack);
}

function buildFemContent(content, annotations) {
  const sorted = [...annotations].sort((a, b) => b.startOffset - a.startOffset);
  let result = content;
  for (const ann of sorted) {
    const markerType = ANNOTATION_TYPES.find(t => t.name === (ann.type === 'suggest' ? 'change' : ann.type));
    if (!markerType) continue;
    const marker = `${markerType.open} ${ann.text} ${markerType.close}`;
    result = result.slice(0, ann.endOffset) + marker + result.slice(ann.endOffset);
  }
  return result;
}
