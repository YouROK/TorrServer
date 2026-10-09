'use strict';

let editHash = null;
let editLinkPoster = '';
let editFoundPosters = [];
let editSelectedPoster = '';
let editSearchTimer = null;
let editSearchSeq = 0;

function modalConfirm(title, text, okKey) {
    return new Promise((resolve) => {
        const overlay = document.getElementById('confirm-overlay');
        const okBtn = document.getElementById('confirm-ok');
        const cancelBtn = document.getElementById('confirm-cancel');

        document.getElementById('confirm-title').textContent = title;
        document.getElementById('confirm-text').textContent = text;
        okBtn.textContent = t(okKey || 'delete');
        cancelBtn.textContent = t('cancel');
        overlay.style.display = 'flex';

        const done = (val) => {
            overlay.style.display = 'none';
            okBtn.removeEventListener('click', onOk);
            cancelBtn.removeEventListener('click', onCancel);
            overlay.removeEventListener('click', onBack);
            resolve(val);
        };
        const onOk = () => done(true);
        const onCancel = () => done(false);
        const onBack = (e) => { if (e.target === overlay) done(false); };

        okBtn.addEventListener('click', onOk);
        cancelBtn.addEventListener('click', onCancel);
        overlay.addEventListener('click', onBack);
    });
}

// editStripUrls собирает список постеров: первым идет ссылка из поля, затем найденные в TMDB
function editStripUrls() {
    const list = [];
    const seen = new Set();

    const push = (url) => {
        const value = (url || '').trim();
        if (!value || seen.has(value)) return;
        seen.add(value);
        list.push(value);
    };

    push(editLinkPoster);
    for (const url of editFoundPosters) push(url);
    return list;
}

function editRenderStrip() {
    const strip = document.getElementById('edit-poster-strip');
    if (!strip) return;

    const list = editStripUrls();
    if (list.length === 0) {
        strip.innerHTML = '<span class="poster-empty">' + esc(t('poster_none')) + '</span>';
        return;
    }

    strip.innerHTML = list.map((url) => {
        const active = url === editSelectedPoster ? ' active' : '';
        return '<button class="poster-item' + active + '" type="button" data-url="' + esc(url) + '">' +
            '<img src="' + esc(url) + '" alt="" loading="lazy" ' +
            'onerror="this.closest(\'.poster-item\').classList.add(\'broken\')">' +
            '</button>';
    }).join('');
}

function editRenderPreview() {
    const img = document.getElementById('edit-preview-img');
    const empty = document.getElementById('edit-preview-empty');
    if (!img || !empty) return;

    if (editSelectedPoster) {
        img.src = editSelectedPoster;
        img.hidden = false;
        empty.hidden = true;
    } else {
        img.removeAttribute('src');
        img.hidden = true;
        empty.hidden = false;
    }
}

function editSetPoster(url) {
    editSelectedPoster = (url || '').trim();
    const input = document.getElementById('edit-poster-url');
    if (input && input.value.trim() !== editSelectedPoster) input.value = editSelectedPoster;
    editLinkPoster = editSelectedPoster;
    editRenderPreview();
    editRenderStrip();
}

function editSetStatus(key) {
    const status = document.getElementById('edit-poster-status');
    if (status) status.textContent = key ? t(key) : '';
}

// editSearchPosters подтягивает постеры из TMDB по текущему названию
function editSearchPosters() {
    if (editSearchTimer) {
        clearTimeout(editSearchTimer);
        editSearchTimer = null;
    }

    const query = shortenPosterQuery(document.querySelector('#edit-form [name="title"]').value);
    if (query.length < POSTER_SEARCH_MIN_LEN) {
        editFoundPosters = [];
        editSetStatus('');
        editRenderStrip();
        editRenderPreview();
        return;
    }

    const seq = ++editSearchSeq;
    editSetStatus('poster_searching');
    editSearchTimer = setTimeout(async () => {
        editSearchTimer = null;
        try {
            const found = await tmdbPosters(query);
            if (seq !== editSearchSeq) return;
            editFoundPosters = found;
            editSetStatus(found.length > 0 ? '' : 'poster_not_found');
        } catch (e) {
            if (seq !== editSearchSeq) return;
            editFoundPosters = [];
            editSetStatus('poster_not_found');
        }
        editRenderStrip();
        editRenderPreview();
    }, POSTER_SEARCH_DELAY);
}

// editSyncFromUrlInput переносит ссылку из поля в начало списка постеров
function editSyncFromUrlInput() {
    editLinkPoster = document.getElementById('edit-poster-url').value.trim();
    editSelectedPoster = editLinkPoster || editFoundPosters[0] || '';
    editRenderStrip();
    editRenderPreview();
}

function openEditModal(hash) {
    const card = cardsStore.get(hash);
    if (!card) return;

    editHash = hash;
    editLinkPoster = card.poster || '';
    editFoundPosters = [];
    editSelectedPoster = editLinkPoster;
    if (editSearchTimer) {
        clearTimeout(editSearchTimer);
        editSearchTimer = null;
    }
    editSearchSeq++;

    const hashEl = document.getElementById('edit-hash');
    hashEl.textContent = shortHash(hash);
    hashEl.title = hash;
    const form = document.getElementById('edit-form');
    form.reset();
    form.querySelector('[name="title"]').value = card.title || '';
    document.getElementById('edit-poster-url').value = card.poster || '';
    categoryFillSelect(document.getElementById('edit-category'), card.category || '');

    editSetStatus('');
    editRenderStrip();
    editRenderPreview();

    document.getElementById('edit-overlay').style.display = 'flex';
    setTimeout(() => form.querySelector('[name="title"]').focus(), 50);
    editSearchPosters();
}

function closeEditModal() {
    document.getElementById('edit-overlay').style.display = 'none';
    if (editSearchTimer) {
        clearTimeout(editSearchTimer);
        editSearchTimer = null;
    }
    editSearchSeq++;
    editHash = null;
}

async function submitEdit(e) {
    e.preventDefault();
    if (!editHash) return;

    const form = e.target;
    const fd = new FormData(form);
    const body = {
        title: (fd.get('title') || '').toString().trim(),
        poster: (fd.get('poster') || '').toString().trim() || editSelectedPoster,
        category: (fd.get('category') || '').toString().trim(),
    };

    const submitBtn = form.querySelector('[type="submit"]');
    submitBtn.disabled = true;

    try {
        const res = await fetch('/api/torrents/' + editHash + '/meta', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });

        if (!res.ok) {
            let errMsg = 'error';
            try {
                const j = await res.json();
                if (j && j.error) errMsg = j.error;
            } catch (x) {}
            showToast(errMsg, 'error');
            return;
        }

        showToast(t('updated_successfully'), 'success');
        closeEditModal();
    } catch (err) {
        showToast(err.message, 'error');
    } finally {
        submitBtn.disabled = false;
    }
}

function bindModals() {
    const editTitleInput = document.querySelector('#edit-form [name="title"]');
    const editPosterInput = document.getElementById('edit-poster-url');
    const editStrip = document.getElementById('edit-poster-strip');
    const editSelect = document.getElementById('edit-category');

    document.getElementById('edit-close').addEventListener('click', closeEditModal);
    document.getElementById('edit-cancel').addEventListener('click', closeEditModal);
    document.getElementById('edit-form').addEventListener('submit', submitEdit);
    document.getElementById('edit-overlay').addEventListener('click', (e) => {
        if (e.target.id === 'edit-overlay') closeEditModal();
    });

    editTitleInput.addEventListener('input', () => editSearchPosters());
    editPosterInput.addEventListener('input', () => editSyncFromUrlInput());
    editStrip.addEventListener('click', (e) => {
        const item = e.target.closest('.poster-item');
        if (!item) return;
        editSetPoster(item.getAttribute('data-url'));
    });

    categoryFillSelect(editSelect, '');
    editRenderStrip();
    editRenderPreview();

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeAddModal();
            closeEditModal();
            closeInfoModal();
            document.getElementById('confirm-overlay').style.display = 'none';
        }
    });
}
