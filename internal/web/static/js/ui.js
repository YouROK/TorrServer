'use strict';

// Модуль точечного обновления интерфейса.
// Позволяет перерисовывать отдельные области, не трогая остальную панель:
// ввод, фокус и позиция курсора сохраняются.
(function (global) {
    // fieldKey возвращает устойчивый ключ поля.
    function fieldKey(el, index) {
        return el.id || el.getAttribute('name') || el.getAttribute('data-key') || ('i' + index);
    }

    // isField проверяет, что узел принимает ввод.
    function isField(el) {
        const tag = el.tagName;
        return tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA';
    }

    // focusableSelector выбирает узлы, которые могут получить фокус.
    const focusableSelector = 'button, input, select, textarea, a[href]';

    // focusSignature описывает, где стоит фокус.
    // Запоминается строка и позиция узла внутри неё.
    function focusSignature(host) {
        const active = document.activeElement;
        if (!active || !host.contains(active)) return null;

        const row = active.closest('[data-row]');
        const scope = row || host;
        const all = Array.from(scope.querySelectorAll(focusableSelector));

        return {
            row: row ? row.getAttribute('data-row') : null,
            index: all.indexOf(active),
        };
    }

    // applyAttrs приводит атрибуты строки к заданному набору.
    // Атрибут, которого нет в наборе или который равен null, снимается.
    function applyAttrs(row, attrs) {
        const next = attrs || {};
        const prev = row.siloAttrs || {};

        Object.keys(prev).forEach((name) => {
            if (!(name in next)) row.removeAttribute(name);
        });

        Object.keys(next).forEach((name) => {
            const value = next[name];
            if (value === null || value === undefined) {
                row.removeAttribute(name);
            } else {
                row.setAttribute(name, value);
            }
        });

        row.siloAttrs = next;
    }

    // restoreFocus возвращает фокус на прежнее место.
    function restoreFocus(host, mark) {
        if (!mark || mark.index < 0) return;

        let scope = host;
        if (mark.row !== null) {
            scope = host.querySelector('[data-row="' + mark.row + '"]');
            if (!scope) return;
        }

        const all = Array.from(scope.querySelectorAll(focusableSelector));
        const target = all[mark.index];
        if (target && target.focus) target.focus();
    }

    const siloUI = {
        // region возвращает именованную область внутри контейнера.
        // Область создаётся один раз и переживает перерисовку соседей.
        region(container, name) {
            const host = typeof container === 'string' ? document.getElementById(container) : container;
            if (!host) return null;

            let el = host.querySelector('[data-region="' + name + '"]');
            if (!el) {
                el = document.createElement('div');
                el.setAttribute('data-region', name);
                host.appendChild(el);
            }
            return el;
        },

        // snapshot собирает значения полей области.
        snapshot(container) {
            const state = {};
            const fields = container.querySelectorAll('input, select, textarea');

            fields.forEach(function (el, index) {
                const key = fieldKey(el, index);
                if (el.type === 'checkbox' || el.type === 'radio') {
                    state[key] = el.checked;
                } else {
                    state[key] = el.value;
                }
            });
            return state;
        },

        // restore возвращает сохранённые значения полей.
        restore(container, state) {
            if (!state) return;

            const fields = container.querySelectorAll('input, select, textarea');
            fields.forEach(function (el, index) {
                const key = fieldKey(el, index);
                if (!(key in state)) return;

                if (el.type === 'checkbox' || el.type === 'radio') {
                    el.checked = state[key];
                } else if (el.value !== state[key]) {
                    el.value = state[key];
                }
            });
        },

        // render заменяет содержимое области, сохраняя ввод и фокус.
        // Возвращает саму область, чтобы сразу повесить обработчики.
        render(target, html) {
            const el = typeof target === 'string' ? document.getElementById(target) : target;
            if (!el) return null;

            const state = this.snapshot(el);

            // Фокус восстанавливается по тому же ключу поля
            const active = document.activeElement;
            let focusKey = null;
            if (active && el.contains(active) && isField(active)) {
                const siblings = Array.from(el.querySelectorAll('input, select, textarea'));
                focusKey = fieldKey(active, siblings.indexOf(active));
            }

            el.innerHTML = html;
            this.restore(el, state);

            if (focusKey) {
                const fields = Array.from(el.querySelectorAll('input, select, textarea'));
                for (let i = 0; i < fields.length; i++) {
                    if (fieldKey(fields[i], i) === focusKey) {
                        fields[i].focus();
                        break;
                    }
                }
            }
            return el;
        },

        // text обновляет текст узла, только если он изменился.
        text(target, value) {
            const el = typeof target === 'string' ? document.getElementById(target) : target;
            if (!el) return;

            const next = String(value);
            if (el.textContent !== next) el.textContent = next;
        },

        // on вешает обработчик через делегирование.
        // Такой обработчик переживает перерисовку области.
        on(container, event, selector, handler) {
            const host = typeof container === 'string' ? document.getElementById(container) : container;
            if (!host) return;

            host.addEventListener(event, function (e) {
                const match = e.target.closest(selector);
                if (match && host.contains(match)) handler(e, match);
            });
        },

        // labels подписывает ячейки строки заголовками колонок.
        // На узком экране таблица становится карточками, и подпись
        // показывает, к какой колонке относится значение.
        labels(row) {
            if (!row || !row.cells || typeof row.closest !== 'function') return;

            const table = row.closest('table');
            const head = table && table.tHead && table.tHead.rows[0];
            if (!head || row.cells.length !== head.cells.length) return;

            Array.from(row.cells).forEach((cell, index) => {
                const label = head.cells[index].textContent.trim();
                if (label) {
                    cell.setAttribute('data-label', label);
                } else {
                    cell.removeAttribute('data-label');
                }
            });
        },

        // plaques заполняет плашки парами [подпись, значение].
        // Разметка строится один раз, дальше меняются только изменившиеся значения.
        plaques(target, pairs) {
            const host = typeof target === 'string' ? document.getElementById(target) : target;
            if (!host) return null;

            const items = pairs || [];
            let values = Array.from(host.querySelectorAll('.plaque-value'));
            const labels = Array.from(host.querySelectorAll('.plaque-label'));
            const rebuild = values.length !== items.length || labels.length !== items.length ||
                labels.some((el, i) => el.textContent !== String(items[i][0]));

            if (rebuild) {
                host.innerHTML = items.map(() =>
                    '<div class="plaque"><span class="plaque-label"></span><span class="plaque-value"></span></div>'
                ).join('');
                values = Array.from(host.querySelectorAll('.plaque-value'));
                Array.from(host.querySelectorAll('.plaque-label'))
                    .forEach((el, i) => { el.textContent = String(items[i][0]); });
            }

            values.forEach((el, i) => this.text(el, items[i][1]));
            return host;
        },

        // items обновляет элементы списка по ключу.
        // Неизменившиеся элементы не трогаются, поэтому фокус и выделение в них живут.
        // options.key возвращает ключ элемента, options.html - его разметку,
        // options.tag - тег (по умолчанию div).
        // options.attrs возвращает атрибуты самого элемента (например class),
        // их нужно задавать здесь: узел создаёт модуль, а не вызывающий код.
        items(target, list, options) {
            const host = typeof target === 'string' ? document.getElementById(target) : target;
            if (!host) return null;

            const opts = options || {};
            const tag = opts.tag || 'div';
            const keyOf = opts.key || ((item) => item.id);
            const htmlOf = opts.html || ((item) => String(item));
            const attrsOf = opts.attrs || null;

            const mark = focusSignature(host);
            const existing = new Map();
            Array.from(host.children).forEach((el) => {
                const key = el.getAttribute('data-row');
                if (key !== null) existing.set(key, el);
            });

            const seen = new Set();
            let previous = null;

            (list || []).forEach((item) => {
                const key = String(keyOf(item));
                const html = htmlOf(item);
                const attrs = attrsOf ? attrsOf(item) : null;
                const attrsJSON = attrs ? JSON.stringify(attrs) : '';
                seen.add(key);

                let el = existing.get(key);
                if (!el) {
                    el = document.createElement(tag);
                    el.setAttribute('data-row', key);
                    el.siloHtml = html;
                    el.siloAttrsJSON = attrsJSON;
                    applyAttrs(el, attrs);
                    el.innerHTML = html;
                } else {
                    if (el.siloAttrsJSON !== attrsJSON) {
                        el.siloAttrsJSON = attrsJSON;
                        applyAttrs(el, attrs);
                    }
                    if (el.siloHtml !== html) {
                        // Сравнивается исходная разметка, а не innerHTML:
                        // браузер нормализует атрибуты и сравнение всегда давало бы различие
                        el.siloHtml = html;
                        el.innerHTML = html;
                    }
                }

                // Узел перемещается, только если стоит не на своём месте.
                // Лишние перемещения сбрасывают фокус и выделение внутри элемента
                const at = previous ? previous.nextSibling : host.firstChild;
                if (el !== at) host.insertBefore(el, at);
                previous = el;
            });

            existing.forEach((el, key) => {
                if (!seen.has(key)) el.remove();
            });

            restoreFocus(host, mark);
            return host;
        },

        // rows обновляет строки таблицы по ключу (см. items).
        // Подписи ячеек зависят от заголовков таблицы, поэтому ставятся здесь.
        rows(target, list, options) {
            const host = this.items(target, list, Object.assign({}, options, { tag: 'tr' }));
            if (!host) return null;

            Array.from(host.children).forEach((row) => this.labels(row));
            return host;
        },

        // timer запускает периодическое обновление области и хранит его ключ.
        timer(key, fn, interval) {
            this.timers = this.timers || {};

            if (this.timers[key]) clearInterval(this.timers[key]);
            fn();
            this.timers[key] = setInterval(fn, interval);
        },

        // stopTimer останавливает обновление области.
        stopTimer(key) {
            if (this.timers && this.timers[key]) {
                clearInterval(this.timers[key]);
                delete this.timers[key];
            }
        },

        // stopAll останавливает все обновления.
        stopAll() {
            if (!this.timers) return;
            Object.keys(this.timers).forEach((key) => clearInterval(this.timers[key]));
            this.timers = {};
        },
    };

    global.siloUI = siloUI;
})(window);
