// ===============================================================
// SERVICE.JS - Plugin Showcase: demonstrates ALL Silo plugin APIs
// ===============================================================
//
// HOW THIS FILE WORKS:
//   1. Executed once when Silo starts (not per-request)
//   2. ts.web.* calls register routes under /plugins/<plugin_id>/
//   3. ts.bus.on/* calls set up event listeners
//   4. ts.bus.emit() can be called from routes to trigger events
//   5. If a route/event is not in info.yaml, it will be denied
//
// ===============================================================

// --- 1. Static File Serving -----------------------------------
// Maps VFS paths to URL paths under /plugins/example-plugin/.
// In HTML use RELATIVE paths: "css/style.css", "js/app.js"
// (NOT absolute "/css/style.css", that goes to server root).
//
// Note: no need to register /i18n. The core reads i18n/*.json from the
// plugin VFS automatically at load time and merges it into /i18n.js.

ts.web.staticFile("/", "index.html");
ts.web.staticFile("/index.html", "index.html");
ts.web.staticDir("/css", "css");
ts.web.staticDir("/js", "js");
ts.web.staticDir("/img", "img");

// --- 2. System Info (ts.info) ---------------------------------
// Read-only. Available immediately. Use for logging or version display.

console.log("[Showcase] Silo " + ts.info.version + " | " + ts.info.os + "/" + ts.info.arch);
console.log("[Showcase] Plugin ID: " + ts.info.plugin_id);

// --- 3. Storage (ts.storage) ----------------------------------
// BoltDB key-value. Keys are prefixed with "example-plugin:" automatically.
// Strings only, use JSON.stringify/parse for objects.

var stored = ts.storage.get("showcase_message");
if (!stored) {
    stored = "Hello from first launch!";
    ts.storage.set("showcase_message", stored);
    console.log("[Showcase] Created default storage value");
} else {
    console.log("[Showcase] Loaded stored value: " + stored);
}

// --- 4. Event Bus Event Log -----------------------------------
// Store last 20 emitted events so the browser can fetch them via polling.
// This replaces SSE for the demo.

var _eventLog = [];

function _logEvent(topic, payload) {
    _eventLog.push({ topic: topic, payload: payload, ts: Date.now() });
    if (_eventLog.length > 20) _eventLog.shift();
}

// --- 5. Web Routes (ts.web) -----------------------------------
// Register HTTP endpoints.
//
// Request object (first arg):
//   req.method   - "GET", "POST", etc.
//   req.path     - URL path
//   req.query    - object with query params
//   req.headers  - lowercase header names as keys
//   req.body     - JSON object, form object, or raw string
//   req.user     - authenticated user or null
//
// Response:
//   return obj                  -> auto JSON
//   return "string"             -> text/plain
//   res.status(200).json({...}) -> explicit JSON
//   res.text("hello")           -> text/plain
//   res.html("<b>bold</b>")     -> text/html
//   res.stream(handle, name)    -> file stream (from ts.torrfs.open)
//   res.end()                   -> no body, only headers (for HEAD)

// GET /api/info - snapshot of ts.info
ts.web.get("/api/info", function(req) {
    return {
        plugin_id:       ts.info.plugin_id,
        silo_version:    ts.info.version,
        torrent_version: ts.info.torrent_version,
        os:              ts.info.os,
        arch:            ts.info.arch,
        num_cpu:         ts.info.num_cpu
    };
});

// POST /api/echo - echo back the request body
ts.web.post("/api/echo", function(req) {
    return {
        received_method: req.method,
        received_path:   req.path,
        received_body:   req.body,
        received_query:  req.query,
        received_user:   req.user
    };
});

// POST /api/storage/get - read from ts.storage
ts.web.post("/api/storage/get", function(req) {
    var key = req.body.key || req.body;
    if (typeof key !== "string") key = String(key);
    var value = ts.storage.get(key);
    return { key: key, value: value || null };
});

// POST /api/storage/set - write to ts.storage
ts.web.post("/api/storage/set", function(req) {
    var key   = req.body.key;
    var value = req.body.value;
    if (!key) return { error: "key is required" };
    ts.storage.set(key, String(value || ""));
    return { ok: true, key: key, value: value };
});

// POST /api/storage/rem - remove from ts.storage
ts.web.post("/api/storage/rem", function(req) {
    var key = req.body.key || req.body;
    if (!key) return { error: "key is required" };
    ts.storage.rem(key);
    return { ok: true, key: key };
});

// POST /api/storage/clear - wipe all keys of this plugin
ts.web.post("/api/storage/clear", function(req) {
    ts.storage.clear();
    return { ok: true };
});

// POST /api/emit - emit a custom event on the bus
ts.web.post("/api/emit", function(req) {
    var topic   = req.body.topic   || "plugin:example-plugin:message";
    var payload = req.body.payload || req.body.message || "Hello from /api/emit!";
    ts.bus.emit(topic, payload);
    console.log("[Showcase] Emitted to " + topic + ": " + JSON.stringify(payload));
    return { emitted: true, topic: topic, payload: payload };
});

// GET /api/events - return event log for polling
ts.web.get("/api/events", function(req) {
    return { events: _eventLog };
});

// GET /api/http-demo - fetch external API via ts.http
ts.web.get("/api/http-demo", function(req) {
    var resp = ts.http.get("https://jsonplaceholder.typicode.com/posts/1");
    return {
        external_status: resp.status,
        external_data:   resp.json(),
        note:            "ts.http.get(url) or ts.http.post(url, body, opts)"
    };
});

// GET /api/torrent/list - list all torrents
ts.web.get("/api/torrent/list", function(req) {
    return { torrents: ts.torrent.list() };
});

// POST /api/torrent/add - add a torrent by magnet or hash
ts.web.post("/api/torrent/add", function(req) {
    var input = req.body.input || req.body;
    var opts  = req.body.opts  || {};
    if (typeof input !== "string") input = String(input);
    var status = ts.torrent.add(input, opts);
    return { added: true, status: status };
});

// GET /api/torrent/get - get one torrent by hash
ts.web.get("/api/torrent/get", function(req) {
    var hash = req.query.hash;
    if (!hash) return { error: "hash query param required" };
    return { torrent: ts.torrent.get(hash) };
});

// GET /api/users/list - list all users
ts.web.get("/api/users/list", function(req) {
    return { users: ts.users.list() };
});

// POST /api/users/get - get one user by ID
ts.web.post("/api/users/get", function(req) {
    var uid = req.body.id || req.body;
    return { user: ts.users.get(String(uid)) };
});

// ===============================================================
// ts.crypto - cryptographic primitives
// ===============================================================
//
// All functions are synchronous and return strings (hex or base64).
// Byte arrays are not used: JS does not have a native byte type, and
// working with Uint8Array is awkward for plugin authors.
//
// Functions:
//   ts.crypto.sha256(text)                          -> hex
//   ts.crypto.sha1(text)                            -> hex
//   ts.crypto.md5(text)                             -> hex
//   ts.crypto.hmacSha256(key, text)                 -> hex
//   ts.crypto.pbkdf2(password, salt, iter, len)     -> hex
//   ts.crypto.bcrypt(password, [cost=10])           -> bcrypt hash string
//   ts.crypto.bcryptVerify(password, hash)          -> bool
//   ts.crypto.randomBytes(n)                        -> hex (2n chars)
//   ts.crypto.randomBase64(n)                       -> url-safe base64
//   ts.crypto.uuid()                                -> UUID v4
//   ts.crypto.base64encode(text) / base64decode     -> string
//   ts.crypto.hexEncode(text) / hexDecode           -> string
//   ts.crypto.aesGcmEncrypt(key, text)              -> base64
//   ts.crypto.aesGcmDecrypt(key, ciphertext)        -> string
//   ts.crypto.constantTimeCompare(a, b)             -> bool
//
// Parameter limits enforced by the core:
//   pbkdf2: iterations 1000..10_000_000, keyLen 16..1024
//   bcrypt: cost 4..14 (default 10)
//   randomBytes/randomBase64: 1..1_048_576 bytes
//   aesGcmEncrypt/Decrypt: any string key, sha256 of it = 32-byte AES-256 key

ts.web.post("/api/crypto/hash", function(req) {
    var text = String(req.body.text || "");
    return {
        input: text,
        sha256: ts.crypto.sha256(text),
        sha1: ts.crypto.sha1(text),
        md5: ts.crypto.md5(text),
        hmac_sha256: ts.crypto.hmacSha256("demo-key", text)
    };
});

ts.web.post("/api/crypto/kdf", function(req) {
    var password = String(req.body.password || "");
    var iterations = parseInt(req.body.iterations, 10) || 100000;
    var salt = ts.crypto.randomBytes(16);
    var hash = ts.crypto.pbkdf2(password, salt, iterations, 32);
    return { password: password, salt: salt, iterations: iterations, hash: hash };
});

ts.web.post("/api/crypto/bcrypt", function(req) {
    var password = String(req.body.password || "");
    return { password: password, hash: ts.crypto.bcrypt(password, 10) };
});

ts.web.post("/api/crypto/bcrypt-verify", function(req) {
    var password = String(req.body.password || "");
    var hash = String(req.body.hash || "");
    return { password: password, hash: hash, valid: ts.crypto.bcryptVerify(password, hash) };
});

ts.web.post("/api/crypto/random", function(req) {
    var size = parseInt(req.body.size, 10) || 16;
    return {
        hex: ts.crypto.randomBytes(size),
        base64: ts.crypto.randomBase64(size),
        uuid: ts.crypto.uuid()
    };
});

ts.web.post("/api/crypto/encode", function(req) {
    var text = String(req.body.text || "");
    return {
        input: text,
        base64: ts.crypto.base64encode(text),
        hex: ts.crypto.hexEncode(text)
    };
});

ts.web.post("/api/crypto/aes", function(req) {
    var mode = String(req.body.mode || "");
    var key = String(req.body.key || "");
    var data = String(req.body.data || "");
    if (mode === "encrypt") {
        return { mode: mode, key: key, input: data, output: ts.crypto.aesGcmEncrypt(key, data) };
    }
    if (mode === "decrypt") {
        return { mode: mode, key: key, input: data, output: ts.crypto.aesGcmDecrypt(key, data) };
    }
    return { error: "mode must be 'encrypt' or 'decrypt'" };
});

ts.web.post("/api/crypto/compare", function(req) {
    var a = String(req.body.a || "");
    var b = String(req.body.b || "");
    return { a: a, b: b, equal: ts.crypto.constantTimeCompare(a, b) };
});

// ===============================================================
// ts.torrfs - virtual filesystem over the user torrent library
// ===============================================================
//
// The tree is built from the database. The torrent engine is not touched
// until open(). Path format: "Category/TorrentName/folder/file.mkv" or ""
// for the root.
// If the user has 0 or 1 categories, torrents live at the root.
// If the user has 2 or more categories, the root contains categories.
//
// Node object:
//   {
//     name:    string,   // name without path
//     path:    string,   // full path from root
//     is_dir:  bool,
//     size:    int64,    // 0 for directories
//     mtime:   int64,    // unix seconds
//     mime:    string    // files only
//   }
//
// Functions:
//   ts.torrfs.list(userID, path)  -> array of nodes or null
//   ts.torrfs.stat(userID, path)  -> node or null
//   ts.torrfs.open(userID, path)  -> handle (opaque object)
//   ts.torrfs.close(handle)       -> close handle manually
//
// A handle is passed to res.stream; the core closes it automatically after
// the response finishes. If a handle was opened but never streamed, the
// core closes it at the end of the request. Explicit close is only needed
// when you want to release the reader earlier.

ts.web.post("/api/torrfs/list", function(req) {
    if (!req.user) return { error: "authentication required" };
    var path = String(req.body.path || "");
    var nodes = ts.torrfs.list(req.user.id, path);
    return { path: path, nodes: nodes || [] };
});

ts.web.post("/api/torrfs/stat", function(req) {
    if (!req.user) return { error: "authentication required" };
    var path = String(req.body.path || "");
    var node = ts.torrfs.stat(req.user.id, path);
    return { path: path, node: node || null };
});

ts.web.get("/api/torrfs/stream", function(req, res) {
    if (!req.user) return res.status(401).text("authentication required");
    var path = String(req.query.path || "");
    if (!path) return res.status(400).text("path is required");
    var node = ts.torrfs.stat(req.user.id, path);
    if (!node) return res.status(404).text("not found");
    if (node.is_dir) return res.status(405).text("is a directory");
    var h = ts.torrfs.open(req.user.id, path);
    if (!h) return res.status(404).text("not found");
    res.header("Content-Type", node.mime || "application/octet-stream");
    res.stream(h, node.name);
});

// --- 6. Bus Subscriptions (ts.bus) ----------------------------
// ts.bus.on(topic, callback) - subscribe to an event.
//   topic must be in info.yaml events[]
//   callback receives payload
//   returns unsubscribe() function
//
// Topic patterns:
//   "system:boot"    - exact match
//   "torrent:*"      - wildcard (all torrent:* topics)

ts.bus.on("system:boot", function(payload) {
    console.log("[Showcase] system:boot received");
    _logEvent("system:boot", payload || {});
});

ts.bus.on("system:shutdown", function(payload) {
    console.log("[Showcase] system:shutdown received");
    _logEvent("system:shutdown", payload || {});
});

ts.bus.on("torrent:metadata", function(payload) {
    console.log("[Showcase] torrent:metadata for " + payload.hash);
    _logEvent("torrent:metadata", payload);
});

ts.bus.on("plugin:example-plugin:message", function(payload) {
    console.log("[Showcase] plugin:example-plugin:message received: " + JSON.stringify(payload));
    _logEvent("plugin:example-plugin:message", payload);
});

ts.bus.on("torrent:*", function(payload) {
    _logEvent("torrent:*", payload);
});

// --- 7. Internationalization (ts.i18n) ------------------------
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

console.log("[Showcase] All routes, bus listeners, and i18n strings registered");