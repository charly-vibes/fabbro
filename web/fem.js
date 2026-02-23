// FEM (Fragmented Editing Markers) parser - ES module port of Go implementation.
// No dependencies, no build step.

export const ANNOTATION_TYPES = [
  { name: 'comment',  open: '{>>', close: '<<}' },
  { name: 'delete',   open: '{--', close: '--}' },
  { name: 'question', open: '{??', close: '??}' },
  { name: 'expand',   open: '{!!', close: '!!}' },
  { name: 'keep',     open: '{==', close: '==}' },
  { name: 'unclear',  open: '{~~', close: '~~}' },
  { name: 'change',   open: '{++', close: '++}' },
];

function escapeRegex(s) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

const patterns = {};
for (const at of ANNOTATION_TYPES) {
  patterns[at.name] = new RegExp(
    escapeRegex(at.open) + '\\s*(.*?)\\s*' + escapeRegex(at.close),
    'g'
  );
}

const openingMarkers = ANNOTATION_TYPES.map(at => at.open);

function containsNestedMarker(text) {
  for (const marker of openingMarkers) {
    if (text.includes(marker)) {
      return true;
    }
  }
  return false;
}

export function parse(content) {
  const lines = content.split('\n');
  const annotations = [];
  const cleanLines = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    let cleanLine = line;

    for (const at of ANNOTATION_TYPES) {
      const pattern = patterns[at.name];
      // Reset lastIndex and collect all matches from the ORIGINAL line
      pattern.lastIndex = 0;
      const matches = [...line.matchAll(pattern)];

      for (const match of matches) {
        if (match.length >= 2) {
          if (containsNestedMarker(match[1])) {
            continue;
          }
          // Remove ALL occurrences of this pattern from cleanLine
          cleanLine = cleanLine.replace(pattern, '');
          const lineNum = i + 1;
          annotations.push({
            type: at.name,
            text: match[1],
            startLine: lineNum,
            endLine: lineNum,
          });
        }
      }
    }

    cleanLines.push(cleanLine);
  }

  const cleanContent = cleanLines.join('\n');
  return { annotations, cleanContent };
}

function quoteYAMLString(s) {
  return "'" + s.replace(/'/g, "''") + "'";
}

export function serialize(content, metadata) {
  let sourceFileLine = '';
  if (metadata.sourceFile) {
    sourceFileLine = `source_file: ${quoteYAMLString(metadata.sourceFile)}\n`;
  }

  return `---
session_id: ${metadata.sessionId}
created_at: ${metadata.createdAt}
${sourceFileLine}---

${content}`;
}
