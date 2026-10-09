'use strict';

// Категории раздач: ключ уходит на сервер, подпись берется из словаря
const TORRENT_CATEGORIES = [
    { key: '', labelKey: 'category_none' },
    { key: 'movie', labelKey: 'category_movie' },
    { key: 'tv', labelKey: 'category_tv' },
    { key: 'music', labelKey: 'category_music' },
    { key: 'other', labelKey: 'category_other' },
];

// categoryLabel возвращает подпись категории по ключу из базы
function categoryLabel(key) {
    const found = TORRENT_CATEGORIES.find((c) => c.key === key);
    if (found) return t(found.labelKey);
    return key || t('category_none');
}

function categoryFillSelect(select, value) {
    if (!select) return;
    select.innerHTML = TORRENT_CATEGORIES
        .map((c) => '<option value="' + esc(c.key) + '">' + esc(t(c.labelKey)) + '</option>')
        .join('');
    select.value = TORRENT_CATEGORIES.some((c) => c.key === value) ? value : '';
}
