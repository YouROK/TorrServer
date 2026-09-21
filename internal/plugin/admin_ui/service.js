ts.web.staticDir("/css",  "css");
ts.web.staticDir("/js",   "js");
ts.web.staticDir("/img",  "img");
ts.web.staticDir("/i18n", "i18n");

function requireAdmin(req, res) {
    const from = encodeURIComponent(req.path);

    if (!req.user) {
        res.status(302).header("Location", "/login?from=" + from);
        return false;
    }

    if (req.user.rank < 50) {
        res.status(302).header("Location", "/login?error=forbidden&from=" + from);
        return false;
    }

    return true;
}

ts.web.get("/", function(req, res) {
    if (!requireAdmin(req, res)) {
        return;
    }
    res.html(ts.vfs.read("index.html"));
});

ts.web.get("/index.html", function(req, res) {
    if (!requireAdmin(req, res)) {
        return;
    }
    res.html(ts.vfs.read("index.html"));
});

ts.bus.on("system:boot", function(payload) {
    console.log("[Admin UI] system:boot received");
});

ts.bus.on("system:shutdown", function(payload) {
    console.log("[Admin UI] system:shutdown received");
});

console.log("[Admin UI] Started");