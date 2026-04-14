#!/usr/bin/env node
// Extract JS seed blocks from the v7 prototype HTML into JSON files.
//
//   node scripts/extract-seed.js [--in prototype/DCGG_Intelligence_Platform_v7.html] [--out prototype/seed]

const fs = require('fs');
const path = require('path');
const vm = require('vm');

const NAMES = [
  'REGS',
  'PEOPLE',
  'ACTIVITIES',
  'STAKEHOLDERS',
  'CLIENT_PROFILES',
  'ACADEMIC_PUBLICATIONS',
  'PHOTOS',
];

function arg(flag, def) {
  const i = process.argv.indexOf(flag);
  return i >= 0 ? process.argv[i + 1] : def;
}

const IN  = arg('--in',  'prototype/DCGG_Intelligence_Platform_v7.html');
const OUT = arg('--out', 'prototype/seed');

const html = fs.readFileSync(IN, 'utf8');
fs.mkdirSync(OUT, { recursive: true });

function extractBalanced(src, from) {
  let i = from;
  while (i < src.length && src[i] !== '{' && src[i] !== '[') i++;
  if (i >= src.length) throw new Error('no opening brace/bracket');
  const open = src[i];
  let depth = 0;
  let inStr = null;
  let escape = false;
  let lineCmt = false;
  let blockCmt = false;
  for (let j = i; j < src.length; j++) {
    const c = src[j];
    const n = src[j + 1];
    if (lineCmt) { if (c === '\n') lineCmt = false; continue; }
    if (blockCmt) { if (c === '*' && n === '/') { blockCmt = false; j++; } continue; }
    if (inStr) {
      if (escape) { escape = false; continue; }
      if (c === '\\') { escape = true; continue; }
      if (c === inStr) inStr = null;
      continue;
    }
    if (c === '/' && n === '/') { lineCmt = true; j++; continue; }
    if (c === '/' && n === '*') { blockCmt = true; j++; continue; }
    if (c === '"' || c === "'" || c === '`') { inStr = c; continue; }
    if (c === '{' || c === '[') depth++;
    else if (c === '}' || c === ']') {
      depth--;
      if (depth === 0) return src.slice(i, j + 1);
    }
  }
  throw new Error('unbalanced');
}

const summary = {};

for (const name of NAMES) {
  const re = new RegExp(`(?:^|\\n)\\s*const\\s+${name}\\s*=\\s*`, 'g');
  const m = re.exec(html);
  if (!m) {
    console.warn(`[warn] ${name}: not found`);
    continue;
  }
  const start = m.index + m[0].length;
  let expr;
  try {
    expr = extractBalanced(html, start);
  } catch (e) {
    console.error(`[fail] ${name}: ${e.message}`);
    continue;
  }
  let value;
  try {
    value = vm.runInNewContext(`(${expr})`, {}, { timeout: 5000 });
  } catch (e) {
    console.error(`[fail] ${name}: eval error: ${e.message}`);
    continue;
  }
  const outPath = path.join(OUT, name.toLowerCase() + '.json');
  fs.writeFileSync(outPath, JSON.stringify(value, null, 2));
  const count = Array.isArray(value) ? value.length : Object.keys(value).length;
  summary[name] = { path: outPath, count, kind: Array.isArray(value) ? 'array' : 'object' };
  console.log(`[ok]   ${name.padEnd(22)} ${String(count).padStart(5)} ${summary[name].kind}  -> ${outPath}`);
}

fs.writeFileSync(path.join(OUT, 'index.json'), JSON.stringify(summary, null, 2));
