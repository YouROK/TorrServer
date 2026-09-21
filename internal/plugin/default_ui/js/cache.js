'use strict';

let cacheData = null;
let cacheMap = new Map();
let cacheZoom = 14;
let lastCacheSig = '';

const CACHE_ZOOM_MIN = 2;
const CACHE_ZOOM_MAX = 32;
const CACHE_ZOOM_STEP = 2;
const CACHE_GAP = 1;

const CACHE_PRIO = {
    5: { letter: 'A', color: '#ffffff' },
    4: { letter: 'N', color: '#ffc861' },
    3: { letter: 'R', color: '#ffc861' },
    2: { letter: 'H', color: '#b6bdc2' },
};

function setCacheData(d) {
    const sig = d ? JSON.stringify(d) : '';
    if (sig === lastCacheSig) return;
    lastCacheSig = sig;

    cacheData = d;
    cacheMap = new Map();
    if (d && Array.isArray(d.ids)) {
        for (let i = 0; i < d.ids.length; i++) {
            cacheMap.set(d.ids[i], { size: d.sizes[i], prio: d.priorities[i] });
        }
    }
    renderCache();
    renderCachePlaques();
}

function renderCachePlaques() {
    const box = document.getElementById('info-plaques-cache');
    if (!box) return;

    if (!cacheData) {
        box.innerHTML = '';
        return;
    }

    const rows = [
        [t('cache_size'),    fmtSize(cacheData.filled) + ' / ' + fmtSize(cacheData.capacity)],
        [t('pieces_count'),  String(cacheData.piece_count)],
        [t('pieces_length'), fmtSize(cacheData.piece_length)],
    ];

    box.innerHTML = rows.map((p) =>
        '<div class="plaque"><span class="plaque-label">' + p[0] + '</span>' +
        '<span class="plaque-value">' + esc(p[1]) + '</span></div>'
    ).join('');
}

function cacheZoomSet(px) {
    cacheZoom = Math.max(CACHE_ZOOM_MIN, Math.min(CACHE_ZOOM_MAX, px));
    const el = document.getElementById('cache-zoom-value');
    if (el) el.textContent = cacheZoom + ' px';
    renderCache();
}

function renderCache() {
    const wrap = document.getElementById('cache-canvas-wrap');
    const canvas = document.getElementById('cache-canvas');
    const emptyEl = document.getElementById('cache-empty');
    if (!wrap || !canvas) return;

    if (!cacheData) {
        canvas.style.display = 'none';
        if (emptyEl) emptyEl.style.display = '';
        return;
    }
    canvas.style.display = 'block';
    if (emptyEl) emptyEl.style.display = 'none';

    const T = cacheZoom;
    const G = CACHE_GAP;
    const cssW = wrap.clientWidth;
    if (cssW <= 0) return;

    const cols = Math.max(1, Math.floor((cssW + G) / (T + G)));
    const rows = Math.ceil(cacheData.piece_count / cols);
    const totalH = rows * (T + G) + G;

    if (canvas.width !== cssW) canvas.width = cssW;
    if (canvas.height !== totalH) canvas.height = totalH;
    canvas.style.width = cssW + 'px';
    canvas.style.height = totalH + 'px';

    const ctx = canvas.getContext('2d');
    ctx.fillStyle = '#14171a';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    const scrollTop = wrap.scrollTop;
    const viewH = wrap.clientHeight;
    const firstRow = Math.max(0, Math.floor(scrollTop / (T + G)));
    const lastRow = Math.min(rows - 1, Math.ceil((scrollTop + viewH) / (T + G)));

    for (let row = firstRow; row <= lastRow; row++) {
        const y = row * (T + G) + G;
        for (let col = 0; col < cols; col++) {
            const i = row * cols + col;
            if (i >= cacheData.piece_count) break;
            const x = col * (T + G) + G;
            const p = cacheMap.get(i);
            if (p) {
                const a = p.size / 255;
                ctx.fillStyle = 'rgba(78, 143, 137, ' + a.toFixed(3) + ')';
            } else {
                ctx.fillStyle = '#22282d';
            }
            ctx.fillRect(x, y, T, T);
        }
    }

    if (cacheData.readers) {
        ctx.lineWidth = 2;
        ctx.strokeStyle = '#f0a832';
        for (const r of cacheData.readers) {
            const i = r.reader;
            if (i < 0 || i >= cacheData.piece_count) continue;
            const row = Math.floor(i / cols);
            if (row < firstRow || row > lastRow) continue;
            const col = i % cols;
            const x = col * (T + G) + G;
            const y = row * (T + G) + G;
            ctx.strokeRect(x - 1, y - 1, T + 2, T + 2);
        }
    }

    if (T >= 8) {
        ctx.font = (T - 2) + 'px "JetBrains Mono", monospace';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        for (let row = firstRow; row <= lastRow; row++) {
            const yBase = row * (T + G) + G;
            for (let col = 0; col < cols; col++) {
                const i = row * cols + col;
                if (i >= cacheData.piece_count) break;
                const p = cacheMap.get(i);
                if (!p) continue;
                const pr = CACHE_PRIO[p.prio];
                if (!pr) continue;
                const x = col * (T + G) + G + T / 2;
                const y = yBase + T / 2;
                ctx.fillStyle = pr.color;
                ctx.fillText(pr.letter, x, y);
            }
        }
    }
}

function bindCache() {
    const zin = document.getElementById('cache-zoom-in');
    const zout = document.getElementById('cache-zoom-out');
    const zval = document.getElementById('cache-zoom-value');
    const wrap = document.getElementById('cache-canvas-wrap');
    if (zin) zin.addEventListener('click', () => cacheZoomSet(cacheZoom + CACHE_ZOOM_STEP));
    if (zout) zout.addEventListener('click', () => cacheZoomSet(cacheZoom - CACHE_ZOOM_STEP));
    if (zval) zval.textContent = cacheZoom + ' px';
    if (wrap) {
        wrap.addEventListener('scroll', () => {
            requestAnimationFrame(renderCache);
        });
    }
}