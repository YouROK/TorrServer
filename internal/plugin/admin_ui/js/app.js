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
let CATALOG = null;
let SILO_VERSION = '';
let refreshTimer = null;
let pluginDragId = null;
let offlineShown = false;

function esc(s) {
    return String(s == null ? '' : s)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

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
        if (e && e.name === 'AbortError') throw new Error('request aborted');
        showOffline();
        throw new Error('server offline');
    }

    if (res.status === 401) {
        // Сессия недействительна - уходим на логин, сняв баннер недоступности.
        hideOffline();
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

    const stop = () => {
        clearInterval(healthTimer);
        healthTimer = null;
    };

    healthTimer = setInterval(async () => {
        try {
            const res = await fetch('/api/system/ping', { cache: 'no-store' });

            // Сессия недействительна: сервер жив, поэтому баннер недоступности
            // снимаем и уходим на логин вместо бесконечного «переподключения».
            if (res.status === 401) {
                stop();
                hideOffline();
                location.href = '/login';
                return;
            }

            if (res.ok) {
                stop();
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

// showModal показывает модальное окно и возвращает значение нажатой кнопки.
// У кнопки с collect: true вызывается opts.getValue(box, value) до удаления окна
// из DOM: после remove() поля формы уже недоступны. Если getValue вернул
// undefined, окно остаётся открытым - так показывают ошибку ввода на месте.
// Остальные кнопки (отмена) возвращают своё value, не читая форму.
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
            btn.addEventListener('click', () => {
                if (b.collect && opts.getValue) {
                    const value = opts.getValue(box, b.value);
                    if (value === undefined) return;
                    box.remove();
                    resolve(value);
                    return;
                }
                box.remove();
                resolve(b.value);
            });
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
    try {
        const v = await api('/system/version');
        SILO_VERSION = v.version || '';
    } catch (e) {}

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
        const rows = (data.users || []).map((u) => {
            // У Owner пароль задаётся в config.yaml, менять его из админки нельзя.
            const passwordBtn = u.rank >= 100
                ? ''
                : `<button class="btn btn-secondary btn-sm" data-act="password" data-id="${u.id}" data-name="${u.username}" type="button">${t('change_password')}</button>`;

            return `
            <tr>
                <td>${u.username}</td>
                <td>${badgeRank(u.rank)}</td>
                <td>${u.is_banned ? '<span class="badge badge-banned">' + t('banned') + '</span>' : t('active')}</td>
                <td class="actions-cell">
                    <button class="btn btn-secondary btn-sm" data-act="rank" data-id="${u.id}" data-rank="${u.rank}" type="button">${t('set_rank')}</button>
                    ${passwordBtn}
                    <button class="btn btn-secondary btn-sm" data-act="torrents" data-id="${u.id}" data-name="${u.username}" type="button">${t('view_torrents')}</button>
                    <button class="btn btn-secondary btn-sm" data-act="ban" data-id="${u.id}" data-banned="${u.is_banned}" type="button">
                        ${u.is_banned ? t('unban') : t('ban')}
                    </button>
                    <button class="btn btn-secondary btn-sm" data-act="token" data-id="${u.id}" type="button">${t('new_token')}</button>
                    <button class="btn btn-danger btn-sm" data-act="del" data-id="${u.id}" type="button">${t('delete')}</button>
                </td>
            </tr>
        `;
        }).join('');

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
        } else if (act === 'password') {
            await showPasswordModal(id, e.target.dataset.name);
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

    // Значение читаем внутри getValue, пока окно ещё в DOM.
    const result = await showModal({
        title: t('set_rank'),
        body: body,
        getValue: (box, value) => {
            if (!value) return null;
            const el = box.querySelector('#rank-select');
            return el ? parseInt(el.value, 10) : null;
        },
        buttons: [
            { label: t('cancel'), value: false },
            { label: t('save'), value: true, primary: true, collect: true },
        ],
    });

    if (!result) return;

    await api('/users/' + id + '/rank', { method: 'POST', body: JSON.stringify({ rank: result }) });
    toast(t('done'));
    loadUsers();
}

// showPasswordModal запрашивает новый пароль пользователя и сохраняет его.
async function showPasswordModal(id, username) {
    const body = `
        <p style="margin-bottom:12px">${t('change_password_for')} <b>${esc(username)}</b></p>
        <label class="field"><span>${t('new_password')}</span>
            <input class="input" id="pw-new" type="password" autocomplete="new-password"></label>
        <label class="field"><span>${t('repeat_password')}</span>
            <input class="input" id="pw-repeat" type="password" autocomplete="new-password"></label>
        <p class="hint" id="pw-hint"></p>
    `;

    const result = await showModal({
        title: t('change_password'),
        body: body,
        // Проверяем ввод до закрытия окна. При ошибке окно остаётся открытым,
        // а текст показывается рядом с полями - введённое не теряется.
        getValue: (box) => {
            const hint = box.querySelector('#pw-hint');
            const pw = box.querySelector('#pw-new').value;
            const repeat = box.querySelector('#pw-repeat').value;

            if (!pw) {
                hint.textContent = t('password_empty');
                return undefined;
            }
            if (pw !== repeat) {
                hint.textContent = t('passwords_differ');
                return undefined;
            }

            hint.textContent = '';
            return { password: pw };
        },
        buttons: [
            { label: t('cancel'), value: null },
            { label: t('save'), value: true, primary: true, collect: true },
        ],
    });

    if (!result) return;

    try {
        await api('/users/' + id, { method: 'PUT', body: JSON.stringify({ password: result.password }) });
        toast(t('done'));
    } catch (err) {
        toast(err.message);
    }
}

// fmtSpeed форматирует байты в секунду
function fmtSpeed(bps) {
    if (!bps || bps <= 0) return '-';
    if (bps >= 1048576) return (bps / 1048576).toFixed(1) + ' MiB/s';
    if (bps >= 1024) return (bps / 1024).toFixed(0) + ' KiB/s';
    return Math.round(bps) + ' B/s';
}

// fmtDuration форматирует секунды в компактный вид
function fmtDuration(sec) {
    sec = Math.max(0, Math.floor(sec || 0));
    const h = Math.floor(sec / 3600);
    const m = Math.floor((sec % 3600) / 60);
    const s = sec % 60;
    if (h > 0) return h + 'h ' + m + 'm';
    if (m > 0) return m + 'm ' + s + 's';
    return s + 's';
}

// shortText сокращает строку посередине: начало и конец обычно информативнее
// середины. Полное значение показывается в подсказке при наведении.
function shortText(s, max) {
    const str = String(s == null ? '' : s);
    if (str.length <= max) return str;

    const keep = max - 3;
    const head = Math.ceil(keep / 2);
    const tail = Math.floor(keep / 2);

    return str.slice(0, head) + '...' + str.slice(str.length - tail);
}

// setText обновляет текст узла, только если значение реально изменилось.
// Узел при этом не пересоздаётся: нет мигания и не сбрасывается выделение текста.
function setText(el, value) {
    const next = value == null ? '' : String(value);
    if (el.textContent !== next) el.textContent = next;
}

// syncList приводит дочерние узлы контейнера в соответствие списку items,
// переиспользуя существующие элементы по ключу. Это «сопоставление по ключу»
// (keyed reconciliation): строки обновляются на месте, лишние удаляются,
// недостающие добавляются - разметка целиком не пересобирается.
// Ровно то же самое внутри делают React/Vue, только автоматически.
function syncList(container, items, keyOf, create, update) {
    const existing = new Map();
    for (const el of Array.from(container.children)) {
        existing.set(el.dataset.key, el);
    }

    let prev = null;
    for (const item of items) {
        const key = String(keyOf(item));
        let el = existing.get(key);

        if (el) {
            existing.delete(key);
        } else {
            el = create(item);
            el.dataset.key = key;
        }

        update(el, item);

        const ref = prev ? prev.nextSibling : container.firstChild;
        if (el !== ref) container.insertBefore(el, ref);
        prev = el;
    }

    for (const el of existing.values()) el.remove();
}

// streamIPTitle собирает расшифровку адреса для подсказки при наведении:
// реальный адрес соединения, адрес из заголовков прокси и клиент плеера.
function streamIPTitle(st) {
    const parts = [];
    if (st.forwarded_ip) parts.push('X-Forwarded-For: ' + st.forwarded_ip);
    if (st.client_ip) parts.push('TCP: ' + st.client_ip);
    if (st.user_agent) parts.push(st.user_agent);
    return parts.join('\n');
}

// Показывает раздачи пользователя: таблица + активные подключения в реальном времени.
// После первой отрисовки обновляются только изменившиеся ячейки, а не вся таблица.
async function showUserTorrents(id, username) {
    let streamTimer = null;

    // Разметка создаётся один раз. Дальше меняем только текст в узлах.
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    overlay.innerHTML = `
        <div class="modal modal-wide">
            <h2>${esc(username)} - ${t('user_torrents')}</h2>
            <div class="modal-body">
                <div class="table-wrap">
                    <table class="table">
                        <thead>
                            <tr>
                                <th>${t('torrent')}</th>
                                <th>${t('connections')}</th>
                                <th>${t('stream_details')}</th>
                                <th>${t('speed')}</th>
                                <th>${t('peers_seeds')}</th>
                                <th>${t('size')}</th>
                            </tr>
                        </thead>
                        <tbody></tbody>
                    </table>
                </div>
                <p class="hint" data-empty hidden>${t('no_torrents')}</p>
            </div>
            <div class="modal-actions">
                <button class="btn btn-secondary" type="button" data-close>${t('ok')}</button>
            </div>
        </div>`;

    const tbody = overlay.querySelector('tbody');
    const emptyHint = overlay.querySelector('[data-empty]');

    // ── Строка одного подключения (потока) ───────────────────────────────
    // Колонки: адрес клиента, скорость и время. Имя файла показываем
    // в подсказке к скорости, чтобы строка оставалась компактной.
    const createStreamRow = () => {
        const row = document.createElement('div');
        row.className = 'stream-row';

        const ip = document.createElement('span');
        ip.className = 'stream-meta stream-ip';

        const speed = document.createElement('span');
        speed.className = 'stream-meta stream-speed';

        const dur = document.createElement('span');
        dur.className = 'stream-meta stream-dur';

        row.append(ip, speed, dur);
        row._c = { ip, speed, dur };
        return row;
    };

    const updateStreamRow = (row, st) => {
        const c = row._c;

        // Показываем адрес клиента: если запрос пришёл через прокси, берём его,
        // а реальный адрес соединения остаётся в подсказке.
        setText(c.ip, st.forwarded_ip || st.client_ip || '-');
        c.ip.title = streamIPTitle(st);

        setText(c.speed, fmtSpeed(st.speed_bps));
        setText(c.dur, fmtDuration(st.duration_sec));

        // В подсказке к скорости - имя файла и объём отданного.
        // В title перевод строки переносится как есть, поэтому строк может быть несколько.
        const speedTip = [];
        if (st.file_name) speedTip.push(t('file') + ': ' + st.file_name);
        if (st.bytes > 0) speedTip.push(t('transferred') + ': ' + fmtBytes(st.bytes));
        c.speed.title = speedTip.join('\n');

        c.dur.title = t('connected_since') + ': ' + new Date(st.started_at * 1000).toLocaleString();
    };

    // ── Строка раздачи ───────────────────────────────────────────────────
    const createRow = () => {
        const row = document.createElement('tr');

        const tdTitle = document.createElement('td');
        tdTitle.className = 'torrent-cell';
        const title = document.createElement('b');
        const sub = document.createElement('div');
        sub.className = 'cell-sub';
        tdTitle.append(title, sub);

        const tdConn = document.createElement('td');
        tdConn.className = 'conn-cell';
        const conn = document.createElement('span');

        const tdStreams = document.createElement('td');
        const list = document.createElement('div');
        list.className = 'stream-list';
        tdStreams.appendChild(list);

        const tdSpeed = document.createElement('td');
        tdSpeed.className = 'speed-cell';
        const speed = document.createElement('span');

        const tdPeers = document.createElement('td');
        const tdSize = document.createElement('td');
        tdSize.className = 'size-cell';

        tdConn.appendChild(conn);
        tdSpeed.appendChild(speed);

        row.append(tdTitle, tdConn, tdStreams, tdSpeed, tdPeers, tdSize);
        row._c = {
            title, sub, conn, list,
            speed, peers: tdPeers, size: tdSize,
        };
        return row;
    };

    const updateRow = (row, tr) => {
        const c = row._c;

        // Название сокращаем, чтобы длинное имя не переносилось и не растило строку.
        // Полное название, хэш и категория - в подсказке при наведении.
        setText(c.title, shortText(tr.title, 46));

        const titleTip = [tr.title || ''];
        if (tr.category) titleTip.push(t('category') + ': ' + tr.category);
        if (tr.StatString) titleTip.push(t('status') + ': ' + tr.StatString);
        if (tr.size > 0) titleTip.push(t('size') + ': ' + fmtBytes(tr.size));
        if (tr.torrent_hash) titleTip.push(t('hash') + ': ' + tr.torrent_hash);
        c.title.title = titleTip.filter(Boolean).join('\n');

        setText(c.sub, (tr.torrent_hash || '').slice(0, 12));

        // Колонка «Потоки»: просто число HTTP-потоков, без бейджа.
        const count = tr.stream_count || 0;
        if (count > 0) {
            c.conn.className = '';
            setText(c.conn, count);
        } else {
            c.conn.className = 'muted';
            setText(c.conn, '-');
        }

        // Колонка «Инфо»: адрес, скорость и время подключения.
        syncList(c.list, tr.streams || [], (st) => st.id, createStreamRow, updateStreamRow);

        if (tr.in_ram) {
            c.speed.className = '';
            // Показываем только загрузку: отдача здесь неинтересна.
            setText(c.speed, fmtSpeed(tr.download_speed));
            setText(c.peers, (tr.active_peers || 0) + '/' + (tr.total_peers || 0) + ' \u00b7 ' + (tr.connected_seeders || 0));

            const speedTip = [
                t('download') + ': ' + fmtSpeed(tr.download_speed),
                t('loaded') + ': ' + fmtBytes(tr.loaded_size) + ' / ' + fmtBytes(tr.size),
            ];
            c.speed.title = speedTip.join('\n');

            const peersTip = [
                t('active_peers') + ': ' + (tr.active_peers || 0),
                t('total_peers') + ': ' + (tr.total_peers || 0),
                t('seeders') + ': ' + (tr.connected_seeders || 0),
            ];
            c.peers.title = peersTip.join('\n');
        } else {
            c.speed.className = 'muted';
            setText(c.speed, t('sleeping'));
            c.speed.title = t('status') + ': ' + (tr.StatString || t('sleeping'));
            setText(c.peers, '-');
            c.peers.title = '';
        }

        setText(c.size, tr.size > 0 ? fmtBytes(tr.size) : '-');
        c.size.title = tr.size > 0 ? tr.size.toLocaleString() + ' ' + t('bytes') : '';

        c.conn.title = t('connections') + ': ' + count;
    };

    const close = () => {
        if (streamTimer) clearInterval(streamTimer);
        streamTimer = null;
        overlay.remove();
    };

    overlay.addEventListener('click', (e) => {
        if (e.target === overlay || e.target.hasAttribute('data-close')) close();
    });

    document.getElementById('modal-root').appendChild(overlay);

    let initialised = false;
    const load = async () => {
        try {
            const data = await api('/users/' + id + '/torrents');
            if (!overlay.isConnected) {
                close();
                return;
            }

            const torrents = data.torrents || [];
            syncList(tbody, torrents, (tr) => tr.torrent_hash, createRow, updateRow);
            emptyHint.hidden = torrents.length > 0;
            initialised = true;
        } catch (e) {
            if (!initialised && overlay.isConnected) {
                emptyHint.hidden = false;
                emptyHint.textContent = e.message;
            }
        }
    };

    await load();
    streamTimer = setInterval(load, 2000);
}

async function renderPlugins() {
    const view = document.getElementById('view');
    view.innerHTML = `
        <details class="panel manual-install">
            <summary>
                <span>${t('manual_install')}</span>
            </summary>
            <div class="manual-body">
                <div class="row">
                    <label class="file-input">
                        <input type="file" id="pl-file" accept=".zip">
                        <span class="file-input-btn">${t('choose_file')}</span>
                        <span class="file-input-name" id="pl-file-name">${t('no_file_selected')}</span>
                    </label>
                    <button class="btn" id="pl-upload" type="button">${t('upload_plugin')}</button>
                </div>
                <p class="hint">${t('upload_hint')}</p>

                <div class="row" style="margin-top:12px">
                    <input class="input" id="pl-url" placeholder="${t('plugin_url_placeholder')}">
                    <button class="btn" id="pl-url-install" type="button">${t('install_from_url')}</button>
                </div>
                <p class="hint">${t('install_from_url_hint')}</p>
            </div>
        </details>

        <div class="panel">
            <h2>${t('store')}</h2>
            <div class="catalog-wrap">
                <button class="carousel-nav left" id="cat-prev" type="button">‹</button>
                <div class="catalog-scroll" id="catalog-scroll">
                    <p class="hint">Loading...</p>
                </div>
                <button class="carousel-nav right" id="cat-next" type="button">›</button>
            </div>
            <p class="hint catalog-stale" id="catalog-stale" style="display:none">${t('catalog_stale')}</p>
        </div>

        <p class="hint">${t('drag_hint')}</p>
        <div id="plugins-table-wrap"></div>
    `;

    // ─── Manual install: file ─────────────────────────────
    const fileInput = document.getElementById('pl-file');
    const fileName = document.getElementById('pl-file-name');
    fileInput.addEventListener('change', () => {
        const f = fileInput.files[0];
        fileName.textContent = f ? f.name : t('no_file_selected');
    });

    document.getElementById('pl-upload').addEventListener('click', async () => {
        const file = fileInput.files[0];
        if (!file) { toast(t('select_file_first')); return; }
        await submitInstall(async (update) => {
            const fd = new FormData();
            fd.append('plugin', file);
            return fetch('/api/plugins/upload' + (update ? '?update' : ''), { method: 'POST', body: fd });
        });
        fileInput.value = '';
        fileName.textContent = t('no_file_selected');
    });

    // ─── Manual install: URL ──────────────────────────────
    document.getElementById('pl-url-install').addEventListener('click', async () => {
        const urlInput = document.getElementById('pl-url');
        const url = urlInput.value.trim();
        if (!url) { toast(t('enter_url_first')); return; }
        const btn = document.getElementById('pl-url-install');
        btn.disabled = true;
        btn.textContent = '...';
        try {
            await submitInstall(async (update) => {
                return fetch('/api/plugins/install-url' + (update ? '?update' : ''), {
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

    // ─── Carousel navigation ──────────────────────────────
    document.getElementById('cat-prev').addEventListener('click', () => {
        const s = document.getElementById('catalog-scroll');
        s.scrollBy({ left: -s.clientWidth * 0.8, behavior: 'smooth' });
    });
    document.getElementById('cat-next').addEventListener('click', () => {
        const s = document.getElementById('catalog-scroll');
        s.scrollBy({ left: s.clientWidth * 0.8, behavior: 'smooth' });
    });

    loadCatalog();
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
                    <td>
                        <span class="plugin-id-wrap">
                            ${p.icon ? '<img class="plugin-icon" src="' + esc(p.icon) + '" alt="" onerror="this.style.display=\'none\'">' : ''}
                            <span>${p.id}</span>
                        </span>
                        ${p.builtin ? ' <span class="badge badge-user">' + t('builtin') + '</span>' : ''}
                    </td>
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
        const list = (arr) => arr.map((x) => '<li>' + esc(x) + '</li>').join('');

        // Иконка рядом с именем
        const iconHTML = m.icon
            ? '<img class="plugin-icon" src="/plugins/' + esc(m.id) + esc(m.icon) + '" alt="" onerror="this.style.display=\'none\'">'
            : '';

        let extra = '';

        if (m.menu && m.menu.length) {
            const items = m.menu.map((e) => {
                const title = e.title_key ? (e.title + ' / ' + e.title_key) : e.title;
                return '<li>' +
                    esc(title) +
                    ' → <code>' + esc(e.route) + '</code>' +
                    ' <small>rank ' + e.rank + '</small>' +
                    '</li>';
            }).join('');
            extra += '<h3>menu</h3><ul class="info-list">' + items + '</ul>';
        }

        if (m.events && m.events.length) {
            extra += '<h3>events</h3><ul class="info-list">' + list(m.events) + '</ul>';
        }
        if (m.routes && m.routes.length) {
            extra += '<h3>routes</h3><ul class="info-list">' + list(m.routes) + '</ul>';
        }

        await modalAlert(t('info'), `
            <div class="info-head">
                ${iconHTML}
                <div class="info-head-text">
                    <div class="info-head-name">${esc(m.name)}</div>
                    <div class="info-head-id">${esc(m.id)}</div>
                </div>
            </div>
            <div class="info-grid">
                <span>version</span><b>${esc(m.version)}</b>
                <span>author</span><b>${esc(m.author || '-')}</b>
                <span>theme_ui</span><b>${m.theme_ui}</b>
                <span>builtin</span><b>${m.builtin}</b>
                <span>entry</span><b>${esc(m.entry || '-')}</b>
                <span>icon</span><b>${esc(m.icon || '-')}</b>
            </div>
            <p class="info-desc">${esc(m.description || '')}</p>
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
                    <input class="input" id="te-cap" type="number" min="1" value="${Math.max(1, Math.round((st.capacity || 0) / 1048576))}"></label>
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

// ─── Catalog ──────────────────────────────────────────────────

function parseSiloNum(v) {
    const m = /^silo\.(\d+)$/i.exec(v || '');
    return m ? parseInt(m[1], 10) : null;
}

function catalogCompatible(entry) {
    if (!entry.min_silo_version) return true;
    const need = parseSiloNum(entry.min_silo_version);
    if (need === null) return false;
    const have = parseSiloNum(SILO_VERSION);
    if (have === null) return true;
    return have >= need;
}

async function loadCatalog() {
    const scroll = document.getElementById('catalog-scroll');
    const stale = document.getElementById('catalog-stale');
    if (!scroll) return;

    scroll.innerHTML = '<p class="hint">Loading...</p>';

    try {
        const [cat, inst] = await Promise.all([
            api('/plugins/catalog'),
            api('/plugins'),
        ]);
        CATALOG = cat;

        const installedMap = new Map();
        for (const p of (inst.plugins || [])) installedMap.set(p.id, p);

        if (stale) stale.style.display = cat.stale ? '' : 'none';

        const list = cat.plugins || [];
        if (!list.length) {
            scroll.innerHTML = '<p class="hint">' + t('catalog_empty') + '</p>';
            return;
        }

        scroll.innerHTML = list.map((e) => catalogCardHTML(e, installedMap.get(e.id))).join('');
        scroll.querySelectorAll('[data-cat-action]').forEach((btn) => {
            btn.addEventListener('click', onCatalogAction);
        });
    } catch (e) {
        if (stale) stale.style.display = 'none';
        scroll.innerHTML = '<p class="hint">' + esc(t('catalog_error')) + ': ' + esc(e.message) + '</p>';
    }
}

function catalogCardHTML(entry, installed) {
    const icon = entry.icon
        ? '<img class="cat-icon" src="' + esc(entry.icon) + '" alt="" onerror="this.src=\'img/ico-plugin.svg\'">'
        : '<img class="cat-icon" src="img/ico-plugin.svg" alt="">';

    const compat = catalogCompatible(entry);
    const versionLine = 'v' + esc(entry.version) +
        (installed ? ' · installed v' + esc(installed.version) : '');

    let actionBtn;
    if (!compat) {
        actionBtn = '<button class="btn btn-sm" disabled>' + t('install') + '</button>';
    } else if (!installed) {
        actionBtn = '<button class="btn btn-sm" data-cat-action="install" data-id="' + esc(entry.id) + '">' + t('install') + '</button>';
    } else if (installed.version === entry.version) {
        actionBtn = '<button class="btn btn-secondary btn-sm" disabled>' + t('installed') + '</button>';
    } else {
        actionBtn = '<button class="btn btn-sm" data-cat-action="install" data-id="' + esc(entry.id) + '">' + t('update') + '</button>';
    }

    const compatBadge = compat
        ? ''
        : '<span class="badge badge-banned">' + t('catalog_requires') + ' ' + esc(entry.min_silo_version) + '</span>';

    return `
        <div class="cat-card">
            ${icon}
            <div class="cat-name" title="${esc(entry.name)}">${esc(entry.name)}</div>
            <div class="cat-meta">${versionLine}</div>
            ${compatBadge}
            <div class="cat-desc">${esc(entry.description || '')}</div>
            <div class="cat-actions">
                ${actionBtn}
                <button class="btn-icon" data-cat-action="info" data-id="${esc(entry.id)}" title="${t('info')}">i</button>
            </div>
        </div>
    `;
}

async function onCatalogAction(e) {
    const id = e.target.dataset.id;
    const act = e.target.dataset.catAction;

    if (act === 'info') {
        showCatalogInfo(id);
        return;
    }

    if (act === 'install') {
        e.target.disabled = true;
        const prev = e.target.textContent;
        e.target.textContent = '...';
        try {
            const res = await fetch('/api/plugins/catalog/install', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: id }),
            });
            if (!res.ok) {
                let msg = 'install failed';
                try { msg = (await res.json()).error || msg; } catch (x) {}
                toast(msg);
                return;
            }
            toast(t('done'));
            loadCatalog();
            loadPlugins();
        } catch (err) {
            toast(err.message);
        } finally {
            e.target.disabled = false;
            e.target.textContent = prev;
        }
    }
}

function showCatalogInfo(id) {
    if (!CATALOG) return;
    const entry = (CATALOG.plugins || []).find((x) => x.id === id);
    if (!entry) return;

    const iconHTML = entry.icon
        ? '<img class="plugin-icon" src="' + esc(entry.icon) + '" alt="" onerror="this.style.display=\'none\'">'
        : '';

    const compat = catalogCompatible(entry);
    const compatText = compat
        ? 'yes'
        : 'no - requires ' + esc(entry.min_silo_version);

    const linkRow = (label, url) =>
        '<span>' + label + '</span><b><a style="color:var(--silo-amber);font-size:11px;word-break:break-all" href="' + esc(url) + '" target="_blank" rel="noopener noreferrer">' + esc(url) + '</a></b>';

    let grid = '<div class="info-grid">' +
        '<span>version</span><b>' + esc(entry.version) + '</b>' +
        '<span>author</span><b>' + esc(entry.author || '-') + '</b>' +
        '<span>size</span><b>' + (entry.size ? fmtBytes(entry.size) : '-') + '</b>' +
        '<span>compatible</span><b>' + compatText + '</b>' +
        '<span>sha256</span><b class="mono" style="font-size:11px;word-break:break-all">' + esc(entry.sha256) + '</b>';
    if (entry.homepage) {
        grid += linkRow('homepage', entry.homepage);
    }
    grid += '</div>';

    modalAlert(t('info'),
        '<div class="info-head">' + iconHTML +
        '<div class="info-head-text">' +
        '<div class="info-head-name">' + esc(entry.name) + '</div>' +
        '<div class="info-head-id">' + esc(entry.id) + '</div>' +
        '</div></div>' +
        grid +
        '<p class="info-desc">' + esc(entry.description || '') + '</p>'
    );
}

boot();