'use strict';

// ─── i18n helper ────────────────────────────────────────────────

function t(key) {
    return (window.i18n && window.i18n.t) ? window.i18n.t(key) : key;
}


// ─── Utilities ──────────────────────────────────────────────────

function showToast(message, type) {
    type = type || 'info';
    var root = document.getElementById('toast-root');
    var el   = document.createElement('div');
    el.className = 'toast toast-' + type;
    el.textContent = message;
    root.appendChild(el);
    requestAnimationFrame(function() { el.classList.add('toast-show'); });
    setTimeout(function() {
        el.classList.remove('toast-show');
        setTimeout(function() { el.remove(); }, 300);
    }, 3500);
}

function jsonEncode(obj) {
    try { return JSON.stringify(obj, null, 2); } catch(e) { return String(obj); }
}

function setOutput(elId, data, isError) {
    var el = document.getElementById(elId);
    if (!el) return;
    el.className = 'demo-output' + (isError ? ' is-error' : '');
    el.textContent = typeof data === 'object' ? jsonEncode(data) : String(data);
}

function modalShow(id) { document.getElementById(id).style.display = 'flex'; }
function modalHide(id) { document.getElementById(id).style.display = 'none'; }

function escapeHtml(str) {
    var d = document.createElement('div');
    d.textContent = str || '';
    return d.innerHTML;
}


// ─── Section 1: ts.info ─────────────────────────────────────────

function bindInfoSection() {
    document.getElementById('btn-info').addEventListener('click', function() {
        setOutput('out-info', t('loading'), false);
        fetch('api/info')
            .then(function(r) { return r.json(); })
            .then(function(d) { setOutput('out-info', d, false); })
            .catch(function(e) { setOutput('out-info', e.message, true); });
    });
}


// ─── Section 2: ts.web ──────────────────────────────────────────

function bindWebSection() {
    document.getElementById('btn-web-get').addEventListener('click', function() {
        setOutput('out-web', t('loading'), false);
        fetch('api/info')
            .then(function(r) { return r.json(); })
            .then(function(d) { setOutput('out-web', d, false); })
            .catch(function(e) { setOutput('out-web', e.message, true); });
    });

    document.getElementById('btn-web-post').addEventListener('click', function() {
        setOutput('out-web', t('loading'), false);
        fetch('api/echo', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ message: 'Hello from browser!', count: Math.floor(Math.random() * 100), nested: { foo: 'bar' } })
        })
            .then(function(r) { return r.json(); })
            .then(function(d) { setOutput('out-web', d, false); })
            .catch(function(e) { setOutput('out-web', e.message, true); });
    });
}


// ─── Section 3: ts.bus (polling, no SSE) ─────────────────────

var _lastEventCount = 0;

function bindBusSection() {
    // Emit button
    document.getElementById('btn-bus-emit').addEventListener('click', function() {
        var topic   = document.getElementById('emit-topic').value.trim();
        var payload = document.getElementById('emit-payload').value.trim();
        if (!topic) { showToast(t('bus_error_no_topic'), 'error'); return; }

        fetch('api/emit', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ topic: topic, payload: payload })
        })
            .then(function(r) { return r.json(); })
            .then(function(d) {
                if (d.emitted) {
                    showToast(t('bus_emitted_ok') + ': ' + topic, 'success');
                    // Immediately fetch events to show result
                    pollEvents();
                } else {
                    showToast(t('bus_emitted_fail'), 'error');
                }
            })
            .catch(function(e) { showToast(t('bus_emitted_fail') + ': ' + e.message, 'error'); });
    });
}

// Poll events every 2 seconds
function pollEvents() {
    fetch('api/events')
        .then(function(r) { return r.json(); })
        .then(function(d) {
            if (!d.events || d.events.length === 0) return;

            // Show only new events
            var newEvents = d.events.slice(_lastEventCount);
            _lastEventCount = d.events.length;

            newEvents.forEach(function(ev) {
                addBusEvent(ev.topic, ev.payload);
                // Show toast for plugin:example-plugin:message
                if (ev.topic === 'plugin:example-plugin:message') {
                    var msg = typeof ev.payload === 'object' ? JSON.stringify(ev.payload) : String(ev.payload);
                    showToast(t('bus_toast_received') + ': ' + msg, 'info');
                }
            });
        })
        .catch(function() { /* silently ignore polling errors */ });
}

function addBusEvent(topic, payload) {
    var outEl = document.getElementById('out-bus');
    var line = (new Date().toLocaleTimeString()) + '  ' + topic + '  →  ' +
        (typeof payload === 'object' ? JSON.stringify(payload) : payload);
    outEl.textContent = (outEl.textContent ? outEl.textContent + '\n' : '') + line;
    outEl.scrollTop = outEl.scrollHeight;
}

setInterval(pollEvents, 2000);


// ─── Section 4: ts.storage ─────────────────────────────────────

function bindStorageSection() {
    document.getElementById('btn-storage-get').addEventListener('click', function() {
        var key = document.getElementById('storage-key').value.trim();
        if (!key) { showToast(t('storage_error_no_key'), 'error'); return; }
        setOutput('out-storage', t('loading'), false);
        fetch('api/storage/get', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ key: key })
        })
            .then(function(r) { return r.json(); })
            .then(function(d) {
                setOutput('out-storage', d, false);
                if (d.value !== null) document.getElementById('storage-val').value = d.value;
            })
            .catch(function(e) { setOutput('out-storage', e.message, true); });
    });

    document.getElementById('btn-storage-set').addEventListener('click', function() {
        var key   = document.getElementById('storage-key').value.trim();
        var value = document.getElementById('storage-val').value;
        if (!key) { showToast(t('storage_error_no_key'), 'error'); return; }
        setOutput('out-storage', t('status_saving'), false);
        fetch('api/storage/set', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ key: key, value: value })
        })
            .then(function(r) { return r.json(); })
            .then(function(d) {
                setOutput('out-storage', d, false);
                showToast(t('storage_saved') + ': ' + key, 'success');
            })
            .catch(function(e) { setOutput('out-storage', e.message, true); });
    });

    document.getElementById('btn-storage-rem').addEventListener('click', function() {
        var key = document.getElementById('storage-key').value.trim();
        if (!key) { showToast(t('storage_error_no_key'), 'error'); return; }
        setOutput('out-storage', t('status_removing'), false);
        fetch('api/storage/rem', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ key: key })
        })
            .then(function(r) { return r.json(); })
            .then(function(d) {
                setOutput('out-storage', d, false);
                showToast(t('storage_removed') + ': ' + key, 'success');
            })
            .catch(function(e) { setOutput('out-storage', e.message, true); });
    });

    document.getElementById('btn-storage-clear').addEventListener('click', function() {
        setOutput('out-storage', t('status_clearing'), false);
        fetch('api/storage/clear', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        })
            .then(function(r) { return r.json(); })
            .then(function(d) {
                setOutput('out-storage', d, false);
                showToast(t('storage_cleared'), 'success');
            })
            .catch(function(e) { setOutput('out-storage', e.message, true); });
    });
}


// ─── Section 5: ts.http ─────────────────────────────────────────

function bindHTTPSection() {
    document.getElementById('btn-http-get').addEventListener('click', function() {
        setOutput('out-http', t('status_http_loading'), false);
        fetch('api/http-demo')
            .then(function(r) { return r.json(); })
            .then(function(d) { setOutput('out-http', d, false); })
            .catch(function(e) { setOutput('out-http', e.message, true); });
    });
}


// ─── Section 6: ts.torrent ─────────────────────────────────────

function bindTorrentSection() {
    document.getElementById('btn-torrent-list').addEventListener('click', function() {
        document.getElementById('torrents-list-content').innerHTML =
            '<p class="modal-loading">' + escapeHtml(t('loading')) + '</p>';
        modalShow('modal-torrents');
        fetch('api/torrent/list')
            .then(function(r) { return r.json(); })
            .then(function(d) {
                var container = document.getElementById('torrents-list-content');
                if (!d.torrents || d.torrents.length === 0) {
                    container.innerHTML = '<p class="modal-empty">' + escapeHtml(t('torrent_no_torrents')) + '</p>';
                } else {
                    container.innerHTML = d.torrents.map(function(torrent) {
                        return '<div class="modal-list-item">' +
                            '<div class="modal-list-title">' + escapeHtml(torrent.title || torrent.hash || '?') + '</div>' +
                            '<div class="modal-list-meta">hash: ' + escapeHtml(torrent.hash || '') + ' | status: ' + torrent.stat + '</div>' +
                            '</div>';
                    }).join('');
                }
            })
            .catch(function(e) {
                document.getElementById('torrents-list-content').innerHTML =
                    '<p class="modal-error">' + escapeHtml(e.message) + '</p>';
            });
    });

    document.getElementById('btn-torrent-get').addEventListener('click', function() {
        var hash = document.getElementById('torrent-hash').value.trim();
        if (!hash) { showToast(t('torrent_error_no_hash'), 'error'); return; }
        setOutput('out-torrent', t('loading'), false);
        fetch('api/torrent/get?hash=' + encodeURIComponent(hash))
            .then(function(r) { return r.json(); })
            .then(function(d) { setOutput('out-torrent', d.torrent || d, false); })
            .catch(function(e) { setOutput('out-torrent', e.message, true); });
    });

    document.getElementById('btn-torrent-add').addEventListener('click', function() {
        var input = document.getElementById('torrent-magnet').value.trim();
        if (!input) { showToast(t('torrent_error_no_input'), 'error'); return; }
        setOutput('out-torrent', t('status_adding'), false);
        fetch('api/torrent/add', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ input: input })
        })
            .then(function(r) { return r.json(); })
            .then(function(d) {
                setOutput('out-torrent', d, false);
                if (d.added) {
                    showToast(t('torrent_added') + ': ' + (d.status || ''), 'success');
                    document.getElementById('torrent-magnet').value = '';
                }
            })
            .catch(function(e) { setOutput('out-torrent', e.message, true); });
    });
}


// ─── Section 7: ts.users ────────────────────────────────────────

function bindUsersSection() {
    document.getElementById('btn-users-list').addEventListener('click', function() {
        document.getElementById('users-list-content').innerHTML =
            '<p class="modal-loading">' + escapeHtml(t('loading')) + '</p>';
        modalShow('modal-users');
        fetch('api/users/list')
            .then(function(r) { return r.json(); })
            .then(function(d) {
                var container = document.getElementById('users-list-content');
                if (!d.users || d.users.length === 0) {
                    container.innerHTML = '<p class="modal-empty">' + escapeHtml(t('users_no_users')) + '</p>';
                } else {
                    container.innerHTML = d.users.map(function(u) {
                        var rankName = u.rank === 100 ? 'owner' : u.rank === 50 ? 'admin' : u.rank === 10 ? 'user' : 'guest';
                        return '<div class="modal-list-item">' +
                            '<div class="modal-list-title">' + escapeHtml(u.username || u.id) + '</div>' +
                            '<div class="modal-list-meta">id: ' + escapeHtml(u.id) + ' | rank: ' + rankName + ' (' + u.rank + ')</div>' +
                            '</div>';
                    }).join('');
                }
            })
            .catch(function(e) {
                document.getElementById('users-list-content').innerHTML =
                    '<p class="modal-error">' + escapeHtml(e.message) + '</p>';
            });
    });

    document.getElementById('btn-users-get').addEventListener('click', function() {
        var uid = document.getElementById('user-id').value.trim();
        if (!uid) { showToast(t('users_error_no_id'), 'error'); return; }
        setOutput('out-users', t('loading'), false);
        fetch('api/users/get', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: uid })
        })
            .then(function(r) { return r.json(); })
            .then(function(d) { setOutput('out-users', d.user || d, false); })
            .catch(function(e) { setOutput('out-users', e.message, true); });
    });
}


// ─── Section 8: ts.i18n ─────────────────────────────────────────

function bindI18nSection() {
    document.getElementById('btn-i18n-demo').addEventListener('click', function() {
        var key   = 'showcase_builtin_translation';
        var value = (window.i18n && window.i18n.t) ? window.i18n.t(key) : key;
        var found = (value !== key);

        document.getElementById('out-i18n').innerHTML =
            '<div class="i18n-demo-result">' +
            '<div class="i18n-demo-key">key: <code>' + escapeHtml(key) + '</code></div>' +
            '<div class="i18n-demo-value">value: <code>' + escapeHtml(value) + '</code></div>' +
            '<div class="i18n-demo-note">' +
            escapeHtml(found ? t('i18n_dynamic_found') : t('i18n_dynamic_fallback')) +
            '</div>' +
            '</div>';
    });
}

// ─── Section 9: ts.vfs ──────────────────────────────────────────

function bindVFSSection() {
    var pathInput = document.getElementById('vfs-path');

    function callVFS(endpoint, extra) {
        var path = pathInput.value.trim() || '.';
        setOutput('out-vfs', t('loading'), false);
        fetch('api/vfs/' + endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(Object.assign({ path: path }, extra || {}))
        })
            .then(function(r) { return r.json(); })
            .then(function(d) { setOutput('out-vfs', d, !d.ok); })
            .catch(function(e) { setOutput('out-vfs', e.message, true); });
    }

    document.getElementById('btn-vfs-list').addEventListener('click', function() {
        callVFS('list');
    });
    document.getElementById('btn-vfs-stat').addEventListener('click', function() {
        callVFS('stat');
    });
    document.getElementById('btn-vfs-exists').addEventListener('click', function() {
        callVFS('exists');
    });
    document.getElementById('btn-vfs-read').addEventListener('click', function() {
        callVFS('read');
    });
}

// ─── Modals ─────────────────────────────────────────────────────

function bindModals() {
    document.getElementById('modal-torrents-close').addEventListener('click', function() { modalHide('modal-torrents'); });
    document.getElementById('modal-torrents').addEventListener('click', function(e) {
        if (e.target === this) modalHide('modal-torrents');
    });

    document.getElementById('modal-users-close').addEventListener('click', function() { modalHide('modal-users'); });
    document.getElementById('modal-users').addEventListener('click', function(e) {
        if (e.target === this) modalHide('modal-users');
    });

    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            modalHide('modal-torrents');
            modalHide('modal-users');
        }
    });
}


// ─── Init ───────────────────────────────────────────────────────

function init() {
    // Apply i18n if available
    if (window.i18n && window.i18n.apply) window.i18n.apply();

    bindInfoSection();
    bindWebSection();
    bindBusSection();
    bindStorageSection();
    bindHTTPSection();
    bindTorrentSection();
    bindUsersSection();
    bindI18nSection();
    bindVFSSection();
    bindModals();

    // Initial event poll
    pollEvents();
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}