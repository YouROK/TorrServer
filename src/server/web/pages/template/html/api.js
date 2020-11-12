// Torrents
function addTorrent(link, title, poster, save, done, fail){
    torrent("add",link,null,title,poster,save,done,fail);
}

function getTorrent(hash, done, fail){
    torrent("get",null,hash,null,null,null,done,fail);
}

function remTorrent(hash, done, fail){
    torrent("rem",null,hash,null,null,null,done,fail);
}

function listTorrent(done, fail){
    torrent("list",null,null,null,null,null,done,fail);
}

function dropTorrent(hash, done, fail){
    torrent("drop",null,hash,null,null,null,done,fail);
}

function torrent(action, link, hash, title, poster, save, done, fail){
    var req = JSON.stringify({ action:action, link: link, title: title, poster: poster, save_to_db: save});
    $.post('/torrents',req)
        .done(function( data ) {
            if (done)
                done(data);
        })
        .fail(function( data ) {
            if (fail)
                fail(data);
        });
}
//

// Settings
function getSettings(done, fail){
    sendApi("get",null,"/settings",done,fail);
}

function setSettings(sets, done, fail){
    sendApi("set",sets,"/settings",done,fail);
}
//

// Viewed
function listViewed(done, fail){
    sendApi("list",null,"/viewed",done,fail);
}

function setViewed(hash, index, done, fail){
    var obj = {"hash":hash, "file_index":index};
    sendApi("set",obj,"/viewed",done,fail);
}

function remViewed(hash, index, done, fail){
    var obj = {"hash":hash, "file_index":index};
    sendApi("rem",obj,"/viewed",done,fail);
}
//

function sendApi(action, obj, path, done, fail){
    if (obj==null)
        obj={};
    obj[action]=action;
    var req = JSON.stringify(obj);
    $.post(path,req)
        .done(function( data ) {
            if (done)
                done(data);
        })
        .fail(function( data ) {
            if (fail)
                fail(data);
        });
}

function humanizeSize(size) {
    if (typeof size == 'undefined' || size == 0)
        return "";
    var i = Math.floor( Math.log(size) / Math.log(1024) );
    return ( size / Math.pow(1024, i) ).toFixed(2) * 1 + ' ' + ['B', 'kB', 'MB', 'GB', 'TB'][i];
}