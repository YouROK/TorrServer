'use strict';

const MEDIA_EXT = new Set([
    '.3g2', '.3gp', '.aaf', '.asf', '.avchd', '.avi', '.drc', '.dv', '.flv', '.iso', '.m2ts', '.m2v', '.m4p', '.m4v', '.mkv', '.mng', '.mov', '.mp2', '.mp4', '.mpe', '.mpeg', '.mpg', '.mpv', '.mts', '.mxf', '.nsv', '.ogv', '.qt', '.rm', '.rmvb', '.roq', '.svi', '.ts', '.vob', '.webm', '.wmv', '.yuv',
    '.aac', '.ac3', '.aiff', '.ape', '.au', '.dff', '.dsf', '.flac', '.gsm', '.it', '.m3u', '.m4a', '.mid', '.mod', '.mp3', '.mpa', '.mpga', '.oga', '.ogg', '.opus', '.pls', '.ra', '.s3m', '.sid', '.spx', '.wav', '.weba', '.wma', '.wv', '.wvc', '.xm',
]);

let infoHash = null;
let infoTimer = null;
let infoStatus = null;
let infoFolder = '';
let infoFilesJSON = '';
let infoFoldersJSON = '';

function isMediaName(name) {
    const i = name.lastIndexOf('.');
    if (i < 0) return false;
    return MEDIA_EXT.has(name.slice(i).toLowerCase());
}

function infoBaseFiles() {
    if (!infoStatus || !Array.isArray(infoStatus.file_stats)) return [];
    const mediaOnly = document.getElementById('info-media-only').checked;
    if (!mediaOnly) return infoStatus.file_stats;
    return infoStatus.file_stats.filter((f) => isMediaName(f.name || f.path));
}

function infoFoldersOf(files) {
    const set = new Set();
    for (const f of files) {
        let idx = f.path.lastIndexOf('/');
        if (idx <= 0) continue;
        let p = f.path.slice(0, idx);
        while (p) {
            set.add(p);
            const i2 = p.lastIndexOf('/');
            if (i2 <= 0) break;
            p = p.slice(0, i2);
        }
    }
    return Array.from(set).sort();
}

function streamUrl(idx, withToken) {
    let url = location.origin + '/api/stream/' + infoHash + '/' + idx;
    if (withToken && window.me && window.me.api_token) url += '?token=' + window.me.api_token;
    return url;
}

function renderInfo() {
    if (!infoStatus) return;

    document.getElementById('info-title').textContent = infoStatus.title || infoHash;
    document.getElementById('info-hash').textContent = infoStatus.hash || infoHash || '';
    const poster = document.getElementById('info-poster');
    if (infoStatus.poster) {
        poster.src = infoStatus.poster;
        poster.style.display = 'block';
    } else {
        poster.style.display = 'none';
    }

    const peersVal = (infoStatus.active_peers || 0) + '/' + (infoStatus.total_peers || 0) + ' · ' + (infoStatus.connected_seeders || 0);
    const plaques = [
        [t('download'), fmtSpeed(infoStatus.download_speed)],
        [t('upload'), fmtSpeed(infoStatus.upload_speed)],
        [t('peers_seeds'), peersVal],
        [t('size'), fmtSize(infoStatus.torrent_size)],
        [t('status'), infoStatus.stat_string || '-'],
        [t('category'), infoStatus.category ? categoryLabel(infoStatus.category) : '-'],
    ];
    // Плашки обновляются точечно: они меняются каждую секунду
    siloUI.render(document.getElementById('info-plaques'), plaques
        .map((p) => '<div class="plaque"><span class="plaque-label">' + p[0] + '</span><span class="plaque-value">' + esc(p[1]) + '</span></div>')
        .join(''));

    const base = infoBaseFiles();
    const folders = infoFoldersOf(base);

    if (infoFolder && folders.indexOf(infoFolder) < 0) infoFolder = '';

    const foldersBox = document.getElementById('info-folders');
    const foldersJSON = JSON.stringify(folders);
    if (foldersJSON !== infoFoldersJSON) {
        infoFoldersJSON = foldersJSON;
        if (folders.length === 0) {
            foldersBox.style.display = 'none';
            foldersBox.innerHTML = '';
        } else {
            foldersBox.style.display = 'flex';
            foldersBox.innerHTML = '<button class="info-folder" type="button" data-folder="">' + t('all_folders') + '</button>' +
                folders.map((f) => '<button class="info-folder" type="button" data-folder="' + esc(f) + '" title="' + esc(f) + '">' + esc(f) + '</button>').join('');
        }
    }
    foldersBox.querySelectorAll('.info-folder').forEach((el) => {
        el.classList.toggle('selected', (el.getAttribute('data-folder') || '') === infoFolder);
    });

    const visible = infoFolder
        ? base.filter((f) => {
            const lastSlash = f.path.lastIndexOf('/');
            const parent = lastSlash > 0 ? f.path.slice(0, lastSlash) : '';
            return parent === infoFolder;
        })
        : base;
    const filesJSON = JSON.stringify(visible.map((f) => f.id));
    if (filesJSON !== infoFilesJSON) {
        infoFilesJSON = filesJSON;
        const filesBox = document.getElementById('info-files');
        if (visible.length === 0) {
            filesBox.innerHTML = '<div class="file-empty">' + t('no_files') + '</div>';
        } else {
            filesBox.innerHTML = visible.map((f) => {
                const idx = f.path.lastIndexOf('/');
                const parent = idx > 0 ? f.path.slice(0, idx) : '';
                const sub = (parent ? parent + ' · ' : '') + fmtSize(f.length);
                return '<div class="file-row">' +
                    '<div class="file-main">' +
                    '<div class="file-name">' + esc(f.name || f.path) + '</div>' +
                    '<div class="file-sub">' + esc(sub) + '</div>' +
                    '</div>' +
                    '<div class="file-actions">' +
                    '<button class="btn btn-sm" type="button" data-open="' + f.id + '">' + t('open') + '</button>' +
                    '<button class="btn-sm-icon" type="button" data-transcode="' + f.id + '" title="' + t('transcode') + '">' +
                    '<img src="img/ico-transcode.svg" alt="" onerror="this.style.display=\'none\'">' +
                    '</button>' +
                    '<button class="btn-sm-icon" type="button" data-fileinfo="' + f.id + '" title="' + t('file_info') + '">' +
                    '<img src="img/ico-info.svg" alt="" onerror="this.textContent=\'i\'">' +
                    '</button>' +
                    '<button class="btn-sm-icon" type="button" data-preload="' + f.id + '" title="' + t('preload') + '">' +
                    '<img src="img/ico-download.svg" alt="" onerror="this.style.display=\'none\'">' +
                    '</button>' +
                    '<button class="btn-sm-icon" type="button" data-copylink="' + f.id + '" title="' + t('copy_link') + '">' +
                    '<img src="img/ico-link.svg" alt="" onerror="this.style.display=\'none\'">' +
                    '</button>' +
                    '</div>' +
                    '</div>';
            }).join('');
        }
    }
}

let currentTab = 'files';
let cacheTimer = null;
let cacheFetching = false;

async function infoPoll() {
    if (!infoHash) return;
    try {
        const res = await fetch('/api/torrents/' + infoHash);
        if (!res.ok) return;
        infoStatus = await res.json();
        renderInfo();
    } catch (e) {}

    try {
        fetch('/api/torrents/' + infoHash + '/wake', { method: 'POST' }).catch(() => {});
    } catch (e) {}
}

async function cacheTick() {
    if (!infoHash || currentTab !== 'cache') return;
    if (cacheFetching) return;
    cacheFetching = true;
    try {
        const res = await fetch('/api/torrents/' + infoHash + '/cache');
        if (res.status === 204) {
            setCacheData(null);
        } else if (res.ok) {
            setCacheData(await res.json());
        }
    } catch (e) {}
    cacheFetching = false;
}

function startCacheTimer() {
    stopCacheTimer();
    cacheTick();
    cacheTimer = setInterval(cacheTick, 100);
}

function stopCacheTimer() {
    if (cacheTimer) {
        clearInterval(cacheTimer);
        cacheTimer = null;
    }
}

function switchTab(tab) {
    currentTab = tab;

    document.querySelectorAll('.info-tab').forEach((el) => {
        el.classList.toggle('active', el.getAttribute('data-tab') === tab);
    });

    const filesPanel = document.getElementById('info-files-panel');
    const cachePanel = document.getElementById('info-cache-panel');
    const cachePlaques = document.getElementById('info-plaques-cache');

    const show = tab === 'files' ? filesPanel : cachePanel;
    const hide = tab === 'files' ? cachePanel : filesPanel;

    if (hide) {
        hide.style.display = 'none';
        hide.style.opacity = '';
    }
    if (show) {
        show.style.display = '';
        show.style.opacity = '0';
        requestAnimationFrame(() => { show.style.opacity = '1'; });
    }

    if (cachePlaques) {
        cachePlaques.classList.toggle('visible', tab === 'cache');
    }

    if (tab === 'cache' && infoHash) {
        startCacheTimer();
        renderCache();
        renderCachePlaques();
    } else {
        stopCacheTimer();
    }
}

function openInfoModal(hash) {
    infoHash = hash;
    infoFolder = '';
    infoFilesJSON = '';
    infoFoldersJSON = '';
    infoStatus = null;
    currentTab = 'files';

    const card = cardsStore.get(hash);
    document.getElementById('info-title').textContent = card ? (card.title || hash) : hash;
    document.getElementById('info-hash').textContent = hash;
    const poster = document.getElementById('info-poster');
    if (card && card.poster) {
        poster.src = card.poster;
        poster.style.display = 'block';
    } else {
        poster.style.display = 'none';
    }
    document.getElementById('info-media-only').checked = true;
    document.getElementById('info-overlay').style.display = 'flex';

    switchTab('files');
    setCacheData(null);

    infoPoll();
    infoTimer = setInterval(infoPoll, 1000);
}

function closeInfoModal() {
    if (infoTimer) clearInterval(infoTimer);
    infoTimer = null;
    stopCacheTimer();
    infoHash = null;
    infoStatus = null;
    setCacheData(null);
    document.getElementById('info-overlay').style.display = 'none';
}

function bindInfoModal() {
    document.querySelectorAll('.info-tab').forEach((el) => {
        el.addEventListener('click', () => switchTab(el.getAttribute('data-tab')));
    });

    document.getElementById('info-close').addEventListener('click', closeInfoModal);
    document.getElementById('info-hash').addEventListener('click', async () => {
        if (!infoHash) return;
        await copyText(infoHash);
        showToast(t('copied'), 'success');
    });
    document.getElementById('info-overlay').addEventListener('click', (e) => {
        if (e.target.id === 'info-overlay') closeInfoModal();
    });
    document.getElementById('info-media-only').addEventListener('change', () => {
        infoFoldersJSON = '';
        infoFilesJSON = '';
        renderInfo();
    });
    document.getElementById('info-folders').addEventListener('click', (e) => {
        const btn = e.target.closest('.info-folder');
        if (!btn) return;
        infoFolder = btn.getAttribute('data-folder') || '';
        renderInfo();
    });
    document.getElementById('info-files').addEventListener('click', async (e) => {
        const openBtn = e.target.closest('[data-open]');
        if (openBtn) {
            window.open(streamUrl(parseInt(openBtn.getAttribute('data-open'), 10), false), '_blank');
            return;
        }
        const preloadBtn = e.target.closest('[data-preload]');
        if (preloadBtn) {
            preloadBtn.disabled = true;
            try {
                const idx = parseInt(preloadBtn.getAttribute('data-preload'), 10);
                const res = await fetch('/api/torrents/' + infoHash + '/files/' + idx + '/preload', { method: 'POST' });
                if (!res.ok) {
                    let errMsg = 'preload failed';
                    try { const j = await res.json(); if (j && j.error) errMsg = j.error; } catch (x) {}
                    showToast(errMsg, 'error');
                    return;
                }
                showToast(t('preload_started'), 'success');
            } catch (err) {
                showToast(err.message, 'error');
            } finally {
                preloadBtn.disabled = false;
            }
            return;
        }
        const copyBtn = e.target.closest('[data-copylink]');
        if (copyBtn) {
            copyText(streamUrl(parseInt(copyBtn.getAttribute('data-copylink'), 10), true));
            showToast(t('copied'), 'success');
        }

        const transcodeBtn = e.target.closest('[data-transcode]');
        if (transcodeBtn) {
            showTranscodePicker(parseInt(transcodeBtn.getAttribute('data-transcode'), 10));
            return;
        }

        const infoBtn = e.target.closest('[data-fileinfo]');
        if (infoBtn) {
            showFileInfo(parseInt(infoBtn.getAttribute('data-fileinfo'), 10));
        }
    });

    bindCache();
    bindTranscodeModal();

    document.getElementById('fileinfo-close').addEventListener('click', () => {
        document.getElementById('fileinfo-overlay').style.display = 'none';
    });
}
// --- Транскодирование из списка файлов ---

let transcodeIndex = 0;
let transcodeProfiles = null;

// apiToken возвращает токен доступа текущего пользователя.
function apiToken() {
    return (window.me && window.me.api_token) ? window.me.api_token : '';
}

// transcodeUrl собирает адрес транскодированного потока.
// withToken добавляет токен: без него ссылку нельзя открыть в другом плеере.
function transcodeUrl(idx, profileID, withToken) {
    let url = location.origin + '/api/transcode/' + infoHash + '/' + idx;
    const params = [];
    if (profileID) params.push('profile=' + encodeURIComponent(profileID));
    if (withToken && apiToken()) params.push('token=' + encodeURIComponent(apiToken()));
    if (params.length) url += '?' + params.join('&');
    return url;
}

// loadProfiles получает доступные пользователю профили.
async function loadProfiles() {
    if (transcodeProfiles) return transcodeProfiles;

    const res = await fetch('/api/transcode/profiles?token=' + encodeURIComponent(apiToken()));
    if (!res.ok) throw new Error(t('transcode_unavailable'));

    const data = await res.json();
    transcodeProfiles = data.profiles || [];
    return transcodeProfiles;
}

// showTranscodePicker открывает список профилей для файла.
async function showTranscodePicker(idx) {
    const overlay = document.getElementById('transcode-overlay');
    const list = document.getElementById('transcode-list');

    document.getElementById('transcode-title').textContent = t('transcode');
    document.getElementById('transcode-close').textContent = t('cancel');
    list.innerHTML = '<div class="transcode-empty">' + t('loading') + '</div>';
    overlay.style.display = 'flex';

    let profiles = [];
    try {
        profiles = await loadProfiles();
    } catch (e) {
        list.innerHTML = '<div class="transcode-empty">' + esc(e.message) + '</div>';
        return;
    }

    if (profiles.length === 0) {
        list.innerHTML = '<div class="transcode-empty">' + t('no_profiles') + '</div>';
        return;
    }

    transcodeIndex = idx;

    // Профиль по умолчанию ставится первым: он используется без параметров
    const sorted = profiles.slice().sort((a, b) => (b.is_default ? 1 : 0) - (a.is_default ? 1 : 0));

    list.innerHTML = sorted.map((p) => {
        const details = [];
        if (p.protocol) details.push(p.protocol === 'hls' ? 'HLS' : 'MP4');
        if (p.video && p.video.codec) details.push(p.video.codec.toUpperCase());
        if (p.video && p.video.max_height) details.push(p.video.max_height + 'p');
        if (p.audio && p.audio.codec) details.push(p.audio.codec.toUpperCase());

        return '<div class="transcode-item" data-profile="' + esc(p.id) + '">' +
            '<div class="transcode-main">' +
            '<div class="transcode-name">' + esc(p.name) +
            (p.is_default ? ' <span class="transcode-star">*</span>' : '') + '</div>' +
            '<div class="transcode-sub">' + esc(details.join(' · ')) + '</div>' +
            '</div>' +
            '<div class="transcode-actions">' +
            '<button class="btn btn-sm" type="button" data-play="' + esc(p.id) + '">' + t('open') + '</button>' +
            '<button class="btn btn-secondary btn-sm" type="button" data-copy="' + esc(p.id) + '">' + t('copy') + '</button>' +
            '</div>' +
            '</div>';
    }).join('');
}

// hdrRange возвращает подпись расширенного диапазона или пустую строку.
// PQ используется в HDR10 и Dolby Vision, HLG - в телевещании.
function hdrRange(s) {
    const transfer = (s.color_transfer || '').toLowerCase();
    if (transfer === 'smpte2084') return 'HDR10';
    if (transfer === 'arib-std-b67') return 'HLG';

    if (s.bit_depth >= 10 && (s.color_primaries || '').toLowerCase() === 'bt2020') return 'HDR';
    return '';
}

// showFileInfo открывает окно с параметрами файла из ffprobe.
async function showFileInfo(idx) {
    const overlay = document.getElementById('fileinfo-overlay');
    const body = document.getElementById('fileinfo-body');

    document.getElementById('fileinfo-title').textContent = t('file_info');
    document.getElementById('fileinfo-close').textContent = t('close');
    body.innerHTML = '<div class="transcode-empty">' + t('loading') + '</div>';
    overlay.style.display = 'flex';

    let info = null;
    try {
        const res = await fetch('/api/transcode/info/' + infoHash + '/' + idx + '?token=' + encodeURIComponent(apiToken()));
        if (!res.ok) throw new Error(t('transcode_unavailable'));
        info = await res.json();
    } catch (e) {
        body.innerHTML = '<div class="transcode-empty">' + esc(e.message) + '</div>';
        return;
    }

    const rows = [];
    rows.push([t('container'), info.container || info.format || '-']);
    rows.push([t('duration'), info.duration_h || '-']);
    rows.push([t('size'), fmtSize(info.size)]);

    let html = '<table class="fileinfo-table">';
    html += rows.map((r) => '<tr><td>' + esc(r[0]) + '</td><td>' + esc(r[1]) + '</td></tr>').join('');
    html += '</table>';

    // Потоки: видео, аудио и субтитры с их параметрами
    const streams = info.streams || [];
    if (streams.length) {
        html += '<h3 class="fileinfo-head">' + t('streams') + '</h3>';
        html += '<table class="fileinfo-table">';
        html += streams.map((s) => {
            const parts = [s.codec];
            if (s.kind === 'video') {
                if (s.width) parts.push(s.width + 'x' + s.height);
                if (s.bit_depth) parts.push(s.bit_depth + ' bit');
                if (s.pix_fmt) parts.push(s.pix_fmt);
                if (s.framerate) parts.push(Number(s.framerate).toFixed(2) + ' fps');
                if (s.bitrate) parts.push(Math.round(s.bitrate / 1000) + ' kbps');
                const range = hdrRange(s);
                if (range) parts.push(range);
            } else if (s.kind === 'audio') {
                if (s.channels) parts.push(s.channels + ' ch');
                if (s.sample_rate) parts.push(s.sample_rate + ' Hz');
                if (s.bitrate) parts.push(Math.round(s.bitrate / 1000) + ' kbps');
            }
            if (s.language) parts.push(s.language);
            if (s.title) parts.push(s.title);

            return '<tr><td>' + esc(s.kind) + ' #' + s.index + '</td><td>' + esc(parts.join(' · ')) + '</td></tr>';
        }).join('');
        html += '</table>';
    }

    body.innerHTML = html;
}

// bindTranscodeModal вешает обработчики окна выбора профиля.
function bindTranscodeModal() {
    const overlay = document.getElementById('transcode-overlay');
    const list = document.getElementById('transcode-list');

    document.getElementById('transcode-close').addEventListener('click', () => {
        overlay.style.display = 'none';
    });

    list.addEventListener('click', (e) => {
        const playBtn = e.target.closest('[data-play]');
        if (playBtn) {
            window.open(transcodeUrl(transcodeIndex, playBtn.getAttribute('data-play'), false), '_blank');
            return;
        }

        const copyBtn = e.target.closest('[data-copy]');
        if (copyBtn) {
            copyText(transcodeUrl(transcodeIndex, copyBtn.getAttribute('data-copy'), true));
            showToast(t('copied'), 'success');
        }
    });
}
