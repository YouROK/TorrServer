package templates

import (
	"server/version"
)

var mainPage = `
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
		</style>
		
		<nav class="navbar navbar-expand-lg navbar-dark bg-dark">
			<span class="navbar-brand mx-auto">
			TorrServer ` + version.Version + `
			</span>
		</nav>
		<div class="content">
			<div>
				<label for="magnet">Добавить торрент: </label>
				<input id="magnet" class="w-100" autocomplete="off">
			</div>
			<div class="btn-group d-flex" role="group">
				<button id="buttonAdd" class="btn w-100" onclick="addTorr()"><i class="fas fa-plus"></i> Добавить</button>
				<button id="buttonUpload" class="btn w-100"><i class="fas fa-file-upload"></i> Загрузить файл</button>
			</div>
			<br>
			<div>
				<a href="/torrent/playlist.m3u" rel="external" class="btn btn-primary w-100" role="button" ><i class="fas fa-th-list"></i> Плейлист всех торрентов</a>
			</div>
			<br>
			<h3>Торренты: </h3>
			<div id="torrents"></div>
			<br>
			<div class="btn-group-vertical d-flex" role="group">
				<a href="/settings" rel="external" class="btn btn-primary w-100" role="button"><i class="fas fa-cog"></i> Настройки</a>
				<a href="/cache" rel="external" class="btn btn-primary w-100" role="button"><i class="fas fa-info"></i> Кэш</a>
				<button id="buttonShutdown" class="btn btn-primary w-100" onclick="shutdown()"><i class="fas fa-power-off"></i> Закрыть сервер</button>
			</div>
			<form id="uploadForm" style="display:none" action="/torrent/upload" method="post">
				<input type="file" id="filesUpload" style="display:none" multiple onchange="uploadTorrent()" name="files"/> 
			</form>
		</div>
		<footer class="page-footer navbar-dark bg-dark">
			<span class="navbar-brand d-flex justify-content-center">
			<a rel="external" style="text-decoration: none;" href="/about">Описание</a>
			</span>
		</footer>
		
		<div class="modal fade" id="preloadModal" role="dialog">
			<div class="modal-dialog">
				<div class="modal-content">
					<div class="modal-header">
						<h4 class="modal-title wrap" id="preloadName"></h4>
					</div>
					<div class="modal-body">
						<p id="preloadStatus"></p>
						<p id="preloadBuffer"></p>
						<p id="preloadPeers"></p>
						<p id="preloadSpeed"></p>
						<div class="progress">
							<div id="preloadProgress" class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar" aria-valuenow="100" aria-valuemin="0" aria-valuemax="100" style="width: 100%"></div>
						</div>
						<br>
						<a id="preloadFileLink" role="button" href="" class="btn btn-secondary wrap w-100"></a>
					</div>
					<div class="modal-footer">
						<button type="button" class="btn btn-danger" data-dismiss="modal">Закрыть</button>
					</div>
				</div>
			</div>
		</div>
		<script>
			function addTorr(){
				var magnet = $("#magnet").val();
				$("#magnet").val("");
				if(magnet!=""){
					addTorrent(magnet,true,
					function( data ) {
						loadTorrents();
					},
					function( data ) {
						alert(data.responseJSON.message);
					});
				}
			}
			
			function removeTorr(hash){
				if(hash!=""){
					removeTorrent(hash,
					function( data ) {
						loadTorrents();
					},
					function( data ) {
						alert(data.responseJSON.message);
					});
				}
			};
			
			function shutdown(){
				shutdownServer(function( data ) {
						alert(data.responseJSON.message);
				});
			}
			
			$( document ).ready(function() {
				watchInfo();
			});
			
			$('#buttonUpload').click(function() {
			  		$('#filesUpload').click();
			});
			
			function uploadTorrent() {
				var form = $("#uploadForm");
				var formData = new FormData(document.getElementById("uploadForm"));
				var data = new FormData();
				$.each($('#filesUpload')[0].files, function(i, file) {
			   		data.append('file-'+i, file);
				});
				$.ajax({
						cache: false,
						processData: false,
						contentType: false,
						type: form.attr('method'),
						url: form.attr('action'),
						data: data
						}).done(function(data) {
							loadTorrents();
						}).fail(function(data) {
							alert(data.responseJSON.message);
						});
			}
			
			$('#uploadForm').submit(function(event) {
				event.preventDefault();
				var form = $(this);
				$.ajax({
					type: form.attr('method'),
					url: form.attr('action'),
					data: form.serialize()
					}).done(function(data) {
						loadTorrents();
					});
			});
			
			function loadTorrents() {
				listTorrent(
					function( data ) {
						var torrents = $("#torrents");
						torrents.empty();
						var html = "";
						var queueInfo = [];
						for(var key in data) {
							var tor = data[key];
							if (tor.Status==1){
								queueInfo.push(tor);
								continue;
							}
							html += tor2Html(tor);
						}
						if (queueInfo.length>0){
							html += "<br><hr><h3>Got info: </h3>";
							for(var key in queueInfo) {
								var tor = queueInfo[key];
								html += tor2Html(tor);
							}
						}
						$(html).appendTo(torrents);
					},
					function( data ) {
						alert(data.responseJSON.message);
					});
			}
			
			function tor2Html(tor){
				var html = '';
				var name = "";
				if (tor.Status==1)
					name = tor.Name+' '+humanizeSize(tor.Length)+' '+tor.Hash;
				else
					name = tor.Name+' '+humanizeSize(tor.Length);
			
				html += '<div class="btn-group d-flex" role="group">';
				html += '	<button type="button" class="btn btn-secondary wrap w-100" data-toggle="collapse" data-target="#info_'+tor.Hash+'">'+name+'</button>';
				if (tor.Status!=1)
					html += '	<a role="button" class="btn btn-secondary" href="'+tor.Playlist+'"><i class="fas fa-th-list"></i> Плейлист</a>';
				else
					html += '	<button type="button" class="btn btn-secondary" onclick="showPreload(\'\', \''+ tor.Hash +'\');"><i class="fas fa-info"></i></a>';
				html += '	<button type="button" class="btn btn-secondary" onclick="removeTorrent(\''+tor.Hash+'\');"><i class="fas fa-trash-alt"></i> Удалить</button>';
				html += '</div>';
				html += '<div class="collapse" id="info_'+tor.Hash+'">';
				for(var i in tor.Files){
					var file = tor.Files[i];
				  	var ico = "";
				  	if (file.Viewed)
				  		ico = '<i class="far fa-eye"></i> ';
					html += '	<div class="btn-group d-flex" role="group">';
					html += '		<a role="button" href="'+file.Link+'" class="btn btn-secondary wrap w-100">'+ico+file.Name+" "+humanizeSize(file.Size)+'</a>';
					html += '		<button type="button" class="btn btn-secondary" onclick="showPreload(\''+ file.Preload +'\', \''+ file.Link +'\', \''+ tor.Hash +'\');"><i class="fas fa-info"></i></button>';
					html += '	</div>';
				}
				html += '<hr></div>';
				return html;
			}
			
			function watchInfo(){
				var lastTorrentCount = 0;
				var lastGettingInfo = 0;
				setInterval(function() {
					listTorrent(
					function( data ) {
						var gettingInfo = 0;
						for(var key in data) {
							var tor = data[key];
							if (tor.Status==1)
								gettingInfo++;
						}
			
						if (lastTorrentCount!=data.length || gettingInfo!=lastGettingInfo){
							loadTorrents();
							lastTorrentCount = data.length;
							lastGettingInfo = gettingInfo;
						}
					});
				}, 1000);
			}
				
			function showPreload(preloadlink, fileLink, hash){
				$('#preloadFileLink').hide(0);
				$('#preloadFileLink').attr("href","");
				$('#preloadProgress').width('100%');
				if (preloadlink!='')
					preloadTorrent(preloadlink);
				var ptimer = setInterval(function() {
					statTorrent(hash,function(data){
						if (data!=null){
							$('#preloadStatus').text("Status: " + data.TorrentStatusString);
							$('#preloadName').text(data.Name);
							$('#preloadPeers').text("Peers: [" + data.ConnectedSeeders + "] " + data.ActivePeers + " / " + data.TotalPeers);
							if (data.DownloadSpeed>0)
								$('#preloadSpeed').text("Speed: "+ humanizeSize(data.DownloadSpeed) + "/Sec");
							else
								$('#preloadSpeed').text("Speed:");
			
							if (data.PreloadSize>0 && data.PreloadedBytes<data.PreloadSize){
								var prc = data.PreloadedBytes * 100 / data.PreloadSize;
								if (prc>100) prc = 100;
								$('#preloadProgress').width(prc+'%');
								$('#preloadBuffer').text("Loaded: " + humanizeSize(data.PreloadedBytes) + " / " + humanizeSize(data.PreloadSize)+" "+prc+"%");
							}else{
								$('#preloadProgress').width('100%');
								$('#preloadBuffer').text("Loaded: " + humanizeSize(data.BytesReadUsefulData));
								$('#preloadProgress').width('100%');
								if (data.BytesReadUsefulData>0 && fileLink && !$('#preloadFileLink').attr("href")){
									$('#preloadFileLink').text(data.Name);
									$('#preloadFileLink').attr("href", fileLink);
									$('#preloadFileLink').show();
								}
							}
						}
					},function(){
						$('#preloadModal').modal('hide');
					})
				}, 500);
				$('#preloadModal').modal('show');
				$("#preloadModal").on('hidden.bs.modal', function () {
					clearInterval(ptimer);
				});
			}
			
		</script>
	</body>
</html>
`

func (t *Template) parseMainPage() {
	parsePage(t, "mainPage", mainPage)
}
