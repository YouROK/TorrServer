package template

const mainJS = `
$(document).ready(function() {
	watchInfo();
});

var lastTorrHtml = '';

function watchInfo(){
	reloadTorrents();
	setInterval(function() {
		
	}, 1000);
}

function reloadTorretns(){
	var torrents = $("#torrents");
	torrents.empty();
}

function loadTorrentInfoHtml(){

}

const torrElem = 


`

//Backbone.js
var torrElem = `"""
<div class="btn-group d-flex" role="group">
	<button type="button" class="btn btn-secondary wrap w-100" data-toggle="collapse" data-target="#info_'+tor.Hash+'"></button>';
	<a role="button" class="btn btn-secondary" href="'+tor.Playlist+'"><i class="fas fa-th-list"></i> Плейлист</a>';
	<button type="button" class="btn btn-secondary"><i class="fas fa-info"></i></a>';
	<button type="button" class="btn btn-secondary"><i class="fas fa-trash-alt"></i> Удалить</button>';
</div>
"""
`
