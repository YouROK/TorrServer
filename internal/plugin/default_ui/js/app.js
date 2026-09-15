'use strict';

const t = (key) => (typeof i18n !== 'undefined' && i18n.t) ? i18n.t(key) : key;

function fmtSize(bytes) {
    if (!bytes || bytes <= 0) return '—';
    if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GiB';
    if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MiB';
    if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KiB';
    return bytes + ' B';
}

function fmtSpeed(bps) {
    if (!bps || bps <= 0) return '—';
    if (bps >= 1048576) return (bps / 1048576).toFixed(1) + ' MB/s';
    if (bps >= 1024) return (bps / 1024).toFixed(0) + ' KB/s';
    return bps.toFixed(0) + ' B/s';
}

function fmtPeers(peers, seeders) {
    if ((!peers || peers <= 0) && (!seeders || seeders <= 0)) return '—';
    return (peers || 0) + '/' + (seeders || 0);
}

function badgeFor(stat) {
    if (stat === 0) return { cls: 'stat-0', key: 'stat_added' };
    if (stat === 1) return { cls: 'stat-1', key: 'stat_getting_info' };
    if (stat === 2) return { cls: 'stat-2', key: 'stat_preload' };
    if (stat === 3) return { cls: 'stat-3', key: 'stat_working' };
    return null;
}

function cardHTML(c) {
    if (!c) return '';
    const hash = c.hash || '';
    const title = c.title || hash || '—';
    const poster = c.poster || '';
    const badge = badgeFor(c.stat);

    const badgeHTML = badge
        ? '<span class="card-badge ' + badge.cls + '" data-i18n="' + badge.key + '">' + t(badge.key) + '</span>'
        : '';

    const posterHTML = poster
        ? '<img src="' + poster + '" alt="" loading="lazy">'
        : '<div class="placeholder" data-i18n="no_poster">' + t('no_poster') + '</div>';

    return '' +
        '<div class="card" data-hash="' + hash + '">' +
        '<div class="poster">' + posterHTML + badgeHTML + '</div>' +
        '<div class="title">' + title + '</div>' +
        '<div class="meta">' +
        '<span class="meta-item" title="' + t('size') + '">' +
        '<img src="img/ico-size.svg" alt="" onerror="this.style.display=\'none\'">' +
        '<em>' + fmtSize(c.size) + '</em>' +
        '</span>' +
        '<span class="meta-item" title="' + t('peers_seeds') + '">' +
        '<img src="img/ico-peers.svg" alt="" onerror="this.style.display=\'none\'">' +
        '<em>' + fmtPeers(c.peers, c.seeders) + '</em>' +
        '</span>' +
        '<span class="meta-item" title="' + t('download_speed') + '">' +
        '<img src="img/ico-speed.svg" alt="" onerror="this.style.display=\'none\'">' +
        '<em>' + fmtSpeed(c.download_speed) + '</em>' +
        '</span>' +
        '</div>' +
        '</div>';
}

function applyTranslations() {
    document.querySelectorAll('[data-i18n]').forEach((el) => {
        const key = el.getAttribute('data-i18n');
        el.textContent = t(key);
    });
}

function renderSync(cards) {
    const grid = document.getElementById('grid');
    if (!grid) return;
    if (!Array.isArray(cards) || cards.length === 0) {
        grid.innerHTML = '<p class="grid-msg" data-i18n="empty_library">' + t('empty_library') + '</p>';
        return;
    }
    grid.innerHTML = cards.map(cardHTML).join('');
    applyTranslations();
}

function upsertCard(c) {
    if (!c || !c.hash) return;
    const grid = document.getElementById('grid');
    if (!grid) return;
    const msg = grid.querySelector('.grid-msg');
    if (msg) msg.remove();

    const existing = grid.querySelector('.card[data-hash="' + c.hash + '"]');
    if (existing) {
        existing.outerHTML = cardHTML(c);
    } else {
        grid.insertAdjacentHTML('beforeend', cardHTML(c));
    }
    applyTranslations();
}

function removeCard(hash) {
    if (!hash) return;
    const grid = document.getElementById('grid');
    if (!grid) return;
    const existing = grid.querySelector('.card[data-hash="' + hash + '"]');
    if (existing) existing.remove();
    if (!grid.querySelector('.card')) {
        grid.innerHTML = '<p class="grid-msg" data-i18n="empty_library">' + t('empty_library') + '</p>';
    }
}

let es = null;
let errorCount = 0;

function startEvents() {
    console.log('[Default UI] Starting SSE connection...');
    es = new EventSource('/api/events');

    es.addEventListener('sync', (e) => {
        console.log('[Default UI] SSE sync received');
        errorCount = 0;
        renderSync(JSON.parse(e.data));
    });

    es.addEventListener('card', (e) => {
        const card = JSON.parse(e.data);
        console.log('[Default UI] SSE card update:', card.hash.slice(0, 8), 'stat=' + card.stat);
        upsertCard(card);
    });

    es.addEventListener('remove', (e) => {
        const hash = JSON.parse(e.data).hash;
        console.log('[Default UI] SSE remove:', hash.slice(0, 8));
        removeCard(hash);
    });

    es.onopen = () => {
        console.log('[Default UI] SSE connected');
        errorCount = 0;
    };

    es.onerror = (e) => {
        errorCount++;
        console.warn('[Default UI] SSE error #' + errorCount + ', readyState=' + es.readyState);
        if (errorCount > 10) {
            console.error('[Default UI] SSE failed after 10 errors, giving up');
        }
    };
}

async function loadUser() {
    try {
        const res = await fetch('/api/auth/me');
        if (res.ok) {
            const user = await res.json();
            document.getElementById('username').textContent = user.username;
            return;
        }
    } catch (e) {}
    document.getElementById('username').textContent = t('guest');
}

async function init() {
    try {
        applyTranslations();
        await loadUser();
        startEvents();
    } catch (e) {
        console.error('[Default UI] init failed:', e);
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}