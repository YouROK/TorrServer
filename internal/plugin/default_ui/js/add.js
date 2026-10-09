'use strict';

// Публичный ключ TMDB: предназначен для клиентских запросов и не является секретом
const TMDB_API_KEY = '45ddf563ac3fb845f2d5c363190d1a33';
const TMDB_API_URL = 'https://api.themoviedb.org/3/search/multi';
const TMDB_IMAGE_URL = 'https://image.tmdb.org/t/p/w342';
const POSTER_SEARCH_DELAY = 600;
const POSTER_SEARCH_MIN_LEN = 2;
const POSTER_SEARCH_MAX_LEN = 50;
const POSTER_SEARCH_MAX_WORDS = 4;

let addFile = null;
let addLinkPoster = '';
let addFoundPosters = [];
let addSelectedPoster = '';
let posterSearchTimer = null;
let posterSearchSeq = 0;

// shortenPosterQuery отрезает релизный мусор, чтобы поиск попадал в название фильма
function shortenPosterQuery(title) {
    if (!title) return '';
    let base = String(title).trim();
    const cut = base.search(/[[(]/);
    if (cut > 0) base = base.slice(0, cut);
    base = base.trim();

    const words = base.split(/\s+/).slice(0, POSTER_SEARCH_MAX_WORDS).join(' ');
    return words.length > POSTER_SEARCH_MAX_LEN ? words.slice(0, POSTER_SEARCH_MAX_LEN) : words;
}

// tmdbPosters ищет постеры по названию раздачи
async function tmdbPosters(title) {
    const query = shortenPosterQuery(title);
    if (query.length < POSTER_SEARCH_MIN_LEN) return [];

    const lang = (typeof i18n !== 'undefined' && i18n.get) ? i18n.get() : 'en';
    const url = TMDB_API_URL +
        '?api_key=' + encodeURIComponent(TMDB_API_KEY) +
        '&language=' + encodeURIComponent(lang) +
        '&include_image_language=' + encodeURIComponent(lang + ',null,en') +
        '&query=' + encodeURIComponent(query);

    const res = await fetch(url, { cache: 'no-store' });
    if (!res.ok) throw new Error('tmdb request failed');

    const data = await res.json();
    const results = Array.isArray(data.results) ? data.results : [];
    const seen = new Set();
    const urls = [];

    for (const item of results) {
        if (!item || !item.poster_path) continue;
        const full = TMDB_IMAGE_URL + item.poster_path;
        if (seen.has(full)) continue;
        seen.add(full);
        urls.push(full);
    }
    return urls;
}

// addStripUrls собирает список постеров: первым идет ссылка из поля, затем найденные в TMDB
function addStripUrls() {
    const list = [];
    const seen = new Set();

    const push = (url) => {
        const value = (url || '').trim();
        if (!value || seen.has(value)) return;
        seen.add(value);
        list.push(value);
    };

    push(addLinkPoster);
    for (const url of addFoundPosters) push(url);
    return list;
}

function addRenderStrip() {
    const strip = document.getElementById('add-poster-strip');
    if (!strip) return;

    const list = addStripUrls();
    if (list.length === 0) {
        strip.innerHTML = '<span class="poster-empty">' + esc(t('poster_none')) + '</span>';
        return;
    }

    strip.innerHTML = list.map((url) => {
        const active = url === addSelectedPoster ? ' active' : '';
        return '<button class="poster-item' + active + '" type="button" data-url="' + esc(url) + '">' +
            '<img src="' + esc(url) + '" alt="" loading="lazy" ' +
            'onerror="this.closest(\'.poster-item\').classList.add(\'broken\')">' +
            '</button>';
    }).join('');
}

function addRenderPreview() {
    const img = document.getElementById('add-preview-img');
    const empty = document.getElementById('add-preview-empty');
    if (!img || !empty) return;

    if (addSelectedPoster) {
        img.src = addSelectedPoster;
        img.hidden = false;
        empty.hidden = true;
    } else {
        img.removeAttribute('src');
        img.hidden = true;
        empty.hidden = false;
    }
}

function addSetPoster(url) {
    addSelectedPoster = (url || '').trim();
    const input = document.getElementById('add-poster-url');
    if (input && input.value.trim() !== addSelectedPoster) input.value = addSelectedPoster;
    addLinkPoster = addSelectedPoster;
    addRenderPreview();
    addRenderStrip();
}

function addSetStatus(key) {
    const status = document.getElementById('add-poster-status');
    if (status) status.textContent = key ? t(key) : '';
}

// addSearchPosters подтягивает постеры из TMDB по текущему названию
function addSearchPosters() {
    if (posterSearchTimer) {
        clearTimeout(posterSearchTimer);
        posterSearchTimer = null;
    }

    const title = document.querySelector('#add-form [name="title"]').value;
    const query = shortenPosterQuery(title);

    if (query.length < POSTER_SEARCH_MIN_LEN) {
        addFoundPosters = [];
        addSelectedPoster = addLinkPoster;
        addSetStatus('');
        addRenderStrip();
        addRenderPreview();
        return;
    }

    const seq = ++posterSearchSeq;
    addSetStatus('poster_searching');
    posterSearchTimer = setTimeout(async () => {
        posterSearchTimer = null;
        try {
            const found = await tmdbPosters(query);
            if (seq !== posterSearchSeq) return;

            addFoundPosters = found;
            addSetStatus(found.length > 0 ? '' : 'poster_not_found');
            if (!addSelectedPoster) addSelectedPoster = addLinkPoster || found[0] || '';
        } catch (e) {
            if (seq !== posterSearchSeq) return;
            addFoundPosters = [];
            addSetStatus('poster_not_found');
        }
        addRenderStrip();
        addRenderPreview();
    }, POSTER_SEARCH_DELAY);
}

// addSyncFromUrlInput переносит ссылку из поля в начало списка постеров
function addSyncFromUrlInput() {
    addLinkPoster = document.getElementById('add-poster-url').value.trim();
    addSelectedPoster = addLinkPoster || addFoundPosters[0] || '';
    addRenderStrip();
    addRenderPreview();
}

function addSetFile(file) {
    addFile = file || null;

    const nameEl = document.getElementById('add-file-name');
    const clearBtn = document.getElementById('add-file-clear');
    const drop = document.getElementById('add-drop');

    if (addFile) {
        nameEl.textContent = addFile.name;
        nameEl.hidden = false;
        clearBtn.hidden = false;
        drop.classList.add('has-file');
    } else {
        nameEl.textContent = '';
        nameEl.hidden = true;
        clearBtn.hidden = true;
        drop.classList.remove('has-file');
    }
}

function addResetForm() {
    document.getElementById('add-form').reset();
    categoryFillSelect(document.getElementById('add-category'), '');
    addSetFile(null);
    addLinkPoster = '';
    addFoundPosters = [];
    addSelectedPoster = '';
    if (posterSearchTimer) {
        clearTimeout(posterSearchTimer);
        posterSearchTimer = null;
    }
    posterSearchSeq++;
    addSetStatus('');
    addRenderStrip();
    addRenderPreview();
}

function openAddModal() {
    const overlay = document.getElementById('add-overlay');
    overlay.style.display = 'flex';
    addResetForm();
    setTimeout(() => document.querySelector('#add-form [name="link"]').focus(), 50);
}

function closeAddModal() {
    document.getElementById('add-overlay').style.display = 'none';
}

async function submitAdd(e) {
    e.preventDefault();
    const form = e.target;
    const fd = new FormData(form);
    const link = (fd.get('link') || '').toString().trim();

    if (!link && !addFile) {
        showToast(t('required_field'), 'error');
        return;
    }

    const title = (fd.get('title') || '').toString().trim();
    const poster = addSelectedPoster || (fd.get('poster') || '').toString().trim();
    const category = (fd.get('category') || '').toString().trim();

    let body;
    let headers = {};
    if (addFile) {
        body = new FormData();
        body.append('file', addFile);
        body.append('link', link);
        body.append('title', title);
        body.append('poster', poster);
        body.append('category', category);
        body.append('save_to_db', 'true');
    } else {
        headers = { 'Content-Type': 'application/json' };
        body = JSON.stringify({ link, title, poster, category, save_to_db: true });
    }

    const submitBtn = form.querySelector('[type="submit"]');
    submitBtn.disabled = true;

    try {
        const res = await fetch('/api/torrents', { method: 'POST', headers, body });
        if (!res.ok) {
            let errMsg = 'error';
            try {
                const j = await res.json();
                if (j && j.error) errMsg = j.error;
            } catch (x) {}
            showToast(errMsg, 'error');
            return;
        }

        showToast(t('added_successfully'), 'success');
        closeAddModal();
    } catch (err) {
        showToast(err.message, 'error');
    } finally {
        submitBtn.disabled = false;
    }
}

function bindAddModal() {
    const drop = document.getElementById('add-drop');
    const fileInput = document.getElementById('add-file');
    const titleInput = document.querySelector('#add-form [name="title"]');
    const posterInput = document.getElementById('add-poster-url');
    const strip = document.getElementById('add-poster-strip');

    document.getElementById('btn-add').addEventListener('click', openAddModal);
    document.getElementById('add-close').addEventListener('click', closeAddModal);
    document.getElementById('add-cancel').addEventListener('click', closeAddModal);
    document.getElementById('add-form').addEventListener('submit', submitAdd);
    document.getElementById('add-overlay').addEventListener('click', (e) => {
        if (e.target.id === 'add-overlay') closeAddModal();
    });

    fileInput.addEventListener('change', () => addSetFile(fileInput.files[0] || null));

    document.getElementById('add-file-clear').addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();
        fileInput.value = '';
        addSetFile(null);
    });

    // Перетаскивание файла на зону загрузки
    drop.addEventListener('dragover', (e) => {
        e.preventDefault();
        drop.classList.add('dragging');
    });
    drop.addEventListener('dragleave', () => drop.classList.remove('dragging'));
    drop.addEventListener('drop', (e) => {
        e.preventDefault();
        drop.classList.remove('dragging');
        const files = e.dataTransfer ? e.dataTransfer.files : null;
        if (!files || files.length === 0) return;
        // В некоторых браузерах присваивание fileList запрещено: тогда шлем файл напрямую
        try {
            fileInput.files = files;
        } catch (x) {}
        addSetFile(files[0]);
    });

    titleInput.addEventListener('input', () => {
        addSearchPosters();
    });

    posterInput.addEventListener('input', () => {
        addSyncFromUrlInput();
    });

    strip.addEventListener('click', (e) => {
        const item = e.target.closest('.poster-item');
        if (!item) return;
        addSetPoster(item.getAttribute('data-url'));
    });

    categoryFillSelect(document.getElementById('add-category'), '');
    addRenderStrip();
    addRenderPreview();
}
