package templates

import "server/version"

var settingsPage = `
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
    <title>TorrServer ` + version.Version + ` Settings</title>
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
    	<a class="btn navbar-btn pull-left" href="/"><i class="fas fa-arrow-left"></i></a>
        <span class="navbar-brand mx-auto">
         TorrServer ` + version.Version + `
         </span>
    </nav>
    <div class="content">
        <form id="settings">
            <div class="input-group">
                <div class="input-group-prepend">
                    <div class="input-group-text">Размер кэша</div>
                </div>
                <input id="CacheSize" class="form-control" type="number" autocomplete="off">
            </div>
		<br>
            <div class="input-group">
                <div class="input-group-prepend">
                    <div class="input-group-text">Размер буфера предзагрузки</div>
                </div>
                <input id="PreloadBufferSize" class="form-control" type="number" autocomplete="off">
            </div>
         	<small class="form-text text-muted">Размеры кэша и буфера указываются в мегабайтах</small>
		<br>
            <div class="form-check">
                <input id="DisableTCP" class="form-check-input" type="checkbox" autocomplete="off">
                <label for="DisableTCP">Отключить TCP</label>
            </div>
            <div class="form-check">
                <input id="DisableUTP" class="form-check-input" type="checkbox" autocomplete="off">
                <label for="DisableUTP">Отключить UTP</label>
            </div>
            <div class="form-check">
                <input id="DisableUPNP" class="form-check-input" type="checkbox" autocomplete="off">
                <label for="DisableUPNP">Отключить UPNP</label>
            </div>
            <div class="form-check">
                <input id="DisableDHT" class="form-check-input" type="checkbox" autocomplete="off">
                <label for="DisableDHT">Отключить DHT</label>
            </div>
            <div class="form-check">
                <input id="DisableUpload" class="form-check-input" type="checkbox" autocomplete="off">
                <label for="DisableUpload">Отключить Отдачу</label>
            </div>
		<br>
            <div class="input-group">
                <div class="input-group-prepend">
                    <div class="input-group-text">Шифрование</div>
                </div>
                <select id="Encryption" class="form-control">
                    <option value="0">По умолчанию</option>
                    <option value="1">Отключить</option>
                    <option value="2">Принудительно</option>
                </select>
            </div>
	    	<br>
            <div class="input-group">
                <div class="input-group-prepend">
                    <div class="input-group-text">Порт для входящих торрент подключений</div>
                </div>
                <input id="PeersListenPort" class="form-control" type="number" autocomplete="off">
            </div>
         	<small class="form-text text-muted">0 - случайный порт</small>
		<br>
            <div class="input-group">
                <div class="input-group-prepend">
                    <div class="input-group-text">Количество соединений</div>
                </div>
                <input id="ConnectionsLimit" class="form-control" type="number" autocomplete="off">
            </div>
		<br>
            <div class="input-group">
                <div class="input-group-prepend">
                    <div class="input-group-text">Ограничение загрузки</div>
                </div>
                <input id="DownloadRateLimit" class="form-control" type="number" autocomplete="off">
            </div>
		<br>
            <div class="input-group">
                <div class="input-group-prepend">
                    <div class="input-group-text">Ограничение отдачи</div>
                </div>
                <input id="UploadRateLimit" class="form-control" type="number" autocomplete="off">
            </div>
            <small class="form-text text-muted">Ограничение устанавливается в Килобайтах, 0 - не ограничивать</small>
	 	<br>
            <div class="input-group">
                <div class="input-group-prepend">
                    <div class="input-group-text">Ретрекеры</div>
                </div>
                <select id="RetrackersMode" class="form-control">
         			<option value="0">Оставить как есть</option>
         			<option value="1">Добавить</option>
                    <option value="2">Удалить</option>
                </select>
            </div>
        </form>
        <br>
        <div class="btn-group d-flex" role="group">
            <button id="buttonSave" class="btn btn-primary w-100" data-icon="check" onclick="saveSettings()"><i class="far fa-save"></i> Сохранить</button>
         	<button id="buttonRefresh" class="btn btn-primary w-100" data-icon="refresh" onclick="refreshSettings()"><i class="fas fa-sync-alt"></i> Получить с сервера</button>
        </div>
    </div>
    <footer class="page-footer navbar-dark bg-dark">
        <span class="navbar-brand d-flex justify-content-center">
         <a rel="external" style="text-decoration: none;" href="/about">Описание</a>
         </span>
    </footer>
    <script>
        function saveSettings() {
            var data = {};
            data.CacheSize = Number($('#CacheSize').val())*(1024*1024);
			data.PreloadBufferSize = Number($('#PreloadBufferSize').val())*(1024*1024);
			
			data.DisableTCP = $('#DisableTCP').prop('checked');
			data.DisableUTP = $('#DisableUTP').prop('checked');
			data.DisableUPNP = $('#DisableUPNP').prop('checked');
			data.DisableDHT = $('#DisableDHT').prop('checked');
			data.DisableUpload = $('#DisableUpload').prop('checked');
			data.Encryption = Number($('#Encryption').val());
 
			data.ConnectionsLimit = Number($('#ConnectionsLimit').val());
 
			data.DownloadRateLimit = Number($('#DownloadRateLimit').val());
			data.UploadRateLimit = Number($('#UploadRateLimit').val());
			
			data.RetrackersMode = Number($('#RetrackersMode').val());
			data.PeersListenPort = Number($('#PeersListenPort').val());
         
            $.post("/settings/write", JSON.stringify(data))
                .done(function(data) {
         			restartService();
                    alert(data);
                })
                .fail(function(data) {
                    alert(data.responseJSON.message);
                });
        }

        function refreshSettings() {
            $.post("/settings/read")
                .done(function(data) {
         			$('#CacheSize').val(data.CacheSize/(1024*1024));
					$('#PreloadBufferSize').val(data.PreloadBufferSize/(1024*1024));
					
         			$('#DisableTCP').prop('checked', data.DisableTCP);
					$('#DisableUTP').prop('checked', data.DisableUTP);
					$('#DisableUPNP').prop('checked', data.DisableUPNP);
					$('#DisableDHT').prop('checked', data.DisableDHT);
					$('#DisableUpload').prop('checked', data.DisableUpload);
					$('#Encryption').val(data.Encryption);
         
         			$('#ConnectionsLimit').val(data.ConnectionsLimit);
         
					$('#DownloadRateLimit').val(data.DownloadRateLimit);
					$('#UploadRateLimit').val(data.UploadRateLimit);
					
         			$('#RetrackersMode').val(data.RetrackersMode);
				$('#PeersListenPort').val(data.PeersListenPort);
                });
        };

        $(document).ready(function() {
            refreshSettings();
        });

		$(document).on("wheel", "input[type=number]", function (e) {
			$(this).blur();
		});
    </script>
</body>

</html>
`

func (t *Template) parseSettingsPage() {
	parsePage(t, "settingsPage", settingsPage)
}
