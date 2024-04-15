#!/usr/bin/env bash
username="torrserver" # system user to add || root
dirInstall="/opt/torrserver" # путь установки torrserver
serviceName="torrserver" # имя службы: systemctl status torrserver.service
scriptname=$(basename "$(test -L "$0" && readlink "$0" || echo -e "$0")")
declare -A colors=( [black]=0 [red]=1 [green]=2 [yellow]=3 [blue]=4 [magenta]=5 [cyan]=6 [white]=7 )

#################################
#       F U N C T I O N S       #
#################################

colorize() {
    printf "%s%s%s" "$(tput setaf "${colors[$1]:-7}")" "$2" "$(tput op)"
}

function isRoot() {
  if [ $EUID -ne 0 ]; then
    return 1
  fi
}

function addUser() {
  if isRoot; then
    [[ $username == "root" ]] && return 0
    egrep "^$username" /etc/passwd >/dev/null
    if [ $? -eq 0 ]; then
      [[ $lang == "en" ]] && echo -e " - $username user exists!" || echo -e " - пользователь $username найден!"
      return 0
    else
      useradd --home-dir "$dirInstall" --create-home --shell /bin/false -c "TorrServer" "$username"
      [ $? -eq 0 ] && {
        chmod 755 "$dirInstall"
        [[ $lang == "en" ]] && echo -e " - User $username has been added to system!" || echo -e " - пользователь $username добавлен!"
      } || {
        [[ $lang == "en" ]] && echo -e " - Failed to add $username user!" || echo -e " - не удалось добавить пользователя $username!"
      }
    fi
  fi
}

function delUser() {
  if isRoot; then
    [[ $username == "root" ]] && return 0
    egrep "^$username" /etc/passwd >/dev/null
    if [ $? -eq 0 ]; then
      userdel --remove "$username" 2>/dev/null # --force 
      [ $? -eq 0 ] && {
        [[ $lang == "en" ]] && echo -e " - User $username has been removed from system!" || echo -e " - Пользователь $username удален!"
      } || {
        [[ $lang == "en" ]] && echo -e " - Failed to remove $username user!" || echo -e " - не удалось удалить пользователя $username!"
      }
    else
      [[ $lang == "en" ]] && echo -e " - $username - no such user!" || echo -e " - пользователь $username не найден!"
      return 1
    fi
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
  checkArch
  checkInstalled
  [[ $lang == "en" ]] && {
    echo -e ""
    echo -e " TorrServer install dir - ${dirInstall}"
    echo -e ""
    echo -e " This action will delete TorrServer including all it's torrents, settings and files on path above!"
    echo -e ""
  } || {
    echo -e ""
    echo -e " Директория c TorrServer - ${dirInstall}"
    echo -e ""
    echo -e " Это действие удалит все данные TorrServer включая базу данных торрентов и настройки по указанному выше пути!"
    echo -e ""
  }
  [[ $lang == "en" ]] && read -p " Are you shure you want to delete TorrServer? ($(colorize red Y)es/$(colorize yellow N)o) " answer_del </dev/tty || read -p " Вы уверены что хотите удалить программу? ($(colorize red Y)es/$(colorize yellow N)o) " answer_del </dev/tty
  if [ "$answer_del" != "${answer_del#[YyДд]}" ]; then
    cleanup
    cleanAll
    [[ $lang == "en" ]] && echo -e " - TorrServer uninstalled!" || echo -e " - TorrServer удален из системы!"
    echo -e ""
  else
    echo -e ""
  fi
}

function cleanup() {
  systemctl stop $serviceName 2>/dev/null
  systemctl disable $serviceName 2>/dev/null
  rm -rf /usr/local/lib/systemd/system/$serviceName.service $dirInstall 2>/dev/null
  delUser
}

function cleanAll() { # guess other installs
  systemctl stop torr torrserver 2>/dev/null
  systemctl disable torr torrserver 2>/dev/null
  rm -rf /home/torrserver 2>/dev/null
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
    PKGS='curl iputils-ping dnsutils'
    source /etc/os-release
    if [[ $ID == "debian" || $ID == "raspbian" ]]; then
      if [[ $VERSION_ID -lt 6 ]]; then
        echo -e " Ваша версия Debian не поддерживается."
        echo -e ""
        echo -e " Скрипт поддерживает только Debian >=6"
        echo -e ""
        exit 1
      fi
    elif [[ $ID == "ubuntu" ]]; then
      OS="ubuntu"
      MAJOR_UBUNTU_VERSION=$(echo -e "$VERSION_ID" | cut -d '.' -f1)
      if [[ $MAJOR_UBUNTU_VERSION -lt 10 ]]; then
        echo -e " Ваша версия Ubuntu не поддерживается."
        echo -e ""
        echo -e " Скрипт поддерживает только Ubuntu >=10"
        echo -e ""
        exit 1
      fi
    fi
    if ! dpkg -s $PKGS >/dev/null 2>&1; then
      [[ $lang == "en" ]] && echo -e " Installing missing packages…" || echo -e " Устанавливаем недостающие пакеты…"
      sleep 1
      apt -y install $PKGS
    fi
  elif [[ -e /etc/system-release ]]; then
    source /etc/os-release
    if [[ $ID == "fedora" || $ID_LIKE == "fedora" ]]; then
      OS="fedora"
      [ -z "$(rpm -qa curl)" ] && yum -y install curl
      [ -z "$(rpm -qa iputils)" ] && yum -y install iputils
    fi
    if [[ $ID == "centos" || $ID == "rocky" || $ID == "redhat" ]]; then
      OS="centos"
      if [[ ! $VERSION_ID =~ (6|7|8) ]]; then
        echo -e " Ваша версия CentOS/RockyLinux/RedHat не поддерживается."
        echo -e ""
        echo -e " Скрипт поддерживает только CentOS/RockyLinux/RedHat версии 6,7 и 8."
        echo -e ""
        exit 1
      fi
      [ -z "$(rpm -qa curl)" ] && yum -y install curl
      [ -z "$(rpm -qa iputils)" ] && yum -y install iputils
    fi
    if [[ $ID == "ol" ]]; then
      OS="oracle"
      if [[ ! $VERSION_ID =~ (6|7|8) ]]; then
        echo -e " Ваша версия Oracle Linux не поддерживается."
        echo -e ""
        echo -e " Скрипт поддерживает только Oracle Linux версии 6,7 и 8."
        exit 1
      fi
      [ -z "$(rpm -qa curl)" ] && yum -y install curl
      [ -z "$(rpm -qa iputils)" ] && yum -y install iputils
    fi
    if [[ $ID == "amzn" ]]; then
      OS="amzn"
      if [[ $VERSION_ID != "2" ]]; then
        echo -e " Ваша версия Amazon Linux не поддерживается."
        echo -e ""
        echo -e " Скрипт поддерживает только Amazon Linux 2."
        echo -e ""
        exit 1
      fi
      [ -z "$(rpm -qa curl)" ] && yum -y install curl
      [ -z "$(rpm -qa iputils)" ] && yum -y install iputils
    fi
  elif [[ -e /etc/arch-release ]]; then
    OS=arch
    [ -z $(pacman -Qqe curl 2>/dev/null) ] &&  pacman -Sy --noconfirm curl
    [ -z $(pacman -Qqe iputils 2>/dev/null) ] &&  pacman -Sy --noconfirm iputils
  else
    echo -e " Похоже, что вы запускаете этот установщик в системе отличной от Debian, Ubuntu, Fedora, CentOS, Amazon Linux, Oracle Linux или Arch Linux."
    exit 1
  fi
}

function checkArch() {
  case $(uname -m) in
    i386) architecture="386" ;;
    i686) architecture="386" ;;
    x86_64) architecture="amd64" ;;
    aarch64) architecture="arm64" ;;
    armv7|armv7l) architecture="arm7" ;;
    armv6|armv6l) architecture="arm5" ;;
    *) [[ $lang == "en" ]] && { echo -e " Unsupported Arch. Can't continue."; exit 1; } || { echo -e " Не поддерживаемая архитектура. Продолжение невозможно."; exit 1; } ;;
  esac
}

function checkInternet() {
  [ -z "`which ping`" ] && echo -e " Сначала установите iputils-ping" && exit 1
  [[ $lang == "en" ]] && echo -e " Check Internet access…" || echo -e " Проверяем соединение с Интернетом…"
  if ! ping -c 2 google.com &> /dev/null; then
    [[ $lang == "en" ]] && echo -e " - No Internet. Check your network and DNS settings." || echo -e " - Нет Интернета. Проверьте ваше соединение, а также разрешение имен DNS."
    exit 1
  fi
  [[ $lang == "en" ]] && echo -e " - Have Internet Access" || echo -e " - соединение с Интернетом успешно"
}

function initialCheck() {
  if ! isRoot; then
    [[ $lang == "en" ]] && echo -e " Script must run as root or user with sudo privileges. Example: sudo $scriptname" || echo -e " Вам нужно запустить скрипт от root или пользователя с правами sudo. Пример: sudo $scriptname"
    exit 1
  fi
  # [ -z "`which curl`" ] && echo -e " Сначала установите curl" && exit 1
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
  [[ $lang == "en" ]] && echo -e " Install and configure TorrServer…" || echo -e " Устанавливаем и настраиваем TorrServer…"
  if checkInstalled; then
    if ! checkInstalledVersion; then
      [[ $lang == "en" ]] && read -p " Want to update TorrServer? ($(colorize green Y)es/$(colorize yellow N)o) " answer_up </dev/tty || read -p " Хотите обновить TorrServer? ($(colorize green Y)es/$(colorize yellow N)o) " answer_up </dev/tty
      if [ "$answer_up" != "${answer_up#[YyДд]}" ]; then
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
    User = $username
    Group = $username
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
    [[ $lang == "en" ]] && read -p " Change TorrServer web-port? ($(colorize yellow Y)es/$(colorize green N)o) " answer_cp </dev/tty || read -p " Хотите изменить порт для TorrServer? ($(colorize yellow Y)es/$(colorize green N)o) " answer_cp </dev/tty
    if [ "$answer_cp" != "${answer_cp#[YyДд]}" ]; then
      [[ $lang == "en" ]] && read -p " Enter port number: " answer_port </dev/tty || read -p " Введите номер порта: " answer_port </dev/tty
      servicePort=$answer_port
    else
      servicePort="8090"
    fi
  }
  [ -z $isAuth ] && {
    [[ $lang == "en" ]] && read -p " Enable server authorization? ($(colorize green Y)es/$(colorize yellow N)o) " answer_auth </dev/tty || read -p " Включить авторизацию на сервере? ($(colorize green Y)es/$(colorize yellow N)o) " answer_auth </dev/tty
    if [ "$answer_auth" != "${answer_auth#[YyДд]}" ]; then
      isAuth=1
    else
      isAuth=0
    fi
  }
  if [ $isAuth -eq 1 ]; then
    [[ ! -f "$dirInstall/accs.db" ]] && {
      [[ $lang == "en" ]] && read -p " User: " answer_user </dev/tty || read -p " Пользователь: " answer_user </dev/tty
      isAuthUser=$answer_user
      [[ $lang == "en" ]] && read -p " Password: " answer_pass </dev/tty || read -p " Пароль: " answer_pass </dev/tty
      isAuthPass=$answer_pass
      [[ $lang == "en" ]] && echo -e " Store $isAuthUser:$isAuthPass to ${dirInstall}/accs.db" || echo -e " Сохраняем $isAuthUser:$isAuthPass в ${dirInstall}/accs.db"
      echo -e "{\n  \"$isAuthUser\": \"$isAuthPass\"\n}" > $dirInstall/accs.db
    } || {
    	auth=$(cat "$dirInstall/accs.db"|head -2|tail -1|tr -d '[:space:]'|tr -d '"')
      [[ $lang == "en" ]] && echo -e " - Use existing auth from ${dirInstall}/accs.db - $auth" || echo -e " - Используйте реквизиты из ${dirInstall}/accs.db для авторизации - $auth"
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
    [[ $lang == "en" ]] && read -p " Start TorrServer in public read-only mode? ($(colorize yellow Y)es/$(colorize green N)o) " answer_rdb </dev/tty || read -p " Запускать TorrServer в публичном режиме без возможности изменения настроек через веб сервера? ($(colorize yellow Y)es/$(colorize green N)o) " answer_rdb </dev/tty
    if [ "$answer_rdb" != "${answer_rdb#[YyДд]}" ]; then
      isRdb=1
    else
      isRdb=0
    fi
  }
  if [ $isRdb -eq 1 ]; then
    [[ $lang == "en" ]] && {
      echo -e " Set database to read-only mode…"
      echo -e " To change remove --rdb option from $dirInstall/$serviceName.config"
      echo -e " or rerun install script without parameters"
    } || {
      echo -e " База данных устанавливается в режим «только для чтения»…"
      echo -e " Для изменения отредактируйте $dirInstall/$serviceName.config, убрав опцию --rdb"
      echo -e " или запустите интерактивную установку без параметров повторно"
    }
    sed -i 's|DAEMON_OPTIONS="--port|DAEMON_OPTIONS="--rdb --port|' $dirInstall/$serviceName.config
  fi
  [ -z $isLog ] && {
    [[ $lang == "en" ]] && read -p " Enable TorrServer log output to file? ($(colorize yellow Y)es/$(colorize green N)o) " answer_log </dev/tty || read -p " Включить запись журнала работы TorrServer в файл? ($(colorize yellow Y)es/$(colorize green N)o) " answer_log </dev/tty
    if [ "$answer_log" != "${answer_log#[YyДд]}" ]; then
      sed -i "s|--path|--logpath $dirInstall/$serviceName.log --path|" "$dirInstall/$serviceName.config"
      [[ $lang == "en" ]] && echo -e " - TorrServer log stored at $dirInstall/$serviceName.log" || echo -e " - лог TorrServer располагается по пути $dirInstall/$serviceName.log"
    fi
  }

  ln -sf $dirInstall/$serviceName.service /usr/local/lib/systemd/system/
  sed -i 's/^[ \t]*//' $dirInstall/$serviceName.service
  sed -i 's/^[ \t]*//' $dirInstall/$serviceName.config

  [[ $lang == "en" ]] && echo -e " Starting TorrServer…" || echo -e " Запускаем службу TorrServer…"
  systemctl daemon-reload 2>/dev/null
  systemctl enable $serviceName.service 2>/dev/null # enable --now
  systemctl restart $serviceName.service 2>/dev/null
  getIP
  [[ $lang == "en" ]] && {
    echo -e ""
    echo -e " TorrServer $(getLatestRelease) installed to ${dirInstall}"
    echo -e ""
    echo -e " You can now open your browser at http://${serverIP}:${servicePort} to access TorrServer web GUI."
    echo -e ""
  } || {
    echo -e ""
    echo -e " TorrServer $(getLatestRelease) установлен в директории ${dirInstall}"
    echo -e ""
    echo -e " Теперь вы можете открыть браузер по адресу http://${serverIP}:${servicePort} для доступа к вебу TorrServer"
    echo -e ""
  }
  if [[ $isAuth -eq 1 && $isAuthUser > 0 ]]; then
    [[ $lang == "en" ]] && echo -e " Use user \"$isAuthUser\" with password \"$isAuthPass\" for authentication" || echo -e " Для авторизации используйте пользователя «$isAuthUser» с паролем «$isAuthPass»"
  echo -e ""
  fi
}

function checkInstalled() {
  if ! addUser; then
    username="root"
  fi
  binName="TorrServer-linux-${architecture}"
  if [[ -f "$dirInstall/$binName" ]] || [[ $(stat -c%s "$dirInstall/$binName" 2>/dev/null) -ne 0 ]]; then
    [[ $lang == "en" ]] && echo -e " - TorrServer found in $dirInstall" || echo -e " - TorrServer найден в директории $dirInstall"
  else
    [[ $lang == "en" ]] && echo -e " - TorrServer not found. It's not installed or have zero size." || echo -e " - TorrServer не найден, возможно он не установлен или размер бинарника равен 0."
    return 1
  fi
}

function checkInstalledVersion() {
  binName="TorrServer-linux-${architecture}"
  if [[ -z "$(getLatestRelease)" ]]; then
    [[ $lang == "en" ]] && echo -e " - No update. Can be server issue." || echo -e " - Не найдено обновление. Возможно сервер не доступен."
    exit 1
  fi
  if [[ "$(getLatestRelease)" == "$($dirInstall/$binName --version 2>/dev/null | awk '{print $2}')" ]]; then
    [[ $lang == "en" ]] && echo -e " - You have latest TorrServer $(getLatestRelease)" || echo -e " - Установлен TorrServer последней версии $(getLatestRelease)"
  else
    [[ $lang == "en" ]] && {
      echo -e " - TorrServer update found!"
      echo -e "   installed: \"$($dirInstall/$binName --version 2>/dev/null | awk '{print $2}')\""
      echo -e "   available: \"$(getLatestRelease)\""
    } || {
      echo -e " - Доступно обновление сервера"
      echo -e "   установлен: \"$($dirInstall/$binName --version 2>/dev/null | awk '{print $2}')\""
      echo -e "   обновление: \"$(getLatestRelease)\""
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
    else
      systemctl stop $serviceName.service
      systemctl start $serviceName.service
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
      echo -e " Вы не указали номер версии" &&
      echo -e " Наберите $scriptname -d|-down|down <версия>, например $scriptname -d 101" &&
      exit 1
    if checkInstalled; then
      DowngradeVersion
    fi
    exit
    ;;
  -r|--remove|remove)
    uninstall
    exit
    ;;
  -h|--help|help)
    helpUsage
    exit
    ;;
  *)
    echo -e ""
    echo -e " Choose Language:"
    echo -e " [$(colorize green 1)] English"
    echo -e " [$(colorize yellow 2)] Русский"
    read -p " Your language (Ваш язык): " answer_lang </dev/tty
    if [ "$answer_lang" != "${answer_lang#[2]}" ]; then
      lang="ru"
    fi
    echo -e ""
    echo -e "============================================================="
    [[ $lang == "en" ]] && echo -e " TorrServer install and configuration script for Linux " || echo -e " Скрипт установки, удаления и настройки TorrServer для Linux "
    echo -e "============================================================="
    echo -e ""
    [[ $lang == "en" ]] && echo -e " Enter $scriptname -h or --help or help for all available commands" || echo -e " Наберите $scriptname -h или --help или help для вызова справки всех доступных команд"
    ;;
esac

while true; do
  echo -e ""
  [[ $lang == "en" ]] && read -p " Want to install or configure TorrServer? ($(colorize green Y)es|$(colorize yellow N)o) Type $(colorize red D)elete to uninstall. " ydn </dev/tty || read -p " Хотите установить, обновить или настроить TorrServer? ($(colorize green Y)es|$(colorize yellow N)o) Для удаления введите «$(colorize red D)elete» " ydn </dev/tty
  case $ydn in
    [YyДд]*)
      initialCheck
      installTorrServer
      break
      ;;
    [DdУу]*)
      uninstall
      break
      ;;
    [NnНн]*)
      break
      ;;
    *) [[ $lang == "en" ]] && echo -e " Enter $(colorize green Y)es, $(colorize yellow N)o or $(colorize red D)elete" || echo -e " Ввведите $(colorize green Y)es, $(colorize yellow N)o или $(colorize red D)elete"
    	;;
  esac
done

echo -e " Have Fun!"
echo -e ""
sleep 3