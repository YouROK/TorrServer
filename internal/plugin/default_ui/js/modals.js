'use strict';

let editHash = null;

function modalConfirm(title, text) {
    return new Promise((resolve) => {
        const overlay = document.getElementById('confirm-overlay');
        const okBtn = document.getElementById('confirm-ok');
        const cancelBtn = document.getElementById('confirm-cancel');

        document.getElementById('confirm-title').textContent = title;
        document.getElementById('confirm-text').textContent = text;
        okBtn.textContent = t('delete');
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

function openAddModal() {
    const overlay = document.getElementById('add-overlay');
    overlay.style.display = 'flex';
    document.getElementById('add-form').reset();
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
    if (!link) {
        showToast(t('required_field'), 'error');
        return;
    }

    const body = {
        link: link,
        title: (fd.get('title') || '').toString().trim(),
        poster: (fd.get('poster') || '').toString().trim(),
        category: (fd.get('category') || '').toString().trim(),
        save_to_db: true,
    };

    const submitBtn = form.querySelector('[type="submit"]');
    submitBtn.disabled = true;

    try {
        const res = await fetch('/api/torrents', {
            method: 'POST',
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

        showToast(t('added_successfully'), 'success');
        closeAddModal();
    } catch (err) {
        showToast(err.message, 'error');
    } finally {
        submitBtn.disabled = false;
    }
}

function openEditModal(hash) {
    const card = cardsStore.get(hash);
    if (!card) return;
    editHash = hash;
    document.getElementById('edit-hash').textContent = hash;
    const form = document.getElementById('edit-form');
    form.reset();
    form.querySelector('[name="title"]').value = card.title || '';
    form.querySelector('[name="poster"]').value = card.poster || '';
    form.querySelector('[name="category"]').value = card.category || '';
    document.getElementById('edit-overlay').style.display = 'flex';
    setTimeout(() => form.querySelector('[name="title"]').focus(), 50);
}

function closeEditModal() {
    document.getElementById('edit-overlay').style.display = 'none';
    editHash = null;
}

async function submitEdit(e) {
    e.preventDefault();
    if (!editHash) return;
    const form = e.target;
    const fd = new FormData(form);
    const body = {
        title: (fd.get('title') || '').toString().trim(),
        poster: (fd.get('poster') || '').toString().trim(),
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
    document.getElementById('btn-add').addEventListener('click', openAddModal);
    document.getElementById('add-close').addEventListener('click', closeAddModal);
    document.getElementById('add-cancel').addEventListener('click', closeAddModal);
    document.getElementById('add-form').addEventListener('submit', submitAdd);
    document.getElementById('add-overlay').addEventListener('click', (e) => {
        if (e.target.id === 'add-overlay') closeAddModal();
    });

    document.getElementById('edit-close').addEventListener('click', closeEditModal);
    document.getElementById('edit-cancel').addEventListener('click', closeEditModal);
    document.getElementById('edit-form').addEventListener('submit', submitEdit);
    document.getElementById('edit-overlay').addEventListener('click', (e) => {
        if (e.target.id === 'edit-overlay') closeEditModal();
    });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeAddModal();
            closeEditModal();
            closeInfoModal();
            document.getElementById('confirm-overlay').style.display = 'none';
        }
    });
}