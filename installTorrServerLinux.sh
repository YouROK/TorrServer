#!/usr/bin/env bash
dirInstall="/opt/torrserver" # путь установки torrserver
serviceName="torrserver" # имя службы: systemctl status torrserver.service
scriptname=$(basename "$(test -L "$0" && readlink "$0" || echo "$0")")

#################################
#       F U N C T I O N S       #
#################################

function isRoot() {
  if [ "$EUID" -ne 0 ]; then
    return 1
  fi
}

function checkRunning() {
  runningPid=$(ps -ax|grep -i torrserver|grep -v grep|grep -v "$scriptname"|awk '{print $1}')
  echo $runningPid
}

function getLang() {
  lang=$(locale | grep LANG | cut -d= -f2 | tr -d '"' | cut -d_ -f1)
  [[ $lang != "ru" ]] && lang="en"
}

function getIP() {
	[ -z "`which dig`" ] && serverIP=$(host myip.opendns.com resolver1.opendns.com | tail -n1 | cut -d' ' -f4-) || serverIP=$(dig +short myip.opendns.com @resolver1.opendns.com)
	# echo $serverIP
}

function uninstall() {
	[[ $lang == "en" ]] && {
		echo ""
		echo " TorrServer install dir - ${dirInstall}"
		echo ""
		echo " This action will delete TorrServer including all it's torrents, settings and files on path above!"
		echo ""
	} || {
		echo ""
		echo " Директория c TorrServer - ${dirInstall}"
		echo ""
		echo " Это действие удалит все данные TorrServer включая базу данных торрентов и настройки по указанному выше пути!"
		echo ""
  }
  [[ $lang == "en" ]] && read -p ' Are you shure you want to delete TorrServer? (Yes/No) ' answer </dev/tty || read -p ' Вы уверены что хотите удалить программу? (Да/Нет) ' answer </dev/tty
  if [ "$answer" != "${answer#[YyДд]}" ] ; then
    cleanup
    cleanAll
    echo ""
    [[ $lang == "en" ]] && echo " TorrServer deleted!" || echo " TorrServer удален!"
    echo ""
  else
    echo ""
  fi
}

function cleanup() {
  systemctl stop $serviceName 2>/dev/null
  systemctl disable $serviceName 2>/dev/null
  rm -rf /usr/local/lib/systemd/system/$serviceName.service $dirInstall 2>/dev/null
}

function cleanAll() { # guess other installs
  systemctl stop torr torrserver 2>/dev/null
  systemctl disable torr torrserver 2>/dev/null
  rm -rf /usr/local/torr 2>/dev/null
  rm -rf /opt/torr{,*} 2>/dev/null
  rm -f /{,etc,usr/local/lib}/systemd/system/tor{,r,rserver}.service 2>/dev/null
}

function helpUsage() {
  [[ $lang == "en" ]] && echo -e "$scriptname
  -i | --install | install - install latest release version
  -u | --update  | update  - install latest update (if any)
  -c | --check   | check   - check update (show only version info)
  -d | --down    | down    - version downgrade, need version number as argument
  -r | --remove  | remove  - uninstall TorrServer
  -h | --help    | help    - this help screen
" || echo -e "$scriptname
  -i | --install | install - установка последней версии
  -u | --update  | update  - установка последнего обновления, если имеется
  -c | --check   | check   - проверка обновления (выводит только информацию о версиях)
  -d | --down    | down    - понизить версию, после опции указывается версия для понижения
  -r | --remove  | remove  - удаление TorrServer
  -h | --help    | help    - эта справка
"
}

function checkOS() {
  if [[ -e /etc/debian_version ]]; then
    OS="debian"
    source /etc/os-release

    if [[ $ID == "debian" || $ID == "raspbian" ]]; then
      if [[ $VERSION_ID -lt 6 ]]; then
        echo "⚠️ Ваша версия Debian не поддерживается."
        echo ""
        echo " Скрипт поддерживает только Debian >=6"
        echo ""
        exit 1
      fi
    elif [[ $ID == "ubuntu" ]]; then
      OS="ubuntu"
      MAJOR_UBUNTU_VERSION=$(echo "$VERSION_ID" | cut -d '.' -f1)
      if [[ $MAJOR_UBUNTU_VERSION -lt 10 ]]; then
        echo "⚠️ Ваша версия Ubuntu не поддерживается."
        echo ""
        echo " Скрипт поддерживает только Ubuntu >=10"
        echo ""
        exit 1
      fi
    fi
  elif [[ -e /etc/system-release ]]; then
    source /etc/os-release
    if [[ $ID == "fedora" || $ID_LIKE == "fedora" ]]; then
      OS="fedora"
      # [ -z "$(rpm -qa wget)" ] && yum -y install wget
    fi
    if [[ $ID == "centos" || $ID == "rocky" || $ID == "redhat" ]]; then
      OS="centos"
      if [[ ! $VERSION_ID =~ (6|7|8) ]]; then
        echo "⚠️ Ваша версия CentOS/RockyLinux/RedHat не поддерживается."
        echo ""
        echo " Скрипт поддерживает только CentOS/RockyLinux/RedHat версии 6,7 и 8."
        echo ""
        exit 1
      fi
      # [ -z "$(rpm -qa wget)" ] && yum -y install wget
    fi
    if [[ $ID == "ol" ]]; then
      OS="oracle"
      if [[ ! $VERSION_ID =~ (6|7|8) ]]; then
        echo "⚠️ Ваша версия Oracle Linux не поддерживается."
        echo ""
        echo " Скрипт поддерживает только Oracle Linux версии 6,7 и 8."
        exit 1
      fi
      # [ -z "$(rpm -qa wget)" ] && yum -y install wget
    fi
    if [[ $ID == "amzn" ]]; then
      OS="amzn"
      if [[ $VERSION_ID != "2" ]]; then
        echo "⚠️ Ваша версия Amazon Linux не поддерживается."
        echo ""
        echo " Скрипт поддерживает только Amazon Linux 2."
        echo ""
        exit 1
      fi
      # [ -z "$(rpm -qa wget)" ] && yum -y install wget
    fi
  elif [[ -e /etc/arch-release ]]; then
    OS=arch
    # [ -z $(pacman -Qqe wget 2>/dev/null) ] &&  pacman -Sy --noconfirm wget
  else
    echo " Похоже, что вы запускаете этот установщик в системе отличной от Debian, Ubuntu, Fedora, CentOS, Amazon Linux, Oracle Linux или Arch Linux."
    exit 1
  fi
}

function checkArch() {
  case $(uname -m) in
    i386) architecture="386" ;;
    i686) architecture="386" ;;
    x86_64) architecture="amd64" ;;
    aarch64) architecture="arm64" ;;
    armv7|armv7l) architecture="arm7";;
    armv6|armv6l) architecture="arm5";;
    *) [[ $lang == "en" ]] && { echo " Unsupported Arch. Can't continue."; exit 1; } || { echo " Не поддерживаемая архитектура. Продолжение невозможно."; exit 1; } ;;
  esac
}

function checkInternet() {
  [ -z "`which ping`" ] && echo " Сначала установите iputils-ping" && exit 1
  [[ $lang == "en" ]] && echo " Check Internet access…" || echo " Проверяем соединение с Интернетом…"
  if ! ping -c 2 google.com &> /dev/null; then
    [[ $lang == "en" ]] && echo " - No Internet. Check your network and DNS settings." || echo " - Нет Интернета. Проверьте ваше соединение, а также разрешение имен DNS."
    exit 1
  fi
  [[ $lang == "en" ]] && echo " - Have Internet Access" || echo " - соединение с Интернетом успешно"
}

function initialCheck() {
  if ! isRoot; then
    [[ $lang == "en" ]] && echo " Script must run as root or user with sudo privileges. Example: sudo $scriptname" || echo " Вам нужно запустить скрипт от root или пользователя с правами sudo. Пример: sudo $scriptname"
    exit 1
  fi
  [ -z "`which curl`" ] && echo " Сначала установите curl" && exit 1
  checkOS
  checkArch
  checkInternet
}

function getLatestRelease() {
  curl -s "https://api.github.com/repos/YouROK/TorrServer/releases" |
  grep -iE '"tag_name":|"version":' |
  sed -E 's/.*"([^"]+)".*/\1/' |
  head -1
}

function installTorrServer() {
  [[ $lang == "en" ]] && echo " Install and configure TorrServer…" || echo " Устанавливаем и настраиваем TorrServer…"
  if checkInstalled; then
    if ! checkInstalledVersion; then
      [[ $lang == "en" ]] && read -p ' Want to update TorrServer? (Yes/No) ' answer </dev/tty || read -p ' Хотите обновить TorrServer? (Да/Нет) ' answer </dev/tty
      if [ "$answer" != "${answer#[YyДд]}" ] ;then
        UpdateVersion
      fi
    fi
  fi
  binName="TorrServer-linux-${architecture}"
  [[ ! -d "$dirInstall" ]] && mkdir -p ${dirInstall}
  [[ ! -d "/usr/local/lib/systemd/system" ]] && mkdir -p "/usr/local/lib/systemd/system"
  urlBin="https://github.com/YouROK/TorrServer/releases/download/$(getLatestRelease)/${binName}"
  if [[ ! -f "$dirInstall/$binName" ]] | [[ ! -x "$dirInstall/$binName" ]] || [[ $(stat -c%s "$dirInstall/$binName" 2>/dev/null) -eq 0 ]]; then
    curl -L --progress-bar -# -o "$dirInstall/$binName" "$urlBin"
    chmod +x "$dirInstall/$binName"
  fi
  cat << EOF > $dirInstall/$serviceName.service
    [Unit]
    Description = TorrServer - stream torrent to http
    Wants = network-online.target
    After = network.target

    [Service]
    User = root
    Group = root
    Type = simple
    NonBlocking = true
    EnvironmentFile = $dirInstall/$serviceName.config
    ExecStart = ${dirInstall}/${binName} \$DAEMON_OPTIONS
    ExecReload = /bin/kill -HUP \${MAINPID}
    ExecStop = /bin/kill -INT \${MAINPID}
    TimeoutSec = 30
    #WorkingDirectory = ${dirInstall}
    Restart = on-failure
    RestartSec = 5s
    #LimitNOFILE = 4096

    [Install]
    WantedBy = multi-user.target
EOF
  [ -z $servicePort ] && {
    [[ $lang == "en" ]] && read -p ' Change TorrServer web-port? (Yes/No) ' answer </dev/tty || read -p ' Хотите изменить порт для TorrServer? (Да/Нет) ' answer </dev/tty
    if [ "$answer" != "${answer#[YyДд]}" ] ;then
      [[ $lang == "en" ]] && read -p ' Enter port number: ' answer </dev/tty || read -p ' Введите номер порта: ' answer </dev/tty
      servicePort=$answer
    else
      servicePort="8090"
    fi
  }
  [ -z $isAuth ] && {
    [[ $lang == "en" ]] && read -p ' Enable server authorization? (Yes/No) ' answer </dev/tty || read -p ' Включить авторизацию на сервере? (Да/Нет) ' answer </dev/tty
    if [ "$answer" != "${answer#[YyДд]}" ] ;then
      isAuth=1
    else
      isAuth=0
    fi
  }
  if [[ "$isAuth" == 1 ]]; then
    [[ ! -f "$dirInstall/accs.db" ]] && {
      [[ $lang == "en" ]] && read -p ' User: ' answer </dev/tty || read -p ' Пользователь: ' answer </dev/tty
      isAuthUser=$answer
      [[ $lang == "en" ]] && read -p ' Password: ' answer </dev/tty || read -p ' Пароль: ' answer </dev/tty
      isAuthPass=$answer
      [[ $lang == "en" ]] && echo " Apply user and password - $isAuthUser:$isAuthPass" || echo " Устанавливаем логин и пароль - $isAuthUser:$isAuthPass"
      echo -e "{\n  \"$isAuthUser\": \"$isAuthPass\"\n}" > $dirInstall/accs.db
    } || {
      [[ $lang == "en" ]] && echo " Use existing auth from ${dirInstall}/accs.db" || echo " Используйте реквизиты из ${dirInstall}/accs.db для входа"
    }
    cat << EOF > $dirInstall/$serviceName.config
    DAEMON_OPTIONS="--port $servicePort --path $dirInstall --httpauth"
EOF
  else
    cat << EOF > $dirInstall/$serviceName.config
    DAEMON_OPTIONS="--port $servicePort --path $dirInstall"
EOF
  fi
  [ -z $isRdb ] && {
    [[ $lang == "en" ]] && read -p ' Start TorrServer in public read-only mode? (Yes/No) ' answer </dev/tty || read -p ' Запускать TorrServer в публичном режиме без возможности изменения настроек через веб сервера? (Да/Нет) ' answer </dev/tty
    if [ "$answer" != "${answer#[YyДд]}" ] ;then
      isRdb=1
    else
      isRdb=0
    fi
  }
  if [[ "$isRdb" == 1 ]]; then
    [[ $lang == "en" ]] && {
    echo " Set database to read-only mode…"
    echo " To change remove --rdb option from $dirInstall/$serviceName.config"
    echo " or rerun install script without parameters"
    } || {
    echo " База данных устанавливается в режим «только для чтения»…"
    echo " Для изменения отредактируйте $dirInstall/$serviceName.config, убрав опцию --rdb"
    echo " или запустите интерактивную установку без параметров повторно"
    }
    sed -i 's|DAEMON_OPTIONS="--port|DAEMON_OPTIONS="--rdb --port|' $dirInstall/$serviceName.config
  fi
  [ -z $isLog ] && {
    [[ $lang == "en" ]] && read -p ' Enable TorrServer log output to file? (Yes/No) ' answer </dev/tty || read -p ' Включить запись журнала работы TorrServer в файл? (Да/Нет) ' answer </dev/tty
    if [ "$answer" != "${answer#[YyДд]}" ] ;then
      sed -i "s|--path|--logpath $dirInstall/$serviceName.log --path|" "$dirInstall/$serviceName.config"
    fi
  }

  ln -sf $dirInstall/$serviceName.service /usr/local/lib/systemd/system/
  sed -i 's/^[ \t]*//' $dirInstall/$serviceName.service
  sed -i 's/^[ \t]*//' $dirInstall/$serviceName.config

  [[ $lang == "en" ]] && echo " Starting TorrServer…" || echo " Запускаем службу TorrServer…"
  systemctl daemon-reload 2>/dev/null
  systemctl enable --now $serviceName.service 2>/dev/null
  getIP
  [[ $lang == "en" ]] && {
    echo ""
    echo " TorrServer $(getLatestRelease) installed to ${dirInstall}"
    echo ""
    echo " You can now open your browser at http://${serverIP}:${servicePort} to access TorrServer web GUI."
    echo ""
  } || {
    echo ""
    echo " TorrServer $(getLatestRelease) установлен в директории ${dirInstall}"
    echo ""
    echo " Теперь вы можете открыть браузер по адресу http://${serverIP}:${servicePort} для доступа к вебу TorrServer"
    echo ""
  }
  if [[ "$isAuth" == 1 && $isAuthUser > 0 ]]; then
    [[ $lang == "en" ]] && echo " Use user \"$isAuthUser\" with password \"$isAuthPass\" for authentication" || echo " Для авторизации введите пользователя $isAuthUser с паролем $isAuthPass"
  echo ""
  fi
}

function checkInstalled() {
  binName="TorrServer-linux-${architecture}"
  if [[ -f "$dirInstall/$binName" ]] || [[ $(stat -c%s "$dirInstall/$binName" 2>/dev/null) -ne 0 ]]; then
    [[ $lang == "en" ]] && echo " - TorrServer found in $dirInstall" || echo " - TorrServer найден в директории $dirInstall"
  else
    [[ $lang == "en" ]] && echo " - TorrServer not found. It's not installed or have zero size." || echo " - TorrServer не найден, возможно он не установлен или размер бинарника равен 0."
    return 1
  fi
}

function checkInstalledVersion() {
  binName="TorrServer-linux-${architecture}"
  if [[ -z "$(getLatestRelease)" ]]; then
    [[ $lang == "en" ]] && echo " - No update. Can be server issue." || echo " - Не найдено обновление. Возможно сервер не доступен."
    exit 1
  fi
  if [[ "$(getLatestRelease)" == "$($dirInstall/$binName --version 2>/dev/null | awk '{print $2}')" ]]; then
    [[ $lang == "en" ]] && echo " - You have latest TorrServer $(getLatestRelease)" || echo " - Установлен TorrServer последней версии $(getLatestRelease)"
  else
  	[[ $lang == "en" ]] && {
			echo " - TorrServer update found!"
			echo "  installed: \"$($dirInstall/$binName --version 2>/dev/null | awk '{print $2}')\""
			echo "  available: \"$(getLatestRelease)\""
    } || {
			echo " - Доступно обновление сервера"
			echo "  установлен: \"$($dirInstall/$binName --version 2>/dev/null | awk '{print $2}')\""
			echo "  обновление: \"$(getLatestRelease)\""
    }
    return 1
  fi
}

function UpdateVersion() {
  systemctl stop $serviceName.service
  binName="TorrServer-linux-${architecture}"
  urlBin="https://github.com/YouROK/TorrServer/releases/download/$(getLatestRelease)/${binName}"
  curl -L --progress-bar -# -o "$dirInstall/$binName" "$urlBin"
  chmod +x "$dirInstall/$binName"
  systemctl start $serviceName.service
}

function DowngradeVersion() {
  systemctl stop $serviceName.service
  binName="TorrServer-linux-${architecture}"
  urlBin="https://github.com/YouROK/TorrServer/releases/download/MatriX.$downgradeRelease/${binName}"
  curl -L --progress-bar -# -o "$dirInstall/$binName" "$urlBin"
  chmod +x "$dirInstall/$binName"
  systemctl start $serviceName.service
}
#####################################
#     E N D   F U N C T I O N S     #
#####################################
getLang
case $1 in
  -i|--install|install)
    initialCheck
    if ! checkInstalled; then
      servicePort="8090"
      isAuth=0
      isRdb=0
      isLog=0
      installTorrServer
    fi
    exit
    ;;
  -u|--update|update)
    initialCheck
    if checkInstalled; then
      if ! checkInstalledVersion; then
        UpdateVersion
      fi
    fi
    exit
    ;;
  -c|--check|check)
    initialCheck
    if checkInstalled; then
      checkInstalledVersion
    fi
    exit
    ;;
  -d|--down|down)
    initialCheck
    downgradeRelease="$2"
    [ -z "$downgradeRelease" ] &&
      echo " Вы не указали номер версии" &&
      echo " Наберите $scriptname -d|-down|down <версия>, например $scriptname -d 101" &&
      exit 1
    if checkInstalled; then
      DowngradeVersion
    fi
    exit
    ;;
  -r|--remove|remove)
    cleanup
    exit
    ;;
  -h|--help|help)
    helpUsage
    exit
    ;;
  *)
    echo ""
    echo " Choose Language:"
    echo " [1] English"
    echo " [2] Русский"
    read -p ' Your language (Ваш язык): ' answer </dev/tty
    if [ "$answer" != "${answer#[2]}" ] ;then
      lang="ru"
    fi
    echo ""
    echo "============================================================="
    [[ $lang == "en" ]] && echo " TorrServer install and configuration script for Linux " || echo " Скрипт установки, удаления и настройки TorrServer для Linux "
    echo "============================================================="
    echo ""
    [[ $lang == "en" ]] && echo " Enter $scriptname -h or --help or help for all available commands" || echo " Наберите $scriptname -h или --help или help для вызова справки всех доступных команд"
    ;;
esac

while true; do
  echo ""
  [[ $lang == "en" ]] && read -p " Want to install or configure TorrServer? (Yes|No) Type Delete to uninstall. " ydn </dev/tty || read -p " Хотите установить, обновить или настроить TorrServer? (Да|Нет) Для удаления введите «Удалить» " ydn </dev/tty
  case $ydn in
    [YyДд]* )
      initialCheck;
      installTorrServer;
      break;;
    [DdУу]* )
      uninstall;
      break;;
    [NnНн]* )
      break;;
    * ) [[ $lang == "en" ]] && echo " Enter Yes, No or Delete" || echo " Ввведите Да, Нет или Удалить" ;;
  esac
done

echo " Have Fun!"
echo ""
sleep 3