// folder_size.js
#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const { promisify } = require('util');
const stat = promisify(fs.stat);
const readdir = promisify(fs.readdir);

const COLORS = {
    green: '\x1b[92m',
    red: '\x1b[91m',
    yellow: '\x1b[93m',
    blue: '\x1b[94m',
    reset: '\x1b[0m'
};

function colorize(text, color) {
    return COLORS[color] + text + COLORS.reset;
}

function humanReadable(size) {
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let s = size;
    let i = 0;
    while (s >= 1024 && i < units.length - 1) {
        s /= 1024;
        i++;
    }
    return `${s.toFixed(1)} ${units[i]}`;
}

async function getFolderSize(dir, recursive, depth, maxDepth, excludeHidden, verbose) {
    const result = {};
    let total = 0;

    try {
        const entries = await readdir(dir, { withFileTypes: true });
        for (const entry of entries) {
            if (excludeHidden && entry.name.startsWith('.')) continue;
            const full = path.join(dir, entry.name);
            if (entry.isFile()) {
                const st = await stat(full);
                total += st.size;
                if (verbose) console.log(`  ${full}: ${humanReadable(st.size)}`);
            } else if (entry.isDirectory() && recursive && (maxDepth === 0 || depth < maxDepth)) {
                const sub = await getFolderSize(full, recursive, depth+1, maxDepth, excludeHidden, verbose);
                for (const [k, v] of Object.entries(sub)) {
                    result[k] = v;
                }
                // Суммируем размер подпапки
                let subTotal = 0;
                for (const v of Object.values(sub)) subTotal += v;
                total += subTotal;
            }
        }
    } catch (err) {
        if (verbose) console.log(colorize(`Permission denied: ${dir}`, 'red'));
        return { [dir]: 0 };
    }
    result[dir] = total;
    return result;
}

async function main() {
    const args = require('minimist')(process.argv.slice(2), {
        string: ['p'],
        boolean: ['r', 'h', 's', 'exclude-hidden', 'v'],
        alias: { p: 'path', r: 'recursive', h: 'human-readable', s: 'sort', t: 'top', d: 'depth' },
        default: { p: '.', r: true, t: 0, d: 0 }
    });
    const root = path.resolve(args.p);
    const recursive = args.r;
    const human = args.h;
    const sortFlag = args.s;
    const top = args.t || 0;
    const maxDepth = args.d || 0;
    const excludeHidden = args['exclude-hidden'] || false;
    const verbose = args.v || false;

    try {
        const st = await stat(root);
        if (!st.isDirectory()) throw new Error('Not a directory');
    } catch (err) {
        console.log(colorize(`Error: ${err.message}`, 'red'));
        process.exit(1);
    }

    const sizeMap = await getFolderSize(root, recursive, 0, maxDepth, excludeHidden, verbose);
    let items = Object.entries(sizeMap);
    if (sortFlag) {
        items.sort((a, b) => b[1] - a[1]);
    }
    if (top > 0 && top < items.length) {
        items = items.slice(0, top);
    }

    for (const [p, size] of items) {
        let sizeStr = human ? humanReadable(size) : `${size} B`;
        let color = 'green';
        if (size > 1024*1024*1024) color = 'red';
        else if (size > 1024*1024) color = 'yellow';
        console.log(`${colorize(sizeStr.padStart(12), color)}  ${p}`);
    }
}

main().catch(err => {
    console.log(colorize(`Error: ${err.message}`, 'red'));
    process.exit(1);
});
