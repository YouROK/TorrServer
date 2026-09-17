'use strict';

const cardsStore = new Map();

function badgeFor(stat) {
    if (stat === 0) return { cls: 'stat-0', key: 'stat_added' };
    if (stat === 1) return { cls: 'stat-1', key: 'stat_getting_info' };
    if (stat === 2) return { cls: 'stat-2', key: 'stat_preload' };
    if (stat === 3) return { cls: 'stat-3', key: 'stat_working' };
    return null;
}

function menuItem(action, icon, key, danger) {
    return '<button class="card-menu-item' + (danger ? ' danger' : '') + '" type="button" data-action="' + action + '">' +
        '<img src="img/' + icon + '" alt="" onerror="this.style.display=\'none\'">' +
        '<span>' + t(key) + '</span>' +
        '</button>';
}

function cardHTML(c) {
    if (!c) return '';
    const hash = c.hash || '';
    const title = c.title || hash || '—';
    const poster = c.poster || '';
    const badge = badgeFor(c.stat);

    const badgeHTML = badge
        ? '<span class="card-badge ' + badge.cls + '">' + t(badge.key) + '</span>'
        : '';

    const posterHTML = poster
        ? '<img src="' + esc(poster) + '" alt="" loading="lazy">'
        : '<div class="placeholder">' + t('no_poster') + '</div>';

    const dotsHTML = '<button class="card-dots" type="button" aria-label="menu">' +
        '<svg viewBox="0 0 24 24"><circle cx="12" cy="5" r="2"></circle><circle cx="12" cy="12" r="2"></circle><circle cx="12" cy="19" r="2"></circle></svg>' +
        '</button>';

    const menuHTML = '<div class="card-menu">' +
        menuItem('info', 'ico-info.svg', 'menu_info') +
        menuItem('edit', 'ico-edit.svg', 'menu_edit') +
        menuItem('torrs', 'ico-torrs.svg', 'menu_copy_torrs') +
        menuItem('magnet', 'ico-magnet.svg', 'menu_copy_magnet') +
        menuItem('hash', 'ico-hash.svg', 'menu_copy_hash') +
        menuItem('delete', 'ico-delete.svg', 'menu_delete', true) +
        '</div>';

    return '' +
        '<div class="card" data-hash="' + hash + '">' +
        '<div class="poster">' + posterHTML + badgeHTML + dotsHTML + menuHTML + '</div>' +
        '<div class="title">' + esc(title) + '</div>' +
        '<div class="meta">' +
        '<span class="meta-item" title="' + t('size') + '">' +
        '<img src="img/ico-size.svg" alt="" onerror="this.style.display=\'none\'">' +
        '<em>' + fmtSize(c.size) + '</em>' +
        '</span>' +
        '<span class="meta-item" title="' + t('peers_seeds') + '">' +
        '<img src="img/ico-peers.svg" alt="" onerror="this.style.display=\'none\'">' +
        '<em>' + fmtPeers(c.peers, c.seeders) + '</em>' +
        '</span>' +
        '<span class="meta-item" title="' + t('download_speed') + '">' +
        '<img src="img/ico-speed.svg" alt="" onerror="this.style.display=\'none\'">' +
        '<em>' + fmtSpeed(c.download_speed) + '</em>' +
        '</span>' +
        '</div>' +
        '</div>';
}

function renderSync(cards) {
    const grid = document.getElementById('grid');
    if (!grid) return;
    cardsStore.clear();
    if (!Array.isArray(cards) || cards.length === 0) {
        grid.innerHTML = '<p class="grid-msg">' + t('empty_library') + '</p>';
        return;
    }
    for (const c of cards) cardsStore.set(c.hash, c);
    grid.innerHTML = cards.map(cardHTML).join('');
}

function patchCard(el, c) {
    const titleEl = el.querySelector('.title');
    const newTitle = c.title || c.hash || '—';
    if (titleEl && titleEl.textContent !== newTitle) {
        titleEl.textContent = newTitle;
    }

    const posterEl = el.querySelector('.poster');
    const img = posterEl.querySelector('img');
    const placeholder = posterEl.querySelector('.placeholder');
    if (c.poster) {
        if (img) {
            if (img.getAttribute('src') !== c.poster) img.setAttribute('src', c.poster);
        } else if (placeholder) {
            placeholder.outerHTML = '<img src="' + esc(c.poster) + '" alt="" loading="lazy">';
        }
    } else if (img) {
        img.outerHTML = '<div class="placeholder">' + t('no_poster') + '</div>';
    }

    const badge = badgeFor(c.stat);
    let badgeEl = posterEl.querySelector('.card-badge');
    if (badge) {
        if (!badgeEl) {
            posterEl.insertAdjacentHTML('afterbegin', '<span class="card-badge ' + badge.cls + '">' + t(badge.key) + '</span>');
        } else {
            badgeEl.className = 'card-badge ' + badge.cls;
            badgeEl.textContent = t(badge.key);
        }
    } else if (badgeEl) {
        badgeEl.remove();
    }

    const ems = el.querySelectorAll('.meta-item em');
    if (ems.length === 3) {
        const vals = [fmtSize(c.size), fmtPeers(c.peers, c.seeders), fmtSpeed(c.download_speed)];
        for (let i = 0; i < 3; i++) {
            if (ems[i].textContent !== vals[i]) ems[i].textContent = vals[i];
        }
    }
}

function upsertCard(c) {
    if (!c || !c.hash) return;
    const grid = document.getElementById('grid');
    if (!grid) return;
    const msg = grid.querySelector('.grid-msg');
    if (msg) msg.remove();

    cardsStore.set(c.hash, c);
    const existing = grid.querySelector('.card[data-hash="' + c.hash + '"]');
    if (existing) {
        patchCard(existing, c);
    } else {
        grid.insertAdjacentHTML('beforeend', cardHTML(c));
    }
}

function removeCard(hash) {
    if (!hash) return;
    const grid = document.getElementById('grid');
    if (!grid) return;
    cardsStore.delete(hash);
    const existing = grid.querySelector('.card[data-hash="' + hash + '"]');
    if (existing) existing.remove();
    if (!grid.querySelector('.card')) {
        grid.innerHTML = '<p class="grid-msg">' + t('empty_library') + '</p>';
    }
}

async function handleCardAction(action, hash) {
    const card = cardsStore.get(hash);

    if (action === 'info') {
        openInfoModal(hash);
        return;
    }

    if (action === 'edit') {
        openEditModal(hash);
        return;
    }

    if (action === 'torrs') {
        if (card && card.torrs) {
            await copyText(card.torrs);
            showToast(t('copied'), 'success');
        }
        return;
    }

    if (action === 'magnet') {
        await copyText('magnet:?xt=urn:btih:' + hash);
        showToast(t('copied'), 'success');
        return;
    }

    if (action === 'hash') {
        await copyText(hash);
        showToast(t('copied'), 'success');
        return;
    }

    if (action === 'delete') {
        const ok = await modalConfirm(t('confirm_delete_title'), t('confirm_delete_text'));
        if (!ok) return;
        try {
            const res = await fetch('/api/torrents/' + hash, { method: 'DELETE' });
            if (!res.ok) throw new Error('delete failed');
            showToast(t('deleted'), 'success');
        } catch (e) {
            showToast(e.message, 'error');
        }
    }
}

function bindGrid() {
    const grid = document.getElementById('grid');

    grid.addEventListener('click', (e) => {
        const dots = e.target.closest('.card-dots');
        if (dots) {
            dots.closest('.card').classList.toggle('menu-open');
            return;
        }
        const item = e.target.closest('.card-menu-item');
        if (item) {
            const cardEl = item.closest('.card');
            cardEl.classList.remove('menu-open');
            handleCardAction(item.getAttribute('data-action'), cardEl.getAttribute('data-hash'));
            return;
        }
        const cardEl = e.target.closest('.card');
        if (cardEl && !cardEl.classList.contains('menu-open')) {
            openInfoModal(cardEl.getAttribute('data-hash'));
        }
    });

    grid.addEventListener('mouseout', (e) => {
        const cardEl = e.target.closest('.card');
        if (!cardEl) return;
        if (!cardEl.contains(e.relatedTarget)) cardEl.classList.remove('menu-open');
    });
}