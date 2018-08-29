package templates

import "server/version"

var aboutPage = `
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
	<title>About</title>
</head>
<body>
<style type="text/css">
	.inline{
		display:inline;
		padding-left: 2%;
	}
	.center {
		display: block;
		margin-left: auto;
		margin-right: auto;
	}
	.content {
		padding: 20px;
	}
</style>
	<nav class="navbar navbar-expand-lg navbar-dark bg-dark">
    	<a class="btn navbar-btn pull-left" href="/"><i class="fas fa-arrow-left"></i></a>
        <span class="navbar-brand mx-auto">
			О программе
		</span>
    </nav>
	<div class="content">
		<img class="center" src='` + faviconB64 + `'/>
		<h3 align="middle">TorrServer</h3>
		<h4 align="middle">` + version.Version + `</h4>
		
		<h4>Поддержка проекта:</h4>
		<a class="inline" target="_blank" href="https://www.paypal.me/yourok">PayPal</a>
		<br>
		<a class="inline" target="_blank" href="https://money.yandex.ru/to/410013733697114/100">Yandex.Деньги</a>
		<br>
		<hr align="left" width="25%">
		<br>
		
		<h4>Инструкция по использованию:</h4>
			<a class="inline" target="_blank" href="https://4pda.ru/forum/index.php?showtopic=896840&st=0#entry72570782">4pda.ru</a>
			<p class="inline">Спасибо <b>MadAndron</b></p> 
		<br>
		<hr align="left" width="25%">
		<br>
		
		<h4>Автор:</h4>
			<b class="inline">YouROK</b>
			<br>
			<i class="inline">Email:</i>
			<a target="_blank" class="inline" href="mailto:8yourok8@gmail.com">8YouROK8@gmail.com</a>
			<br>
			<i class="inline">Site: </i>
			<a target="_blank" class="inline" href="https://github.com/YouROK">GitHub.com/YouROK</a>
		<br>
		<hr align="left" width="25%">
		<br>
		
		<h4>Спасибо всем, кто тестировал и помогал:</h4>
			<b class="inline">kuzzman</b>
			<br>
			<i class="inline">Site: </i>
			<a target="_blank" class="inline" href="https://4pda.ru/forum/index.php?showuser=1259550">4pda.ru</a>
			<a target="_blank" class="inline" href="http://tv-box.pp.ua">tv-box.pp.ua</a>
		<br>
		<br>
			<b class="inline">MadAndron</b>
			<br>
			<i class="inline">Site:</i>
			<a target="_blank" class="inline" href="https://4pda.ru/forum/index.php?showuser=1543999">4pda.ru</a>
		<br>
		<br>
			<b class="inline">SpAwN_LMG</b>
			<br>
			<i class="inline">Site:</i>
			<a target="_blank" class="inline" href="https://4pda.ru/forum/index.php?showuser=700929">4pda.ru</a>
		<br>
		<br>
			<b class="inline">Zivio</b>
			<br>
			<i class="inline">Site:</i>
			<a target="_blank" class="inline" href="https://4pda.ru/forum/index.php?showuser=1195633">4pda.ru</a>
			<a target="_blank" class="inline" href="http://forum.hdtv.ru/index.php?showtopic=19020">forum.hdtv.ru</a>
		<br>
		<br>
			<b class="inline">Tw1cker Руслан Пахнев</b>
			<br>
			<i class="inline">Site:</i>
			<a target="_blank" class="inline" href="https://4pda.ru/forum/index.php?showuser=2002724">4pda.ru</a>
			<a target="_blank" class="inline" href="https://github.com/Nemiroff">GitHub.com/Nemiroff</a>
		<br>
		<br>
	</div>
	<footer class="page-footer navbar-dark bg-dark">
		<span class="navbar-brand d-flex justify-content-center">
			<center><h4>TorrServer ` + version.Version + `</h4></center>
		</span>
    </footer>
</body>
</html>
`

func (t *Template) parseAboutPage() {
	parsePage(t, "aboutPage", aboutPage)
}
