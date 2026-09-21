'use strict';

let es = null;

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

    es.onerror = () => {
        console.warn('[Default UI] SSE disconnected, reconnecting automatically');
    };
}

async function loadUser() {
    try {
        const res = await fetch('/api/auth/me');
        if (res.ok) {
            window.me = await res.json();
            renderUser(window.me);
            return;
        }
    } catch (e) {}
    renderUser(null);
}

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