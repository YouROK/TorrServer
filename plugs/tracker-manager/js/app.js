'use strict';

const t = (key, def) => i18n.t(key, def);

function toast(msg) {
    const el = document.getElementById('toast');
    if (!el) return;
    el.textContent = msg;
    el.classList.add('show');
    setTimeout(() => el.classList.remove('show'), 2500);
}

document.addEventListener('DOMContentLoaded', () => {
    const policySelect = document.getElementById('policy');
    const trackersTextarea = document.getElementById('trackers');
    const btnSave = document.getElementById('btn-save');
    const btnDefault = document.getElementById('btn-default');

    const DEFAULT_TRACKERS = [
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

    async function loadConfig() {
        try {
            const res = await fetch('/plugins/tracker-manager/api/config');
            if (!res.ok) throw new Error('Network response was not ok');
            const data = await res.json();

            if (data.error) {
                toast(t('plugin.tracker-manager.error_load') + " (" + data.error + ")");
                return;
            }

            policySelect.value = data.mode || 'append';
            trackersTextarea.value = (data.trackers || []).join('\n');
        } catch (err) {
            console.error('Load config error:', err);
            toast(t('plugin.tracker-manager.error_load'));
        }
    }

    async function saveConfig(mode, trackersText) {
        try {
            btnSave.disabled = true;
            btnDefault.disabled = true;

            const res = await fetch('/plugins/tracker-manager/api/save', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ mode: mode, trackers: trackersText })
            });

            const data = await res.json();
            if (data.error) {
                toast(t('plugin.tracker-manager.error_save') + " (" + data.error + ")");
                return;
            }

            if (data.config) {
                policySelect.value = data.config.mode;
                trackersTextarea.value = data.config.trackers.join('\n');
            }

            toast(t('plugin.tracker-manager.saved_success'));
        } catch (err) {
            console.error('Save config error:', err);
            toast(t('plugin.tracker-manager.error_save'));
        } finally {
            btnSave.disabled = false;
            btnDefault.disabled = false;
        }
    }

    btnSave.addEventListener('click', () => {
        saveConfig(policySelect.value, trackersTextarea.value);
    });

    btnDefault.addEventListener('click', () => {
        const defaultText = DEFAULT_TRACKERS.join('\n');
        policySelect.value = 'append';
        trackersTextarea.value = defaultText;

        saveConfig('append', defaultText);
    });

    i18n.apply();

    loadConfig();
});