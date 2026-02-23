import { escapeHtml, offsetToLine } from './util.js';

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

  document.getElementById('back-btn').addEventListener('click', onBack);
}
