#!/bin/bash
dirInstall="/Users/Shared/TorrServer"
serviceName="ru.yourok.torrserver"

function checkArch() {
  case $(uname -m) in
    i386) architecture="386" ;;
    i686) architecture="386" ;;
    x86_64) architecture="amd64" ;;
    aarch64) architecture="arm64" ;;
    *) echo "Извините, не поддерживаемая архитектура. Продолжение невозможно" && exit 1;;
  esac
}

function getLatestRelease() {
  curl -s "https://api.github.com/repos/YouROK/TorrServer/releases" |
  grep -iE '"tag_name":|"version":' |
  sed -E 's/.*"([^"]+)".*/\1/' |
  head -1
}

function cleanup() {
  sudo rm -f /Library/LaunchAgents/*torrserver* 1>/dev/null 2>&1
  sudo rm -f /Library/LaunchDaemons/*torrserver* 1>/dev/null 2>&1
  sudo rm -f $HOME/Library/LaunchAgents/*torrserver* 1>/dev/null 2>&1
  sudo rm -f $HOME/Library/LaunchDaemons/*torrserver* 1>/dev/null 2>&1
  self="$(basename "$0")"
  runningPid=$(ps -ax|grep -i torrserver|grep -v grep|grep -v "$self"|awk '{print $1}')
  sudo kill -9 $runningPid 1>/dev/null 2>&1
}

function uninstall() {
  echo ""
  echo "Директория c TorrServer - ${dirInstall}"
  echo ""
  echo "Это действие удалит все данные TorrServer включая базу данных торрентов и настройки по указанному выше пути."
  echo ""
  printf 'Вы уверены что хотите удалить программу? '
  read answer
  if [ "$answer" != "${answer#[YyДд]}" ] ; then
    cleanup
    sudo rm -rf $dirInstall
    echo ""
    echo "TorrServer удален c вашего Mac"
    echo ""
    sleep 5
  else
    echo ""
    echo "OK"
    echo ""
    sleep 5
  fi
}
  
checkArch

function installTorrServer() {
  echo ""
  echo "Устанавливаем TorrServer $(getLatestRelease) ..."
  echo ""
  binName="TorrServer-darwin-${architecture}"
  [[ ! -d "$dirInstall" ]] && mkdir -p ${dirInstall}
  urlBin="https://github.com/YouROK/TorrServer/releases/download/$(getLatestRelease)/${binName}"
  if [[ ! -f "$dirInstall/$binName" ]] | [[ ! -x "$dirInstall/$binName" ]] || [[ $(stat -c%s "$dirInstall/$binName" 2>/dev/null) -eq 0 ]]; then
    curl -L --progress-bar -# -o "$dirInstall/$binName" "$urlBin"
    chmod a+rx "$dirInstall/$binName"
    xattr -r -d com.apple.quarantine "$dirInstall/$binName"
  fi
  echo ""
  echo "Создаем сервис автозагрузки TorrServer $(getLatestRelease) ..."
  echo ""
  echo "Система запросит ваш пароль администратора"
  echo ""
  cat << EOF > $dirInstall/$serviceName.plist
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>${serviceName}</string>
  <key>ServiceDescription</key>
  <string>TorrServer service for MacOS</string>
  <key>LaunchOnlyOnce</key>
  <true/>
  <key>RunAtLoad</key>
  <true/>
  <key>ProgramArguments</key>
  <array>
    <string>${dirInstall}/TorrServer-darwin-${architecture}</string>
    <string>--port</string>
    <string>8090</string>
    <string>--path</string>
    <string>${dirInstall}</string>
    <string>--logpath</string>
    <string>${dirInstall}/torrserver.log</string>
    <string>--httpauth</string>
  </array>
  <key>StandardOutPath</key>
  <string>${dirInstall}/torrserver.log</string>
  <key>StandardErrorPath</key>
  <string>${dirInstall}/torrserver.log</string>
</dict>
</plist>
EOF
###
  cleanup
###
  printf 'Включить авторизацию на сервере? '
  read answer
  if [ "$answer" != "${answer#[YyДд]}" ] ;then
    isAuth=1
  else
    isAuth=0
  fi
  echo ""
  if [[ "$isAuth" == 1 ]]; then

    echo "Вы выбрали установку с авторизацией"
    [[ ! -f "$dirInstall/accs.db" ]] && {
      echo ""
      printf 'Пользователь: '
      read answer
      isAuthUser=$answer
      echo ""
      printf 'Пароль: '
      read answer
      isAuthPass=$answer
      echo ""
      echo "Устанавливаем логин и пароль: $isAuthUser:$isAuthPass"
      echo ""
      echo -e "{\n  \"$isAuthUser\": \"$isAuthPass\"\n}" > $dirInstall/accs.db
    } || {
      echo ""
      echo "Используйте реквизиты из ${dirInstall}/accs.db для входа"
      echo ""
    }
  else
    sed -i '' -e '/httpauth/d' $dirInstall/$serviceName.plist
  fi
  printf 'Автозагрузка для текушего пользователя (1) или всех (2)? '
  read answer
  if [ "$answer" != "${answer#[1]}" ] ;then
    # user
    sysPath="${HOME}/Library/LaunchAgents"
    [[ ! -d "$sysPath" ]] && mkdir -p ${sysPath}
    cp "$dirInstall/$serviceName.plist" $sysPath
    chmod 0644 "$sysPath/$serviceName.plist"
    launchctl load -w "$sysPath/$serviceName.plist" 1>/dev/null 2>&1
  else
    # root
    sysPath="/Library/LaunchDaemons"
    [[ ! -d "$sysPath" ]] && sudo mkdir -p ${sysPath}
    sudo cp "$dirInstall/$serviceName.plist" $sysPath
    sudo chown root:wheel "$sysPath/$serviceName.plist"
    sudo chmod 0644 "$sysPath/$serviceName.plist"
    sudo launchctl load -w "$sysPath/$serviceName.plist" 1>/dev/null 2>&1
  fi
  echo ""
  echo "Сервис автозагрузки записан в ${sysPath}"
  echo ""
  echo "TorrServer $(getLatestRelease) для ${architecture} Mac установлен в ${dirInstall}"
  echo ""
  sleep 15
}

while true; do
  echo ""
  read -p "Хотите установить или обновить TorrServer? Для удаления введите «Удалить» " yn
  case $yn in
    [YyДд]* ) installTorrServer; break;;
    [DdУу]* ) uninstall; break;;
    [NnНн]* ) exit;;
    * ) echo "Ввведите Да (Yes) Нет (No) или Удалить (Delete).";;
  esac
done
echo "Have Fun!"
echo ""
sleep 5