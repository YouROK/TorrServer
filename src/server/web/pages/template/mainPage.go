package template

import (
	"server/version"
)

func (t *Template) parseMainPage() {
	t.parsePage("mainPage", mainPage)
}

const mainPage = `
<!DOCTYPE html>
<html lang="en">
	` + header + `
<body ng-app="app">
<script src="/api.js"></script>
<script src="/main.js"></script>

<nav class="navbar navbar-expand-lg navbar-light bg-light {{active}}" ng-click="$event.preventDefault()">
    <div class="navbar-nav">
        <a href="#" class="nav-item nav-link torrents" ng-click="active='torrents'">Torrents</a>
        <a href="#" class="nav-item nav-link settings" ng-click="active='settings'">Settings</a>
        <a href="#" class="nav-item nav-link cache" ng-click="active='cache'">Cache</a>
        <a href="#" class="nav-item nav-link about" ng-click="active='about'">About</a>
    </div>
</nav>
    
    <p ng-hide="active">Please click a menu item</p>
    <p ng-show="active">You chose <b>{{active}}</b></p>
 
		
</body>
</html>
`

const tmp = `

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
	</body>

<script id="contactTemplate" type="text/template">
    <img src="<%= photo %>" alt="<%= name %>" />
    <h1><%= name %><span><%= type %></span></h1>
    <div><%= address %></div>
    <dl>
        <dt>Tel:</dt><dd><%= tel %></dd>
        <dt>Email:</dt><dd><a href="mailto:<%= email %>"><%= email %></a></dd>
    </dl>
</script>

</html>
`
