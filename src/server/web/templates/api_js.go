package templates

import (
	"net/http"

	"server/settings"
	"server/web/helpers"

	"github.com/labstack/echo"
)

var apijs = `
function addTorrent(link, save, info, done, fail){
	var reqJson = JSON.stringify({ Link: link, Info: info, DontSave: !save});
	$.post('/torrent/add',reqJson)
	.done(function( data ) {
		if (done)
			done(data);
	})
	.fail(function( data ) {
		if (fail)
			fail(data);
	});
}

function getTorrent(hash, done, fail){
	var reqJson = JSON.stringify({ Hash: hash});
	$.post('/torrent/get',reqJson)
	.done(function( data ) {
		if (done)
			done(data);
	})
	.fail(function( data ) {
		if (fail)
			fail(data);
	});
}

function removeTorrent(hash, done, fail){
	var reqJson = JSON.stringify({ Hash: hash});
	$.post('/torrent/rem',reqJson)
	.done(function( data ) {
		if (done)
			done(data);
	})
	.fail(function( data ) {
		if (fail)
			fail(data);
	});
}

function statTorrent(hash, done, fail){
	var reqJson = JSON.stringify({ Hash: hash});
	$.post('/torrent/stat',reqJson)
	.done(function( data ) {
		if (done)
			done(data);
	})
	.fail(function( data ) {
		if (fail)
			fail(data);
	});
}

function cacheTorrent(hash, done, fail){
	var reqJson = JSON.stringify({ Hash: hash});
	$.post('/torrent/cache',reqJson)
	.done(function( data ) {
		if (done)
			done(data);
	})
	.fail(function( data ) {
		if (fail)
			fail(data);
	});
}

function listTorrent(done, fail){
	$.post('/torrent/list')
	.done(function( data ) {
		if (done)
			done(data);
	})
	.fail(function( data ) {
		if (fail)
			fail(data);
	});
}

function restartService(done, fail){
	$.get('/torrent/restart')
	.done(function( data ) {
		if (done)
			done();
	})
	.fail(function( data ) {
		if (fail)
			fail(data);
	});
}

function preloadTorrent(preloadLink, done, fail){
	$.get(preloadLink)
	.done(function( data ) {
		if (done)
			done();
	})
	.fail(function( data ) {
		if (fail)
			fail(data);
	});
}

function shutdownServer(fail){
	$.post('/shutdown')
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
`

func Api_JS(c echo.Context) error {
	http.ServeContent(c.Response(), c.Request(), "api.js", settings.StartTime, helpers.NewSeekingBuffer(apijs))
	return c.NoContent(http.StatusOK)
}
