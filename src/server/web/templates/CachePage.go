package templates

import (
	"server/version"
)

var cachePage = `
<!DOCTYPE html>
<html lang="ru">
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="` + faviconB64 + `" rel="icon" type="image/x-icon">
		<script src="/js/api.js"></script>
		<link rel="stylesheet" href="https://use.fontawesome.com/releases/v5.1.0/css/all.css" integrity="sha384-lKuwvrZot6UHsBSfcMvOkWwlCMgc0TaWr+30HWe3a4ltaBwTZhyTEggF5tJv8tbt" crossorigin="anonymous">
		<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.1/css/bootstrap.min.css" integrity="sha384-WskhaSGFgHYWDcbwN70/dfYBj47jz9qbsMId/iRN3ewGhXQFZCSftd1LZCfmhktB" crossorigin="anonymous">
		<script src="http://code.jquery.com/jquery-1.11.3.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.3/umd/popper.min.js" integrity="sha384-ZMP7rVo3mIykV+2+9J3UJ46jBk0WLaUAdn689aCwoqbBJiSnjAK/l8WvCWPIPm49" crossorigin="anonymous"></script>
		<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.1/js/bootstrap.min.js" integrity="sha384-smHYKdLADwkXOn1EmN1qk/HfnUcbVRZyYmZ4qpPea6sjB/pTJ0euyQp0Mk8ck+5T" crossorigin="anonymous"></script>
		<title>TorrServer ` + version.Version + `</title>
	</head>
	<body>
		<style>
			.wrap {
				white-space: normal;
				word-wrap: break-word;
				word-break: break-all;
			}
			.content {
				margin: 1%;
			}
			.cache {
				display: grid;
            	grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
			}
			.piece {
				border: 1px dashed white;
				font-size: 16px;
				padding: 2px;
				text-align: center;
			}
		</style>
		
		<nav class="navbar navbar-expand-lg navbar-dark bg-dark">
			<a class="btn navbar-btn pull-left" href="/"><i class="fas fa-arrow-left"></i></a>
			<span class="navbar-brand mx-auto">
			TorrServer ` + version.Version + `
			</span>
		</nav>
		<div class="content">
			<div id="torrents"></div>
			<div id="cacheInfo"></div>
			<div class="cache" id="cache"></div>
		</div>
		<footer class="page-footer navbar-dark bg-dark">
			<span class="navbar-brand d-flex justify-content-center">
			<a rel="external" style="text-decoration: none;" href="/about">Описание</a>
			</span>
		</footer>
	</body>
	<script>
		$( document ).ready(function() {
			setInterval(updateState, 1000);
		});
		
		var cacheHash = "";
		var hashTorrents = "";
		
		function updateTorrents(){
			listTorrent(function(data){
				var currHashTorrs = ""; 
				for(var key in data) {
					var tor = data[key];
					currHashTorrs += tor.Hash;
				}
				if (currHashTorrs != hashTorrents){
					hashTorrents = currHashTorrs; 
					var html = "";
					html += '<div class="btn-group-vertical d-flex" role="group">';
					for(var key in data) {
						var tor = data[key];
						html += '<button type="button" class="btn btn-secondary wrap w-100" onclick="setCache(\''+tor.Hash+'\')">'+tor.Name+'</button>';
					}
					html += '</div>'
					$("#torrents").empty();
					$(html).appendTo($("#torrents"));
				}
			});
		}
		
		function updateCache(){
			var cache = $("#cache");
			if (cacheHash!=""){
				cacheTorrent(cacheHash, function(data){
					var html = "";
					var st = data; 
					html += '<span>Hash: '+st.Hash+'</span><br>';
					html += '<span>Capacity: '+humanizeSize(st.Capacity)+'</span><br>';
					html += '<span>Filled: '+humanizeSize(st.Filled)+'</span><br>';
					html += '<span>Pieces length: '+humanizeSize(st.PiecesLength)+'</span><br>';
					html += '<span>Pieces count: '+st.PiecesCount+'</span><br>';
					$("#cacheInfo").html(html);
					makePieces(st.PiecesCount);
					for(var i = 0; i < st.PiecesCount; i++) {
						var color = "silver";
						var size = "";
						var piece = st.Pieces[i];
						if (piece){
							if (piece.Completed && piece.BufferSize >= st.PiecesLength)
								color = "green";
							else if (piece.Completed && piece.BufferSize == 0)
								color = "silver";
							else if (!piece.Completed && piece.BufferSize > 0)
								color = "red";
							size = ' ' + humanizeSize(piece.BufferSize);
						}
						setPiece(i,color,size);
					}
				},function(){
					$("#cacheInfo").empty();
					cache.empty();
				});
			}
		}
		
		function makePieces(len){
			var cache = $("#cache");
			if (cache.children().length==len)
				return;
			var html = "";
			for(var i = 0; i < len; i++) {
				html += '<span class="piece" id="p'+i+'" style="background-color: silver;">'+i+'</span>';
			}
			cache.html(html);
		}
			
		function setPiece(i, color, size){
			var piece = $("#p"+i);
			piece.delay(100).css("background-color",color);
			piece.text(i+''+size);
		}
			
		function contains(arr, elem) {
			for (var i = 0; i < arr.length; i++) {
				if (arr[i].Id === elem) {
					return true;
				}
			}
			return false;
		}
			
		function updateState(){
			updateTorrents();
			updateCache();
		}
		
		function setCache(hash){
			cacheHash = hash;
			updateCache();
		}
	</script>
</html>
`

func (t *Template) parseCachePage() {
	parsePage(t, "cachePage", cachePage)
}
