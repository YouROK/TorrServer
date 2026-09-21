'use strict';

const DONATE_URLS = {
    boosty:   'https://boosty.to/yourok',
    yoomoney: 'https://yoomoney.ru/to/410013733697114/200',
    tbank:    'https://www.tbank.ru/cf/742qEMhKhKn',
};

const CONTRIBUTORS = [
    { initials: 'MJ', name: 'Matt Joiner',        url: 'https://github.com/anacrolix' },
    { initials: 'DS', name: 'Daniel Shleifman',   url: 'https://github.com/dancheskus' },
    { initials: 'NI', name: 'nikk',               url: 'https://github.com/tsynik' },
    { initials: 'KO', name: 'kolsys',             url: 'https://github.com/kolsys' },
    { initials: 'TW', name: 'tw1cker',            url: 'https://github.com/Nemirov' },
    { initials: 'SP', name: 'SpAwN_LMG',          url: 'https://github.com/spawnlmg' },
    { initials: 'DA', name: 'damiva',             url: 'https://github.com/damiva' },
    { initials: 'VL', name: 'Vladlenas',          url: 'https://github.com/vladlenas' },
    { initials: 'PP', name: 'Pavel Pikta',        url: 'https://github.com/pavelpikta' },
    { initials: 'AP', name: 'Anton Potekhin',     url: 'https://github.com/Anton111111' },
    { initials: 'FA', name: 'FaintGhost',         url: 'https://github.com/FaintGhost' },
    { initials: 'TO', name: 'TopperBG',           url: 'https://github.com/TopperBG' },
    { initials: 'EV', name: 'Evgeni',             url: 'https://github.com/lieranderl' },
    { initials: 'CO', name: 'cocool97',           url: 'https://github.com/cocool97' },
    { initials: 'SH', name: 'shadeov',            url: 'https://github.com/shadeov' },
    { initials: 'PA', name: 'Pavel',              url: 'https://github.com/butaford' },
    { initials: 'AF', name: 'Alexey Filimonov',   url: 'https://github.com/filimonic' },
    { initials: 'VE', name: 'Viacheslav Evseev',  url: 'https://github.com/leporel' },
];

// ─── Mobile sidebar ────────────────────────────────────────────

function openSidebar() {
    document.getElementById('sidebar').classList.add('open');
    document.getElementById('sidebar-overlay').classList.add('show');
}

function closeSidebar() {
    document.getElementById('sidebar').classList.remove('open');
    document.getElementById('sidebar-overlay').classList.remove('show');
}

// ─── Brand / category / donate / about ─────────────────────────

function bindMenuBase() {
    document.getElementById('brand').addEventListener('click', () => {
        location.href = '/';
    });

    document.getElementById('btn-category').addEventListener('click', () => {
        showToast(t('category_coming_soon'), 'info');
    });

    document.getElementById('btn-donate').addEventListener('click', () => {
        document.getElementById('donate-overlay').style.display = 'flex';
    });

    document.getElementById('btn-about').addEventListener('click', () => {
        openAboutModal();
    });

    // Универсальный обработчик [data-close]
    document.querySelectorAll('[data-close]').forEach((el) => {
        el.addEventListener('click', () => {
            document.getElementById(el.getAttribute('data-close')).style.display = 'none';
        });
    });

    // Клик по фону закрывает модалки
    ['donate-overlay', 'about-overlay'].forEach((id) => {
        document.getElementById(id).addEventListener('click', (e) => {
            if (e.target.id === id) e.target.style.display = 'none';
        });
    });
}

// ─── About modal ────────────────────────────────────────────────

async function openAboutModal() {
    document.getElementById('about-overlay').style.display = 'flex';

    const ver = document.getElementById('about-version');
    ver.textContent = '—';
    try {
        const res = await fetch('/api/system/version');
        if (res.ok) {
            const d = await res.json();
            const torrent = d.torrent ? ' · torrent ' + d.torrent : '';
            ver.textContent = 'TorrServer ' + (d.version || 'Silo.0') + torrent;
        }
    } catch (e) {}
}

function renderContributors() {
    const box = document.getElementById('about-contributors');
    if (!box || box.dataset.rendered) return;
    box.dataset.rendered = '1';
    box.innerHTML = CONTRIBUTORS.map((c) =>
        '<li><a href="' + esc(c.url) + '" target="_blank" rel="noopener noreferrer">' +
        '<span class="uc-initials">' + esc(c.initials) + '</span>' +
        '<span>' + esc(c.name) + '</span>' +
        '</a></li>'
    ).join('');
}

// ─── Plugins menu ───────────────────────────────────────────────

function pluginItemHTML(item) {
    const icon = item.icon || 'img/ico-plugin.svg';
    const label = item.title_key ? t(item.title_key) : (item.title || item.plugin_id);
    const href = item.href || ('/plugins/' + item.plugin_id + '/');
    return '<a class="nav-item" href="' + esc(href) + '">' +
        '<img class="nav-icon" src="' + esc(icon) + '" alt="" onerror="this.src=\'img/ico-plugin.svg\'">' +
        '<span>' + esc(label) + '</span>' +
        '</a>';
}

async function loadPluginsMenu() {
    const section = document.getElementById('plugins-section');
    const box = document.getElementById('plugins-menu');
    try {
        const res = await fetch('/api/ui/menu');
        if (!res.ok) { section.hidden = true; return; }
        const data = await res.json();
        const items = (data.menu || []);
        if (!items.length) { section.hidden = true; return; }

        // Группируем по plugin_id, сохраняя порядок появления
        const groups = new Map();
        for (const it of items) {
            const key = it.plugin_id;
            if (!groups.has(key)) {
                groups.set(key, { name: it.plugin_name || key, items: [] });
            }
            groups.get(key).items.push(it);
        }

        let html = '';
        for (const [, g] of groups) {
            html += '<div class="plugin-group">' +
                '<div class="plugin-group-name" title="' + esc(g.name) + '">' + esc(g.name) + '</div>' +
                g.items.map(pluginItemHTML).join('') +
                '</div>';
        }
        box.innerHTML = html;
        section.hidden = false;
    } catch (e) {
        section.hidden = true;
    }
}

// ─── Remove All ─────────────────────────────────────────────────

async function removeAllTorrents() {
    const ok = await modalConfirm(t('remove_all_title'), t('remove_all_text'), 'remove_all');
    if (!ok) return;

    let list;
    try {
        const res = await fetch('/api/torrents');
        if (!res.ok) throw new Error('failed to load list');
        const data = await res.json();
        list = data.torrents || [];
    } catch (e) {
        showToast(e.message, 'error');
        return;
    }

    if (!list.length) {
        showToast(t('empty_library'), 'info');
        return;
    }

    let removed = 0;
    let failed = 0;
    for (const torrent of list) {
        try {
            const res = await fetch('/api/torrents/' + torrent.hash, { method: 'DELETE' });
            if (res.ok) removed++; else failed++;
        } catch (e) {
            failed++;
        }
    }

    if (failed) {
        showToast(t('remove_all_removed') + ': ' + removed + ', ' + t('remove_all_failed') + ': ' + failed, 'error');
    } else {
        showToast(t('remove_all_removed') + ': ' + removed, 'success');
    }
}

// ─── Logout / Turn Off ──────────────────────────────────────────

async function doLogout() {
    try { await fetch('/api/auth/logout', { method: 'POST' }); } catch (e) {}
    location.href = '/login';
}

async function doTurnOff() {
    const ok = await modalConfirm(t('turn_off_title'), t('turn_off_text'), 'turn_off');
    if (!ok) return;
    try {
        await fetch('/api/system/shutdown', { method: 'POST' });
        showToast(t('turn_off_done'), 'info');
    } catch (e) {
        showToast(e.message, 'error');
    }
}

// ─── User block ─────────────────────────────────────────────────

function renderUser(me) {
    const nameEl = document.getElementById('username');
    const avatarEl = document.getElementById('user-avatar');
    const turnOffBtn = document.getElementById('btn-turn-off');

    if (!me) {
        nameEl.textContent = t('guest');
        avatarEl.textContent = '?';
        return;
    }

    nameEl.textContent = me.username || me.id || '—';
    avatarEl.textContent = (me.username || me.id || '?').charAt(0).toUpperCase();

    if (typeof me.rank === 'number' && me.rank >= 100) {
        turnOffBtn.hidden = false;
    }
}

// ─── Init ───────────────────────────────────────────────────────

function bindMenu() {
    bindMenuBase();
    renderContributors();

    document.getElementById('btn-remove-all').addEventListener('click', removeAllTorrents);
    document.getElementById('btn-logout').addEventListener('click', doLogout);
    document.getElementById('btn-turn-off').addEventListener('click', doTurnOff);

    // Мобильный гамбургер
    const ham = document.getElementById('hamburger');
    if (ham) ham.addEventListener('click', openSidebar);
    document.getElementById('sidebar-overlay').addEventListener('click', closeSidebar);

    // Закрываем сайдбар после перехода по ссылке плагина (мобильный)
    document.getElementById('plugins-menu').addEventListener('click', (e) => {
        if (e.target.closest('a')) closeSidebar();
    });
}