#!/bin/bash
dirInstall="/Users/Shared/TorrServer"
serviceName="torrserver"

function getLang() {
  lang=$(locale | grep LANG | cut -d= -f2 | tr -d '"' | cut -d_ -f1)
  [[ $lang != "ru" ]] && lang="en"
}

function checkArch() {
  case $(uname -m) in
    i386) architecture="386" ;;
    i686) architecture="386" ;;
    x86_64) architecture="amd64" ;;
    arm64) architecture="arm64" ;;
    aarch64) architecture="arm64" ;;
    *) [[ $lang == "en" ]] && { echo ""; echo " Unsupported Arch. Can't continue."; exit 1; } || { echo ""; echo " Не поддерживаемая архитектура. Продолжение невозможно."; exit 1; } ;;
  esac
}

function getLatestRelease() {
  curl -s "https://api.github.com/repos/YouROK/TorrServer/releases" |
  grep -iE '"tag_name":|"version":' |
  sed -E 's/.*"([^"]+)".*/\1/' |
  head -1
}

function killRunning() {
  self="$(basename "$0")"
  runningPid=$(ps -ax|grep -i torrserver|grep -v grep|grep -v "$self"|awk '{print $1}')
  [[ -z $runningPid ]] || sudo kill -9 $runningPid
}

function cleanup() {
  sudo rm -f /Library/LaunchAgents/*torrserver*
  sudo rm -f /Library/LaunchDaemons/*torrserver*
  sudo rm -f $HOME/Library/LaunchAgents/*torrserver*
  sudo rm -f $HOME/Library/LaunchDaemons/*torrserver*
  killRunning
}

function uninstall() {
  [[ $lang == "en" ]] && {
    echo ""
    echo " TorrServer install dir - ${dirInstall}"
    echo ""
    echo " This action will delete TorrServer including all it's torrents, settings and files on path above."
    echo ""
  } || {
    echo ""
    echo " Директория c TorrServer - ${dirInstall}"
    echo ""
    echo " Это действие удалит все данные TorrServer включая базу данных торрентов и настройки по указанному выше пути."
    echo ""
  }
  [[ $lang == "en" ]] && read -p ' Are you shure you want to delete TorrServer? (Yes/No) ' answer_del </dev/tty || read -p ' Вы уверены что хотите удалить программу? (Да/Нет) ' answer_del </dev/tty
  if [ "$answer_del" != "${answer_del#[YyДд]}" ]; then
    cleanup
    sudo rm -rf $dirInstall
    echo ""
    [[ $lang == "en" ]] && echo " TorrServer deleted from Mac" || echo " TorrServer удален c вашего Mac"
    echo ""
    sleep 5
  else
    echo ""
    echo "OK"
    echo ""
    sleep 5
  fi
}

function installTorrServer() {
  [[ $lang == "en" ]] && {
    echo ""
    echo " Install TorrServer $(getLatestRelease)…"
    echo ""
  } || {
    echo ""
    echo " Устанавливаем TorrServer $(getLatestRelease)…"
    echo ""
  }
  user=$(whoami)
  binName="TorrServer-darwin-${architecture}"
  [[ ! -d "$dirInstall" ]] && mkdir -p ${dirInstall} && chmod a+rw ${dirInstall}
  urlBin="https://github.com/YouROK/TorrServer/releases/download/$(getLatestRelease)/${binName}"
  if [[ ! -f "$dirInstall/$binName" ]] | [[ ! -x "$dirInstall/$binName" ]] || [[ $(stat -c%s "$dirInstall/$binName" 2>/dev/null) -eq 0 ]]; then
    curl -L --progress-bar -# -o "$dirInstall/$binName" "$urlBin"
    chmod a+rx "$dirInstall/$binName"
    xattr -r -d com.apple.quarantine "$dirInstall/$binName"
  fi
  [[ $lang == "en" ]] && {
    echo ""
    echo " Add autostart service for TorrServer $(getLatestRelease)…"
    echo ""
    echo " System can ask your admin account password"
    echo ""
  } || {
    echo ""
    echo " Создаем сервис автозагрузки TorrServer $(getLatestRelease)…"
    echo ""
    echo " Система может запросить ваш пароль администратора"
    echo ""
  }
###
  cleanup
###
  [[ $lang == "en" ]] && read -p ' Change TorrServer web port? (Yes/No) ' answer_cp </dev/tty || read -p ' Хотите изменить веб-порт для TorrServer? (Да/Нет) ' answer_cp </dev/tty
  if [ "$answer_cp" != "${answer_cp#[YyДд]}" ]; then
    echo ""
    [[ $lang == "en" ]] && read -p ' Enter port number: ' answer_port </dev/tty || read -p ' Введите номер порта: ' answer_port </dev/tty
    servicePort=$answer_port
    echo ""
  else
    servicePort="8090"
    echo ""
  fi
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
    <string>${servicePort}</string>
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
  [[ $lang == "en" ]] && read -p ' Enable HTTP Authorization? (Yes/No) ' answer_auth </dev/tty || read -p ' Включить авторизацию на сервере? (Да/Нет) ' answer_auth </dev/tty
  if [ "$answer_auth" != "${answer_auth#[YyДд]}" ]; then
    isAuth=1
  else
    isAuth=0
  fi
  echo ""
  if [[ "$isAuth" == 1 ]]; then
    [[ $lang == "en" ]] && echo " HTTP Auth Install choosen" || echo " Вы выбрали установку с авторизацией"
    [[ ! -f "$dirInstall/accs.db" ]] && {
      echo ""
      [[ $lang == "en" ]] && read -p ' User: ' answer_user </dev/tty || read -p ' Пользователь: ' answer_user </dev/tty 
      isAuthUser=$answer_user
      echo ""
      [[ $lang == "en" ]] && read -p ' Password: ' answer_pass </dev/tty || read -p ' Пароль: ' answer_pass </dev/tty
      isAuthPass=$answer_pass
      echo ""
      [[ $lang == "en" ]] && echo " Added credentials: $isAuthUser:$isAuthPass" || echo " Устанавливаем логин и пароль: $isAuthUser:$isAuthPass"
      echo ""
      echo -e "{\n  \"$isAuthUser\": \"$isAuthPass\"\n}" > $dirInstall/accs.db
    } || {
      echo ""
      [[ $lang == "en" ]] && echo " Use ${dirInstall}/accs.db credentials for access" || echo " Используйте реквизиты из ${dirInstall}/accs.db для входа"
      echo ""
    }
  else
    sed -i '' -e '/httpauth/d' $dirInstall/$serviceName.plist
  fi
  [[ $lang == "en" ]] && read -p ' Add autostart for current user (1) or all users (2)? ' answer_cu </dev/tty || read -p ' Добавить автозагрузку для текущего пользователя (1) или для всех (2)? ' answer_cu </dev/tty
  if [ "$answer_cu" != "${answer_cu#[1]}" ]; then
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
  [[ $lang == "en" ]] && {
    echo ""
    echo " Autostart service added to ${sysPath}"
    echo ""
    echo " TorrServer $(getLatestRelease) for ${architecture} Mac installed to ${dirInstall}"
    echo ""
    echo " You can now open browser URL http://localhost:$servicePort to access TorrServer GUI"
    echo ""
  } || {
    echo ""
    echo " Сервис автозагрузки записан в ${sysPath}"
    echo ""
    echo " TorrServer $(getLatestRelease) для ${architecture} Mac установлен в ${dirInstall}"
    echo ""
    echo " Теперь вы можете открыть браузер по адресу http://localhost:$servicePort для доступа к вебу TorrServer"
    echo ""
  }
  if [[ "$isAuth" == 1 && $isAuthUser > 0 ]]; then
    [[ $lang == "en" ]] && echo " Use user \"$isAuthUser\" with password \"$isAuthPass\" for web auth" || echo " Для авторизации введите пользователя $isAuthUser с паролем $isAuthPass"
    echo ""
  fi
  sleep 30
}

while true; do
  getLang
  echo ""
  echo "=============================================================="
  [[ $lang == "en" ]] && echo " TorrServer install, update and uninstall script for MacOS " || echo " Скрипт установки, обновления и удаления TorrServer для MacOS "
  echo "=============================================================="
  echo ""
  [[ $lang == "en" ]] && read -p " Do You want to install or update TorrServer? (Yes or No). Enter \"Delete\" to Uninstall TorrServer. " ydn </dev/tty || read -p " Хотите установить или обновить TorrServer? (Да|Нет). Для удаления введите «Удалить». " ydn </dev/tty
  case $ydn in
    [YyДд]*) checkArch; installTorrServer; break ;;
    [DdУу]*) uninstall; break ;;
    [NnНн]*) exit ;;
    *) [[ $lang == "en" ]] && { echo ""; echo " Enter \"Yes\", \"No\" or \"Delete\"."; } || { echo ""; echo " Ввведите «Да», «Нет» или «Удалить»."; } ;;
  esac
done
echo " Have Fun!"
echo ""
sleep 5
