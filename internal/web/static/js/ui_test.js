'use strict';

// Проверки модуля точечного обновления.
// Запуск: node internal/web/static/js/ui_test.js

const assert = require('assert');
const path = require('path');

// Минимальная заглушка DOM: модулю нужны только те узлы, что он использует
class El {
    constructor(tag) {
        this.tagName = tag.toUpperCase();
        this.children = [];
        this.attributes = {};
        this._html = '';
        this.value = '';
        this.checked = false;
        this.id = '';
        this.type = '';
    }
    get innerHTML() { return this._html; }
    set innerHTML(v) { this._html = v; this.children = parseFields(v); }
    setAttribute(k, v) { this.attributes[k] = v; }
    getAttribute(k) { return this.attributes[k] || null; }
    appendChild(c) { this.children.push(c); return c; }
    closest(sel) {
        if (sel === '[data-row]') return this.isRow ? this : null;
        return null;
    }
    querySelector(sel) {
        for (const c of this.children) {
            if (sel.indexOf('data-row=') >= 0) {
                const key = sel.replace(/.*data-row="([^"]*)".*/, '$1');
                if (c.attributes['data-row'] === key) return c;
            }
            if (sel === 'button' && c.tagName === 'BUTTON') return c;
        }
        return null;
    }
    querySelectorAll(sel) {
        if (sel.indexOf('input') >= 0) return this.children.filter((c) => c.tagName === 'INPUT');
        if (sel.indexOf('button') >= 0) return this.children.filter((c) => c.tagName === 'BUTTON');
        return [];
    }
    get firstChild() { return this.children[0] || null; }
    insertBefore(node, ref) {
        const at = ref ? this.children.indexOf(ref) : -1;
        if (at < 0) this.children.push(node);
        else this.children.splice(at, 0, node);
        return node;
    }
    get nextSibling() {
        return null;
    }
    focus() { global.document.activeElement = this; }
    remove() { this.removed = true; }
    contains() { return false; }
    addEventListener() {}
}

// parseFields извлекает поля из простой разметки ввода
function parseFields(html) {
    const out = [];
    const re = /<(input|select|textarea)[^>]*id="([^"]*)"[^>]*>/g;
    let m;
    while ((m = re.exec(html)) !== null) {
        const el = new El(m[1]);
        el.id = m[2];
        out.push(el);
    }
    return out;
}

global.window = global;
global.document = {
    getElementById: () => null,
    activeElement: null,
    createElement: (t) => new El(t),
    querySelectorAll: () => [],
};

require(path.join(__dirname, 'ui.js'));

const ui = global.siloUI;
assert.ok(ui, 'siloUI must be exported');

// render сохраняет значения полей
{
    const box = new El('div');
    box.innerHTML = '<input id="a"><input id="b">';
    box.querySelectorAll = (sel) => box.children;
    box.children[0].value = 'kept';

    const plain = new El('div');
    plain.querySelectorAll = () => [];
    global.document.getElementById = () => box;

    ui.render(box, '<input id="a"><input id="b">');
    assert.strictEqual(box.children[0].value, 'kept', 'значение поля должно сохраниться');
}

// on использует делегирование
{
    const box = new El('div');
    let called = false;
    box.addEventListener = (event, fn) => {
        assert.strictEqual(event, 'click');
        called = true;
    };
    ui.on(box, 'click', '[data-x]', () => {});
    assert.ok(called, 'обработчик должен подписаться на контейнер');
}

// timer перезапускает обновление под тем же ключом
{
    ui.timer('t', () => {}, 1000);
    const first = ui.timers.t;
    ui.timer('t', () => {}, 1000);
    assert.notStrictEqual(ui.timers.t, first, 'таймер должен перезапускаться');
    ui.stopAll();
    assert.deepStrictEqual(ui.timers, {}, 'stopAll должен очистить таймеры');
}

// text не трогает узел без изменений
{
    const el = new El('span');
    el.textContent = 'same';
    ui.text(el, 'same');
    assert.strictEqual(el.textContent, 'same');
    ui.text(el, 'new');
    assert.strictEqual(el.textContent, 'new');
}

console.log('ui.js: all checks passed');

// --- Проверки построчного обновления ---

// makeTbody собирает минимальный tbody со связями между узлами
function makeTbody() {
    const tbody = new El('tbody');
    tbody._kids = [];

    Object.defineProperty(tbody, 'children', {
        get() { return tbody._kids; },
        set(v) { tbody._kids = v; v.forEach((r) => { r.parent = tbody; }); },
    });
    Object.defineProperty(tbody, 'firstChild', { get() { return tbody._kids[0] || null; } });

    tbody.insertBefore = function (node, ref) {
        const kids = this._kids;
        const i = kids.indexOf(node);
        if (i >= 0) kids.splice(i, 1);
        const at = ref ? kids.indexOf(ref) : -1;
        if (at < 0) kids.push(node);
        else kids.splice(at, 0, node);
        node.parent = this;
        return node;
    };
    tbody.querySelector = function (sel) {
        const key = sel.replace(/.*data-row="([^"]*)".*/, '$1');
        return this._kids.find((c) => c.attributes['data-row'] === key) || null;
    };
    return tbody;
}

// previousSibling вычисляется по родителю
El.prototype.remove = function () {
    if (!this.parent) return;
    const i = this.parent._kids.indexOf(this);
    if (i >= 0) this.parent._kids.splice(i, 1);
    this.parent = null;
};
El.prototype.__defineGetter__('previousSibling', function () {
    if (!this.parent) return null;
    const i = this.parent._kids.indexOf(this);
    return i > 0 ? this.parent._kids[i - 1] : null;
});
El.prototype.__defineGetter__('nextSibling', function () {
    if (!this.parent) return null;
    const i = this.parent._kids.indexOf(this);
    return i >= 0 && i + 1 < this.parent._kids.length ? this.parent._kids[i + 1] : null;
});

{
    const tbody = makeTbody();
    const items = [{ id: 'a', text: 'A' }, { id: 'b', text: 'B' }];
    const opts = {
        key: (x) => x.id,
        html: (x) => '<td>' + x.text + '</td>',
    };

    ui.rows(tbody, items, opts);
    assert.strictEqual(tbody.children.length, 2, 'должны появиться две строки');
    assert.strictEqual(tbody.children[0].attributes['data-row'], 'a');

    const nodeA = tbody.children[0];
    const nodeB = tbody.children[1];

    // Без изменений узлы переиспользуются
    ui.rows(tbody, items, opts);
    assert.strictEqual(tbody.children[0], nodeA, 'строка без изменений переиспользуется');
    assert.strictEqual(tbody.children[1], nodeB, 'строка без изменений переиспользуется');

    // Меняется только изменившаяся строка
    items[1].text = 'B2';
    ui.rows(tbody, items, opts);
    assert.strictEqual(tbody.children[0], nodeA, 'соседняя строка не пересоздаётся');
    assert.ok(tbody.children[1]._html.indexOf('B2') >= 0, 'строка обновилась');

    // Лишняя строка удаляется
    ui.rows(tbody, [items[0]], opts);
    assert.strictEqual(tbody.children.length, 1, 'лишняя строка удаляется');

    // Порядок меняется
    ui.rows(tbody, [items[1], items[0]], opts);
    assert.strictEqual(tbody.children[0].attributes['data-row'], 'b', 'порядок изменился');
    assert.strictEqual(tbody.children[1].attributes['data-row'], 'a', 'порядок изменился');
}

console.log('ui.js rows: all checks passed');
