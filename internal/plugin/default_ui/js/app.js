'use strict';

let es = null;
let offlineShown = false;
let offlineRetryTimer = null;

// ─── Offline detection ────────────────────────────────────────

function setOffline(off) {
    if (offlineShown === off) return;
    offlineShown = off;

    const banner = document.getElementById('offline-banner');
    if (banner) banner.classList.toggle('show', off);

    if (off) {
        if (!offlineRetryTimer) {
            offlineRetryTimer = setInterval(pingOnce, 3000);
        }
    } else {
        if (offlineRetryTimer) {
            clearInterval(offlineRetryTimer);
            offlineRetryTimer = null;
        }
    }
}

async function pingOnce() {
    try {
        const res = await fetch('/api/system/ping', { cache: 'no-store' });
        setOffline(!res.ok);
    } catch (e) {
        setOffline(true);
    }
}

// ─── SSE ──────────────────────────────────────────────────────

function startEvents() {
    es = new EventSource('/api/events');

    es.addEventListener('sync', (e) => {
        renderSync(JSON.parse(e.data));
    });

    es.addEventListener('card', (e) => {
        upsertCard(JSON.parse(e.data));
    });

    es.addEventListener('remove', (e) => {
        removeCard(JSON.parse(e.data).hash);
    });

    // Сервер отвечает — точно онлайн
    es.onopen = () => {
        setOffline(false);
    };

    // Ошибка SSE — проверяем реальное состояние один раз
    es.onerror = () => {
        if (offlineShown) return;
        pingOnce();
    };
}

// ─── User ─────────────────────────────────────────────────────

async function loadUser() {
    try {
        const res = await fetch('/api/auth/me');
        if (res.ok) {
            window.me = await res.json();
            renderUser(window.me);
            return;
        }
        renderUser(null);
    } catch (e) {
        setOffline(true);
        renderUser(null);
    }
}

// ─── Init ─────────────────────────────────────────────────────

async function init() {
    try {
        applyTranslations();
        await loadUser();
        bindModals();
        bindGrid();
        bindInfoModal();
        bindMenu();
        await loadPluginsMenu();
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