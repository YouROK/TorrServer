// ═══════════════════════════════════════════════════════════════
// SERVICE.JS — Plugin Showcase: demonstrates ALL Silo plugin APIs
// ═══════════════════════════════════════════════════════════════
//
// HOW THIS FILE WORKS:
// 1. Executed once when Silo starts (not per-request)
// 2. ts.web.* calls register routes under /plugins/<plugin_id>/
// 3. ts.bus.on/* calls set up event listeners
// 4. ts.bus.emit() can be called from routes to trigger events
// 5. If a route/event is not in info.yaml → it will be DENIED
//
// ═══════════════════════════════════════════════════════════════


// ─── 1. Static File Serving ──────────────────────────────────
// Maps VFS paths to URL paths under /plugins/example-plugin/
// In HTML use RELATIVE paths: "css/style.css", "js/app.js"
// (NOT absolute "/css/style.css" — that goes to server root!)

ts.web.staticFile("/",           "index.html");
ts.web.staticFile("/index.html", "index.html");
ts.web.staticDir("/css",        "css");
ts.web.staticDir("/js",         "js");
ts.web.staticDir("/img",         "img");
ts.web.staticDir("/i18n",       "i18n");


// ─── 2. System Info (ts.info) ─────────────────────────────────
// Read-only. Available immediately. Use for logging or version display.

console.log("[Showcase] Silo " + ts.info.version + " | " + ts.info.os + "/" + ts.info.arch);
console.log("[Showcase] Plugin ID: " + ts.info.plugin_id);


// ─── 3. Storage (ts.storage) ────────────────────────────────────
// BoltDB key-value. Keys are prefixed with "example-plugin:" automatically.
// Strings only — use JSON.stringify/parse for objects.

var stored = ts.storage.get("showcase_message");
if (!stored) {
    stored = "Hello from first launch!";
    ts.storage.set("showcase_message", stored);
    console.log("[Showcase] Created default storage value");
} else {
    console.log("[Showcase] Loaded stored value: " + stored);
}


// ─── 4. Event Bus Event Log (for browser polling) ─────────────
// Store last 20 emitted events so the browser can fetch them via polling.
// This replaces SSE — no EventSource, no reconnecting.

var _eventLog = [];  // [{topic, payload, ts}, ...]

function _logEvent(topic, payload) {
    _eventLog.push({ topic: topic, payload: payload, ts: Date.now() });
    if (_eventLog.length > 20) _eventLog.shift();  // keep last 20
}


// ─── 5. Web Routes (ts.web) ────────────────────────────────────
// Register HTTP endpoints. Methods: get, post, put, patch, delete, head, options
// Also: ts.web.any() for all methods, ts.web.route(method, path, fn)
//
// req object:
//   req.method   — "GET", "POST", etc.
//   req.path     — URL path
//   req.query    — object with query params
//   req.headers  — lowercase header names as keys
//   req.body     — JSON object, form object, or raw string
//   req.user     — authenticated user or null
//
// res:
//   return obj               — auto JSON
//   return "string"         — text/plain
//   res.status(200).json({}) — explicit JSON
//   res.text("hello")
//   res.html("<b>bold</b>")

// GET /api/info — snapshot of ts.info
ts.web.get("/api/info", function(req) {
    return {
        plugin_id:        ts.info.plugin_id,
        silo_version:     ts.info.version,
        torrent_version:  ts.info.torrent_version,
        os:              ts.info.os,
        arch:            ts.info.arch,
        num_cpu:         ts.info.num_cpu
    };
});

// POST /api/echo — echo back the request body
ts.web.post("/api/echo", function(req) {
    return {
        received_method: req.method,
        received_path:   req.path,
        received_body:   req.body,
        received_query:  req.query,
        received_user:   req.user
    };
});

// POST /api/storage/get — read from ts.storage
ts.web.post("/api/storage/get", function(req) {
    var key = req.body.key || req.body;
    if (typeof key !== "string") key = String(key);
    var value = ts.storage.get(key);
    return { key: key, value: value || null };
});

// POST /api/storage/set — write to ts.storage
ts.web.post("/api/storage/set", function(req) {
    var key   = req.body.key;
    var value = req.body.value;
    if (!key) return { error: "key is required" };
    ts.storage.set(key, String(value || ""));
    return { ok: true, key: key, value: value };
});

// POST /api/storage/rem — remove from ts.storage
ts.web.post("/api/storage/rem", function(req) {
    var key = req.body.key || req.body;
    if (!key) return { error: "key is required" };
    ts.storage.rem(key);
    return { ok: true, key: key };
});

// POST /api/storage/clear — remove from ts.storage
ts.web.post("/api/storage/clear", function(req) {
    ts.storage.clear();
    return { ok: true };
});

// POST /api/emit — EMIT a custom event on the bus
// Stores the event in _eventLog for browser polling.
ts.web.post("/api/emit", function(req) {
    var topic   = req.body.topic   || "plugin:example-plugin:message";
    var payload = req.body.payload || req.body.message || "Hello from /api/emit!";
    ts.bus.emit(topic, payload);
    console.log("[Showcase] Emitted to " + topic + ": " + JSON.stringify(payload));
    return { emitted: true, topic: topic, payload: payload };
});

// GET /api/events — return event log for polling (instead of SSE)
ts.web.get("/api/events", function(req) {
    return { events: _eventLog };
});

// GET /api/http-demo — fetch external API via ts.http
ts.web.get("/api/http-demo", function(req) {
    var resp = ts.http.get("https://jsonplaceholder.typicode.com/posts/1");
    return {
        external_status: resp.status,
        external_data:   resp.json(),
        note:            "ts.http.get(url) or ts.http.post(url, body, opts)"
    };
});

// GET /api/torrent/list — list all torrents
ts.web.get("/api/torrent/list", function(req) {
    return { torrents: ts.torrent.list() };
});

// POST /api/torrent/add — add a torrent by magnet or hash
ts.web.post("/api/torrent/add", function(req) {
    var input = req.body.input || req.body;
    var opts  = req.body.opts  || {};
    if (typeof input !== "string") input = String(input);
    var status = ts.torrent.add(input, opts);
    return { added: true, status: status };
});

// GET /api/torrent/get — get one torrent by hash
ts.web.get("/api/torrent/get", function(req) {
    var hash = req.query.hash;
    if (!hash) return { error: "hash query param required" };
    return { torrent: ts.torrent.get(hash) };
});

// GET /api/users/list — list all users
ts.web.get("/api/users/list", function(req) {
    return { users: ts.users.list() };
});

// POST /api/users/get — get one user by ID
ts.web.post("/api/users/get", function(req) {
    var uid = req.body.id || req.body;
    return { user: ts.users.get(String(uid)) };
});


// ─── 6. Bus Subscriptions (ts.bus) ─────────────────────────────
// ts.bus.on(topic, callback) — subscribe to an event
//   • topic must be in info.yaml events[]
//   • callback receives payload
//   • returns unsubscribe() function
//
// ts.bus.emit(topic, payload) — publish an event
//
// Topic patterns:
//   "system:boot"    — exact match
//   "torrent:*"      — wildcard (all torrent:* topics)
//   "plugin:<id>:*"  — all events from a specific plugin

// system:boot — fires once when Silo finishes startup
ts.bus.on("system:boot", function(payload) {
    console.log("[Showcase] system:boot received");
    _logEvent("system:boot", payload || {});
});

// system:shutdown — fires before Silo stops
ts.bus.on("system:shutdown", function(payload) {
    console.log("[Showcase] system:shutdown received");
    _logEvent("system:shutdown", payload || {});
});

// torrent:metadata — fires when torrent metadata is loaded
ts.bus.on("torrent:metadata", function(payload) {
    console.log("[Showcase] torrent:metadata for " + payload.hash);
    _logEvent("torrent:metadata", payload);
});

// plugin:example-plugin:message — our own event, emitted from /api/emit
ts.bus.on("plugin:example-plugin:message", function(payload) {
    console.log("[Showcase] plugin:example-plugin:message received: " + JSON.stringify(payload));
    _logEvent("plugin:example-plugin:message", payload);
});

// Catch-all for torrent:* events
ts.bus.on("torrent:*", function(payload) {
    _logEvent("torrent:*", payload);
});


// ─── 7. Internationalization (ts.i18n) ─────────────────────────
// Plugins can add translations via:
//   1. Static: i18n/<lang>.json files (loaded automatically)
//   2. Dynamic: ts.i18n.add/addMany() in service.js
// Both merge into the global i18n registry served as /i18n.js

ts.i18n.lang("de", "Deutsch");

ts.i18n.add("en", "showcase_builtin_translation",
    "This string was added by service.js via ts.i18n.add()");
ts.i18n.add("ru", "showcase_builtin_translation",
    "Эта строка добавлена через ts.i18n.add() в service.js");

ts.i18n.addMany("en", {
    "showcase_i18n_dynamic":   "Dynamic translation from service.js (EN)",
    "showcase_i18n_dynamic_2": "Another dynamic EN string"
});
ts.i18n.addMany("ru", {
    "showcase_i18n_dynamic":   "Динамический перевод из service.js (RU)",
    "showcase_i18n_dynamic_2": "Еще одна динамическая строка (RU)"
});

// ─── 8. VFS (ts.vfs) ──────────────────────────────────────────
// Read-only доступ к файлам самого плагина через изолированную
// VFS (fs.FS). Плагин видит только свои файлы: распакованную папку
// или содержимое ZIP-архива. Пути — только относительные, прямые
// слэши. Ошибки чтения бросаются как JS-исключения.

// POST /api/vfs/list — список файлов в папке
ts.web.post("/api/vfs/list", function(req) {
    var path = req.body.path || ".";
    try {
        return { ok: true, path: path, entries: ts.vfs.list(path) };
    } catch (e) {
        return { ok: false, path: path, error: String(e) };
    }
});

// POST /api/vfs/stat — метаданные файла или папки
ts.web.post("/api/vfs/stat", function(req) {
    var path = req.body.path;
    if (!path) return { ok: false, error: "path is required" };
    try {
        return { ok: true, path: path, info: ts.vfs.stat(path) };
    } catch (e) {
        return { ok: false, path: path, error: String(e) };
    }
});

// POST /api/vfs/read — прочитать содержимое файла (как строку)
ts.web.post("/api/vfs/read", function(req) {
    var path = req.body.path;
    if (!path) return { ok: false, error: "path is required" };
    try {
        var content = ts.vfs.read(path);
        return { ok: true, path: path, size: content.length, content: content };
    } catch (e) {
        return { ok: false, path: path, error: String(e) };
    }
});

// POST /api/vfs/exists — есть ли файл (не бросает исключение)
ts.web.post("/api/vfs/exists", function(req) {
    var path = req.body.path;
    if (!path) return { ok: false, error: "path is required" };
    return { ok: true, path: path, exists: ts.vfs.exists(path) };
});


// ─── Startup: quick VFS sanity check ─────────────────────────
if (ts.vfs.exists("info.yaml")) {
    var st = ts.vfs.stat("info.yaml");
    console.log("[Showcase] VFS OK — info.yaml, " + st.size + " bytes, mode " + st.mode);
} else {
    console.log("[Showcase] VFS: info.yaml not found in plugin VFS!");
}

console.log("[Showcase] All routes, bus listeners, and i18n strings registered");