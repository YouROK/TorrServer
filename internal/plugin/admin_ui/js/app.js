'use strict';

const t = (key, def) => i18n.t(key, def);

const SECTIONS = [
    { id: 'dashboard', rank: 50 },
    { id: 'users', rank: 50 },
    { id: 'logs', rank: 50 },
    { id: 'plugins', rank: 100 },
    { id: 'settings', rank: 100 },
];

let ME = null;
let refreshTimer = null;
let pluginDragId = null;
let offlineShown = false;

function buildLangSelect() {
    const sel = document.getElementById('lang-select');
    sel.innerHTML = '';

    for (const l of i18n.langs()) {
        const opt = document.createElement('option');
        opt.value = l.code;
        opt.textContent = l.name;
        sel.appendChild(opt);
    }

    sel.value = i18n.get();
    sel.addEventListener('change', () => {
        i18n.set(sel.value);
        refreshLangUI();
    });
}

function refreshLangUI() {
    i18n.apply();
    buildNav();
    route();
}

function openSidebar() {
    document.getElementById('sidebar').classList.add('open');
    document.getElementById('sidebar-overlay').classList.add('show');
}

function closeSidebar() {
    document.getElementById('sidebar').classList.remove('open');
    document.getElementById('sidebar-overlay').classList.remove('show');
}

async function api(path, opts = {}) {
    let res;
    try {
        res = await fetch('/api' + path, Object.assign({
            headers: { 'Content-Type': 'application/json' },
        }, opts));
    } catch (e) {
        showOffline();
        throw new Error('server offline');
    }

    if (res.status === 401) {
        location.href = '/login';
        throw new Error('unauthorized');
    }

    if (!res.ok) {
        let msg = 'request failed: ' + res.status;
        try { msg = (await res.json()).error || msg; } catch (e) {}
        throw new Error(msg);
    }

    hideOffline();
    return res.json();
}

function showOffline() {
    if (offlineShown) return;
    offlineShown = true;

    let banner = document.getElementById('offline-banner');
    if (!banner) {
        banner = document.createElement('div');
        banner.id = 'offline-banner';
        banner.className = 'offline-banner';
        banner.innerHTML = `
            <div class="offline-card">
                <h2>${t('server_offline_title')}</h2>
                <p>${t('server_offline_text')}</p>
                <p class="offline-hint">${t('retrying')}</p>
            </div>
        `;
        document.body.appendChild(banner);
    }
    banner.classList.add('show');
    startHealthCheck();
}

function hideOffline() {
    if (!offlineShown) return;
    offlineShown = false;
    const banner = document.getElementById('offline-banner');
    if (banner) banner.classList.remove('show');
}

let healthTimer = null;
function startHealthCheck() {
    if (healthTimer) return;
    healthTimer = setInterval(async () => {
        try {
            const res = await fetch('/api/system/ping', { cache: 'no-store' });
            if (res.ok) {
                clearInterval(healthTimer);
                healthTimer = null;
                hideOffline();
                route();
            }
        } catch (e) {}
    }, 3000);
}

function toast(msg) {
    const el = document.getElementById('toast');
    if (!el) return;
    el.textContent = msg;
    el.classList.add('show');
    setTimeout(() => el.classList.remove('show'), 2500);
}

const val = (id) => document.getElementById(id).value;
const chk = (id) => document.getElementById(id).checked;

function fmtBytes(b) {
    if (b > 1073741824) return (b / 1073741824).toFixed(1) + ' GiB';
    if (b > 1048576) return (b / 1048576).toFixed(1) + ' MiB';
    if (b > 1024) return (b / 1024).toFixed(1) + ' KiB';
    return b + ' B';
}

function fmtUptime(sec) {
    const d = Math.floor(sec / 86400);
    const h = Math.floor((sec % 86400) / 3600);
    const m = Math.floor((sec % 3600) / 60);
    if (d > 0) return d + 'd ' + h + 'h';
    if (h > 0) return h + 'm'.replace('h', h + 'h ' + m + 'm');
    return m + 'm ' + (sec % 60) + 's';
}

function badgeRank(rank) {
    if (rank >= 100) return '<span class="badge badge-owner">Owner</span>';
    if (rank >= 50) return '<span class="badge badge-admin">Admin</span>';
    return '<span class="badge badge-user">User</span>';
}

function wrapTable(tableHtml) {
    return '<div class="table-wrap">' + tableHtml + '</div>';
}

function showModal(opts) {
    return new Promise((resolve) => {
        const root = document.getElementById('modal-root');
        const box = document.createElement('div');
        box.className = 'modal-overlay';
        box.innerHTML = `
            <div class="modal">
                <h2>${opts.title || ''}</h2>
                <div class="modal-body">${opts.body || ''}</div>
                <div class="modal-actions"></div>
            </div>`;

        box.addEventListener('click', (e) => {
            if (e.target === box) { box.remove(); resolve(false); }
        });

        const actions = box.querySelector('.modal-actions');
        (opts.buttons || [{ label: t('ok'), value: true, primary: true }]).forEach((b) => {
            const btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'btn' + (b.primary ? '' : ' btn-secondary') + (b.danger ? ' btn-danger' : '');
            btn.textContent = b.label;
            btn.addEventListener('click', () => { box.remove(); resolve(b.value); });
            actions.appendChild(btn);
        });

        root.appendChild(box);
    });
}

const modalConfirm = (title, message) => showModal({
    title: title,
    body: '<p>' + message + '</p>',
    buttons: [
        { label: t('cancel'), value: false },
        { label: t('ok'), value: true, primary: true, danger: true },
    ],
});

const modalAlert = (title, bodyHtml) => showModal({ title: title, body: bodyHtml });

async function boot() {
    document.getElementById('hamburger').addEventListener('click', openSidebar);
    document.getElementById('sidebar-overlay').addEventListener('click', closeSidebar);

    try {
        ME = await api('/auth/me');
    } catch (e) {
        return;
    }

    document.getElementById('whoami').textContent = ME.username + ' · ' + rankName(ME.rank);
    document.getElementById('logout').addEventListener('click', async () => {
        await api('/auth/logout', { method: 'POST' });
        location.href = '/login';
    });

    buildLangSelect();
    i18n.apply();

    buildNav();
    window.addEventListener('hashchange', route);
    route();
}

function rankName(rank) {
    if (rank >= 100) return 'Owner';
    if (rank >= 50) return 'Admin';
    return 'User';
}

function buildNav() {
    const nav = document.getElementById('nav');
    nav.innerHTML = '';

    for (const sec of SECTIONS) {
        if (ME.rank < sec.rank) continue;
        const a = document.createElement('a');
        a.className = 'nav-item';
        a.href = '#/' + sec.id;
        a.dataset.section = sec.id;
        a.textContent = t(sec.id);
        a.addEventListener('click', closeSidebar);
        nav.appendChild(a);
    }
}

function route() {
    if (refreshTimer) {
        clearInterval(refreshTimer);
        refreshTimer = null;
    }

    let id = (location.hash || '#/dashboard').slice(2);
    const sec = SECTIONS.find((s) => s.id === id);
    if (!sec || ME.rank < sec.rank) id = 'dashboard';

    document.getElementById('page-title').textContent = t(id);
    document.querySelectorAll('.nav-item').forEach((el) => {
        el.classList.toggle('active', el.dataset.section === id);
    });

    const renderers = {
        dashboard: renderDashboard,
        users: renderUsers,
        plugins: renderPlugins,
        settings: renderSettings,
        logs: renderLogs,
    };
    renderers[id]();
}

async function renderDashboard() {
    const view = document.getElementById('view');
    view.innerHTML = '<div class="cards" id="cards"></div>';

    const load = async () => {
        try {
            const data = await api('/system/stats');
            const s = data.stats;
            document.getElementById('cards').innerHTML = `
                <div class="card"><div class="label">${t('active_torrents')}</div>
                    <div class="value amber">${s.active_sessions}</div></div>
                <div class="card"><div class="label">${t('active_readers')}</div>
                    <div class="value">${s.active_readers}</div></div>
                <div class="card"><div class="label">${t('memory')}</div>
                    <div class="value">${fmtBytes(s.mem_alloc_bytes)} / ${fmtBytes(s.mem_sys_bytes)}</div></div>
                <div class="card"><div class="label">${t('uptime')}</div>
                    <div class="value">${fmtUptime(data.uptime_sec)}</div></div>
                <div class="card"><div class="label">${t('goroutines')}</div>
                    <div class="value">${s.goroutines}</div></div>
            `;
        } catch (e) {
            toast(e.message);
        }
    };

    await load();
    refreshTimer = setInterval(load, 5000);
}

async function renderUsers() {
    const view = document.getElementById('view');
    view.innerHTML = `
        <div class="panel">
            <h2>${t('create_user')}</h2>
            <div class="row">
                <input class="input" id="nu-name" placeholder="${t('username')}">
                <input class="input" id="nu-pass" type="password" placeholder="${t('password')}">
                <select class="input" id="nu-rank" style="max-width:140px">
                    <option value="50">Admin</option>
                    <option value="10" selected>User</option>
                </select>
                <button class="btn" id="nu-create" type="button">${t('create_user')}</button>
            </div>
        </div>
        <div id="users-table-wrap"></div>
    `;

    document.getElementById('nu-create').addEventListener('click', async () => {
        try {
            await api('/users', {
                method: 'POST',
                body: JSON.stringify({
                    username: val('nu-name'),
                    password: val('nu-pass'),
                    rank: parseInt(val('nu-rank'), 10) || 10,
                }),
            });
            toast(t('done'));
            loadUsers();
        } catch (e) {
            toast(e.message);
        }
    });

    loadUsers();
}

async function loadUsers() {
    try {
        const data = await api('/users');
        const rows = (data.users || []).map((u) => `
            <tr>
                <td>${u.username}</td>
                <td>${badgeRank(u.rank)}</td>
                <td>${u.is_banned ? '<span class="badge badge-banned">' + t('banned') + '</span>' : t('active')}</td>
                <td class="actions-cell">
                    <button class="btn btn-secondary btn-sm" data-act="rank" data-id="${u.id}" data-rank="${u.rank}" type="button">${t('set_rank')}</button>
                    <button class="btn btn-secondary btn-sm" data-act="torrents" data-id="${u.id}" data-name="${u.username}" type="button">${t('view_torrents')}</button>
                    <button class="btn btn-secondary btn-sm" data-act="ban" data-id="${u.id}" data-banned="${u.is_banned}" type="button">
                        ${u.is_banned ? t('unban') : t('ban')}
                    </button>
                    <button class="btn btn-secondary btn-sm" data-act="token" data-id="${u.id}" type="button">${t('new_token')}</button>
                    <button class="btn btn-danger btn-sm" data-act="del" data-id="${u.id}" type="button">${t('delete')}</button>
                </td>
            </tr>
        `).join('');

        const html = `
            <table class="table" id="users-table">
                <tr><th>${t('username')}</th><th>${t('rank')}</th><th>${t('status')}</th><th>${t('actions')}</th></tr>
                ${rows}
            </table>
        `;
        document.getElementById('users-table-wrap').innerHTML = wrapTable(html);

        document.querySelectorAll('#users-table button').forEach((btn) => {
            btn.addEventListener('click', onUserAction);
        });
    } catch (e) {
        toast(e.message);
    }
}

async function onUserAction(e) {
    const id = e.target.dataset.id;
    const act = e.target.dataset.act;

    try {
        if (act === 'ban') {
            const banned = e.target.dataset.banned !== 'true';
            await api('/users/' + id, { method: 'PUT', body: JSON.stringify({ banned: banned }) });
            toast(t('done'));
            loadUsers();
        } else if (act === 'token') {
            const res = await api('/users/' + id + '/regenerate-token', { method: 'POST' });
            await modalAlert(t('token_title'), '<p class="mono">' + res.token + '</p>');
        } else if (act === 'del') {
            if (!(await modalConfirm(t('delete'), t('confirm_delete')))) return;
            await api('/users/' + id, { method: 'DELETE' });
            toast(t('done'));
            loadUsers();
        } else if (act === 'rank') {
            const currentRank = parseInt(e.target.dataset.rank, 10);
            await showRankModal(id, currentRank);
        } else if (act === 'torrents') {
            await showUserTorrents(id, e.target.dataset.name);
        }
    } catch (err) {
        toast(err.message);
    }
}

async function showRankModal(id, currentRank) {
    const body = `
        <p style="margin-bottom:12px">${t('select_rank_for_user')}</p>
        <select class="input" id="rank-select">
            <option value="10" ${currentRank < 50 ? 'selected' : ''}>${t('user_rank')}</option>
            <option value="50" ${currentRank >= 50 && currentRank < 100 ? 'selected' : ''}>${t('admin_rank')}</option>
        </select>
    `;

    const result = await showModal({
        title: t('set_rank'),
        body: body,
        buttons: [
            { label: t('cancel'), value: false },
            { label: t('save'), value: true, primary: true },
        ],
    });

    if (!result) return;

    const newRank = parseInt(document.getElementById('rank-select').value, 10);
    await api('/users/' + id + '/rank', { method: 'POST', body: JSON.stringify({ rank: newRank }) });
    toast(t('done'));
    loadUsers();
}

async function showUserTorrents(id, username) {
    try {
        const data = await api('/users/' + id + '/torrents');
        const torrents = data.torrents || [];
        let body = '';

        if (torrents.length === 0) {
            body = '<p>' + t('no_torrents') + '</p>';
        } else {
            body = '<div class="info-grid">' + torrents.map((tr) =>
                `<span>${tr.category || '—'}</span><b>${tr.title} <small style="color:var(--silo-text-faint)">${tr.torrent_hash.slice(0, 8)}</small></b>`
            ).join('') + '</div>';
        }

        await modalAlert(username + ' — ' + t('user_torrents'), body);
    } catch (e) {
        toast(e.message);
    }
}

async function renderPlugins() {
    const view = document.getElementById('view');
    view.innerHTML = `
        <div class="panel">
            <h2>${t('upload_plugin')}</h2>
            <div class="row">
                <label class="file-input">
                    <input type="file" id="pl-file" accept=".zip">
                    <span class="file-input-btn">${t('choose_file')}</span>
                    <span class="file-input-name" id="pl-file-name">${t('no_file_selected')}</span>
                </label>
                <button class="btn" id="pl-upload" type="button">${t('upload_plugin')}</button>
            </div>
            <p class="hint">${t('upload_hint')}</p>
        </div>
        <div class="panel">
            <h2>${t('install_from_url')}</h2>
            <div class="row">
                <input class="input" id="pl-url" placeholder="${t('plugin_url_placeholder')}">
                <button class="btn" id="pl-url-install" type="button">${t('install_from_url')}</button>
            </div>
            <p class="hint">${t('install_from_url_hint')}</p>
        </div>
        <p class="hint">${t('drag_hint')}</p>
        <div id="plugins-table-wrap"></div>
    `;

    const fileInput = document.getElementById('pl-file');
    const fileName = document.getElementById('pl-file-name');

    fileInput.addEventListener('change', () => {
        const f = fileInput.files[0];
        fileName.textContent = f ? f.name : t('no_file_selected');
    });

    document.getElementById('pl-upload').addEventListener('click', async () => {
        const file = fileInput.files[0];
        if (!file) {
            toast(t('select_file_first'));
            return;
        }

        await submitInstall(async (update) => {
            const fd = new FormData();
            fd.append('plugin', file);
            const url = '/api/plugins/upload' + (update ? '?update' : '');
            return fetch(url, { method: 'POST', body: fd });
        });

        fileInput.value = '';
        fileName.textContent = t('no_file_selected');
    });

    document.getElementById('pl-url-install').addEventListener('click', async () => {
        const urlInput = document.getElementById('pl-url');
        const url = urlInput.value.trim();
        if (!url) {
            toast(t('enter_url_first'));
            return;
        }

        const btn = document.getElementById('pl-url-install');
        btn.disabled = true;
        btn.textContent = '...';

        try {
            await submitInstall(async (update) => {
                const endpoint = '/api/plugins/install-url' + (update ? '?update' : '');
                return fetch(endpoint, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ url: url }),
                });
            });
            urlInput.value = '';
        } finally {
            btn.disabled = false;
            btn.textContent = t('install_from_url');
        }
    });

    loadPlugins();
}

async function submitInstall(doRequest) {
    let res;
    try {
        res = await doRequest(false);
    } catch (e) {
        toast(e.message);
        return;
    }

    if (res.status === 401) {
        location.href = '/login';
        return;
    }

    if (res.status === 409) {
        let body = {};
        try { body = await res.json(); } catch (e) {}

        if (body.is_builtin) {
            toast(t('plugin_builtin_cannot_replace'));
            return;
        }

        const ok = await showModal({
            title: t('plugin_exists_title'),
            body: '<p>' + t('plugin_exists_text')
                .replace('{id}', body.id)
                .replace('{version}', body.version) + '</p>',
            buttons: [
                { label: t('cancel'), value: false },
                { label: t('update'), value: true, primary: true },
            ],
        });
        if (!ok) return;

        res = await doRequest(true);
    }

    if (!res.ok) {
        let msg = 'request failed: ' + res.status;
        try { msg = (await res.json()).error || msg; } catch (e) {}
        toast(msg);
        return;
    }

    const data = await res.json();
    const count = (data.installed || []).length;
    const base = data.updated ? t('plugin_updated') : t('plugin_installed');
    toast(base + ' (' + count + ')');
    loadPlugins();
}

async function loadPlugins() {
    try {
        const data = await api('/plugins');
        const rows = (data.plugins || []).map((p) => {
            let themeCell;
            if (p.current_theme) {
                themeCell = '<span class="badge badge-owner">' + t('current') + '</span>';
            } else if (p.theme_ui) {
                themeCell = '<button class="btn btn-secondary btn-sm" data-act="theme" data-id="' + p.id + '" type="button">' + t('set_theme') + '</button>';
            } else {
                themeCell = '<span class="badge badge-user">' + t('no_theme') + '</span>';
            }

            const pageCell = (p.enabled && p.page)
                ? `<a class="btn btn-secondary btn-sm" href="/plugins/${p.id}/" target="_blank">${t('open')}</a>`
                : '';

            const statusBadge = p.enabled
                ? '<span class="badge badge-admin">' + t('active') + '</span>'
                : '<span class="badge badge-banned">' + t('disabled') + '</span>';

            const deleteBtn = p.builtin
                ? ''
                : `<button class="btn btn-danger btn-sm" data-act="del" data-id="${p.id}" type="button">${t('delete')}</button>`;

            return `
                <tr draggable="true" data-id="${p.id}">
                    <td class="drag-cell"><span class="drag-handle"></span></td>
                    <td>${p.id}${p.builtin ? ' <span class="badge badge-user">' + t('builtin') + '</span>' : ''}</td>
                    <td>${p.name}</td>
                    <td>${p.version}</td>
                    <td>${statusBadge}</td>
                    <td class="theme-cell">${themeCell}</td>
                    <td class="page-cell">${pageCell}</td>
                    <td class="actions-cell">
                        <button class="btn-icon" data-act="info" data-id="${p.id}" type="button" title="${t('info')}">i</button>
                        <button class="btn btn-secondary btn-sm" data-act="toggle" data-id="${p.id}" data-enabled="${p.enabled}" type="button">
                            ${p.enabled ? t('disable') : t('enable')}
                        </button>
                        ${deleteBtn}
                    </td>
                </tr>
            `;
        }).join('');

        const html = `
            <table class="table" id="plugins-table">
                <tr><th></th><th>${t('id')}</th><th>${t('name')}</th><th>${t('version')}</th><th>${t('status')}</th><th>${t('theme')}</th><th>${t('page')}</th><th>${t('actions')}</th></tr>
                ${rows}
            </table>
        `;
        document.getElementById('plugins-table-wrap').innerHTML = wrapTable(html);

        document.querySelectorAll('#plugins-table button').forEach((btn) => {
            btn.addEventListener('click', onPluginAction);
        });
        wireDrag(document.getElementById('plugins-table'));
    } catch (e) {
        toast(e.message);
    }
}

async function onPluginAction(e) {
    const id = e.target.dataset.id;
    const act = e.target.dataset.act;

    try {
        if (act === 'info') {
            await showPluginInfo(id);
        } else if (act === 'toggle') {
            const enabled = e.target.dataset.enabled !== 'true';
            await api('/plugins/' + id + '/enable', { method: 'POST', body: JSON.stringify({ enabled: enabled }) });
            toast(t('done'));
            loadPlugins();
        } else if (act === 'del') {
            if (!(await modalConfirm(t('delete'), t('confirm_delete')))) return;
            await api('/plugins/' + id, { method: 'DELETE' });
            toast(t('done'));
            loadPlugins();
        } else if (act === 'theme') {
            await setPluginAsTheme(id);
        }
    } catch (err) {
        toast(err.message);
    }
}

async function setPluginAsTheme(id) {
    try {
        const data = await api('/plugins');
        const ids = (data.plugins || []).map((p) => p.id);
        const idx = ids.indexOf(id);
        if (idx < 0) return;

        if (idx > 0) {
            ids.splice(idx, 1);
            ids.unshift(id);
        }

        await api('/plugins/order', { method: 'PUT', body: JSON.stringify({ order: ids }) });
        toast(t('done'));
        loadPlugins();
    } catch (e) {
        toast(e.message);
    }
}

async function showPluginInfo(id) {
    try {
        const m = await api('/plugins/' + id + '/info');
        const list = (arr) => arr.map((x) => '<li>' + x + '</li>').join('');

        let extra = '';
        if (m.events && m.events.length) {
            extra += '<h3>events</h3><ul class="info-list">' + list(m.events) + '</ul>';
        }
        if (m.routes && m.routes.length) {
            extra += '<h3>routes</h3><ul class="info-list">' + list(m.routes) + '</ul>';
        }

        await modalAlert(t('info'), `
            <div class="info-grid">
                <span>id</span><b>${m.id}</b>
                <span>name</span><b>${m.name}</b>
                <span>version</span><b>${m.version}</b>
                <span>author</span><b>${m.author || '—'}</b>
                <span>theme_ui</span><b>${m.theme_ui}</b>
                <span>entry</span><b>${m.entry || '—'}</b>
            </div>
            <p class="info-desc">${m.description || ''}</p>
            ${extra}
        `);
    } catch (e) {
        toast(e.message);
    }
}

function wireDrag(table) {
    table.querySelectorAll('tr[draggable]').forEach((row) => {
        row.addEventListener('dragstart', () => {
            pluginDragId = row.dataset.id;
            row.classList.add('dragging');
        });

        row.addEventListener('dragend', () => {
            pluginDragId = null;
            table.querySelectorAll('tr').forEach((r) => r.classList.remove('drag-over', 'dragging'));
        });

        row.addEventListener('dragover', (e) => {
            e.preventDefault();
            if (!pluginDragId || row.dataset.id === pluginDragId) return;
            table.querySelectorAll('tr').forEach((r) => r.classList.remove('drag-over'));
            row.classList.add('drag-over');
        });

        row.addEventListener('drop', async (e) => {
            e.preventDefault();
            const targetId = row.dataset.id;
            if (!pluginDragId || pluginDragId === targetId) return;

            const ids = Array.from(table.querySelectorAll('tr[draggable]')).map((r) => r.dataset.id);
            ids.splice(ids.indexOf(pluginDragId), 1);
            ids.splice(ids.indexOf(targetId), 0, pluginDragId);

            try {
                await api('/plugins/order', { method: 'PUT', body: JSON.stringify({ order: ids }) });
                toast(t('saved'));
                loadPlugins();
            } catch (err) {
                toast(err.message);
            }
        });
    });
}

async function renderSettings() {
    const view = document.getElementById('view');
    view.innerHTML = '<div class="panel"><p>Loading...</p></div>';

    try {
        const cfg = await api('/system/torrent-config');
        const st = cfg.storage || {};

        view.innerHTML = `
        <div class="panel">
            <h2>${t('torrent_engine')}</h2>

            <fieldset class="subblock">
                <legend>${t('engine')}</legend>
                <label class="field"><span>${t('listen_port')}</span>
                    <input class="input" id="te-port" type="number" value="${cfg.listen_port}"></label>
                <label class="field"><span>${t('download_rate')}</span>
                    <input class="input" id="te-down" type="number" value="${cfg.download_rate_kb}"></label>
                <label class="field"><span>${t('upload_rate')}</span>
                    <input class="input" id="te-up" type="number" value="${cfg.upload_rate_kb}"></label>
                <label class="field"><span>${t('preload_size')}</span>
                    <input class="input" id="te-preload" type="number" value="${Math.round((cfg.preload_size || 0) / 1048576)}" min="0"></label>
                <div class="checks">
                    <label><input type="checkbox" id="te-dht" ${cfg.disable_dht ? 'checked' : ''}> ${t('disable_dht')}</label>
                    <label><input type="checkbox" id="te-pex" ${cfg.disable_pex ? 'checked' : ''}> ${t('disable_pex')}</label>
                    <label><input type="checkbox" id="te-upnp" ${cfg.disable_upnp ? 'checked' : ''}> ${t('disable_upnp')}</label>
                    <label><input type="checkbox" id="te-utp" ${cfg.disable_utp ? 'checked' : ''}> ${t('disable_utp')}</label>
                    <label><input type="checkbox" id="te-tcp" ${cfg.disable_tcp ? 'checked' : ''}> ${t('disable_tcp')}</label>
                    <label><input type="checkbox" id="te-ipv6" ${cfg.enable_ipv6 ? 'checked' : ''}> ${t('enable_ipv6')}</label>
                </div>
            </fieldset>

            <fieldset class="subblock">
                <legend>${t('cache')}</legend>
                <label class="field"><span>${t('cache_size')}</span>
                    <input class="input" id="te-cap" type="number" value="${Math.round((st.capacity || 0) / 1048576)}"></label>
                <label class="field"><span>${t('connections_limit')}</span>
                    <input class="input" id="te-conn" type="number" value="${st.connections_limit || 0}"></label>
                <label class="field"><span>${t('read_ahead')}</span>
                    <input class="input" id="te-ahead" type="number" value="${st.reader_read_ahead || 95}"></label>
                <label class="field"><span>${t('disk_cache_path')}</span>
                    <input class="input" id="te-path" value="${st.torrents_save_path || ''}"></label>
                <div class="checks">
                    <label><input type="checkbox" id="te-disk" ${st.use_disk ? 'checked' : ''}> ${t('use_disk_instead_of_ram')}</label>
                    <label><input type="checkbox" id="te-rm" ${st.remove_cache_on_drop ? 'checked' : ''}> ${t('remove_cache_on_drop')}</label>
                </div>
            </fieldset>

            <button class="btn" id="te-save" type="button">${t('save')}</button>
        </div>`;

        document.getElementById('te-save').addEventListener('click', async () => {
            const payload = {
                listen_port: parseInt(val('te-port'), 10) || 0,
                download_rate_kb: parseInt(val('te-down'), 10) || 0,
                upload_rate_kb: parseInt(val('te-up'), 10) || 0,
                preload_size: (parseInt(val('te-preload'), 10) || 0) * 1048576,
                disable_dht: chk('te-dht'),
                disable_pex: chk('te-pex'),
                disable_upnp: chk('te-upnp'),
                disable_utp: chk('te-utp'),
                disable_tcp: chk('te-tcp'),
                enable_ipv6: chk('te-ipv6'),
                storage: {
                    capacity: (parseInt(val('te-cap'), 10) || 0) * 1048576,
                    use_disk: chk('te-disk'),
                    torrents_save_path: val('te-path'),
                    remove_cache_on_drop: chk('te-rm'),
                    connections_limit: parseInt(val('te-conn'), 10) || 0,
                    reader_read_ahead: parseInt(val('te-ahead'), 10) || 95,
                },
            };

            try {
                const res = await api('/system/torrent-config', { method: 'POST', body: JSON.stringify(payload) });
                toast(res.restart_required ? t('restart_note') : t('saved'));
            } catch (e) {
                toast(e.message);
            }
        });
    } catch (e) {
        view.innerHTML = '<div class="panel"><p>' + e.message + '</p></div>';
    }
}

async function renderLogs() {
    const view = document.getElementById('view');
    view.innerHTML = `
        <div class="row">
            <label><input type="checkbox" id="lg-auto" checked> ${t('auto_refresh')}</label>
        </div>
        <div class="logs-box" id="logs-box"></div>
    `;

    const load = async () => {
        try {
            const data = await api('/system/logs');
            const box = document.getElementById('logs-box');
            box.textContent = (data.logs || []).join('\n');
            box.scrollTop = box.scrollHeight;
        } catch (e) {
            toast(e.message);
        }
    };

    await load();
    refreshTimer = setInterval(() => {
        if (document.getElementById('lg-auto').checked) load();
    }, 3000);
}

boot();