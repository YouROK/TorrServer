const defaultTrackers = [
    "http://retracker.local/announce",
    "http://bt4.t-ru.org/ann?magnet",
    "http://retracker.mgts.by:80/announce",
    "http://tracker.city9x.com:2710/announce",
    "http://tracker.electro-torrent.pl:80/announce",
    "http://tracker.internetwarriors.net:1337/announce",
    "http://tracker2.itzmx.com:6961/announce",
    "udp://opentor.org:2710",
    "udp://public.popcorn-tracker.org:6969/announce",
    "udp://tracker.opentrackr.org:1337/announce",
    "http://bt.svao-ix.ru/announce",
    "udp://explodie.org:6969/announce",
    "wss://tracker.btorrent.xyz",
    "wss://tracker.openwebtorrent.com"
];

const defaultConfig = {
    mode: "append",
    trackers: defaultTrackers
};

let currentConfig;
const savedConfigStr = ts.storage.get('config');

if (!savedConfigStr) {
    currentConfig = defaultConfig;
    ts.storage.set('config', JSON.stringify(currentConfig));
    ts.torrent.setTrackerPolicy(currentConfig);
    console.log("Initialized with default configuration");
} else {
    try {
        currentConfig = JSON.parse(savedConfigStr);
        ts.torrent.setTrackerPolicy(currentConfig);
        console.log("Loaded saved configuration");
    } catch (e) {
        currentConfig = defaultConfig;
        ts.storage.set('config', JSON.stringify(currentConfig));
        ts.torrent.setTrackerPolicy(currentConfig);
        console.log("Corrupted configuration detected, reset to default");
    }
}

ts.web.staticFile('/', 'index.html');
ts.web.staticFile('/index.html', 'index.html');
ts.web.staticDir('/js', 'js');

ts.web.get('/api/config', function(req, res) {
    if (req.user && req.user.rank < 100) {
        return res.status(403).json({ error: "forbidden" });
    }
    return res.json(currentConfig);
});

ts.web.post('/api/save', function(req, res) {
    if (req.user && req.user.rank < 100) {
        return res.status(403).json({ error: "forbidden" });
    }

    let body = req.body;
    if (typeof body === 'string') {
        try { body = JSON.parse(body); } catch (e) { body = {}; }
    }
    if (!body) body = {};

    const mode = String(body.mode || "append");
    const trackersRaw = String(body.trackers || "");

    const trackers = trackersRaw
        .split('\n')
        .map(function(t) { return t.trim(); })
        .filter(function(t) { return t.length > 0; });

    currentConfig = { mode: mode, trackers: trackers };

    ts.storage.set('config', JSON.stringify(currentConfig));
    ts.torrent.setTrackerPolicy(currentConfig);

    console.log("Configuration saved and applied to the torrent engine");
    return res.json({ status: "ok", config: currentConfig });
});