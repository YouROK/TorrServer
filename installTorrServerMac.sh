#!/usr/bin/env bash

set -euo pipefail

#############################################
#           GLOBAL VARIABLES
#############################################

# Installation settings
DEFAULT_INSTALL_DIR="/Users/Shared/TorrServer"
DEFAULT_SERVICE_NAME="torrserver"
DEFAULT_PORT="8090"

# Runtime variables
dirInstall="${DEFAULT_INSTALL_DIR}"
serviceName="${DEFAULT_SERVICE_NAME}"
scriptname=$(basename "$(test -L "$0" && readlink "$0" || echo "$0")")

# Flags
SILENT_MODE=0
USE_USER_LAUNCHAGENT=0
USER_PROMPTED=0

# Command-line state
parsedCommand=""
specificVersion=""
downgradeRelease=""

# Service configuration
servicePort=""
isAuth=""
isRdb=""
isLog=""
isAuthUser=""
isAuthPass=""
sysPath=""

# Constants
readonly REPO_URL="https://github.com/YouROK/TorrServer"
readonly REPO_API_URL="https://api.github.com/repos/YouROK/TorrServer"
readonly VERSION_PREFIX="MatriX"
readonly BINARY_NAME_PREFIX="TorrServer-darwin"

# Color support
getColorCode() {
  case "$1" in
    black) echo 0 ;;
    red) echo 1 ;;
    green) echo 2 ;;
    yellow) echo 3 ;;
    blue) echo 4 ;;
    magenta) echo 5 ;;
    cyan) echo 6 ;;
    white) echo 7 ;;
    *) echo 7 ;; # default to white
  esac
}
supports_color_output=0

if command -v tput >/dev/null 2>&1 && [[ -t 1 ]]; then
  if [[ $(tput colors 2>/dev/null) -ge 8 ]]; then
    supports_color_output=1
  fi
fi

# Language
lang="en"

#############################################
#     TRANSLATION SYSTEM
#############################################

# Message dictionary
# Message dictionary - bash 3.2 compatible (no associative arrays)
# English messages
MSG_EN_lang_choice="Choose Language:"
MSG_EN_lang_english="English"
MSG_EN_lang_russian="Русский"
MSG_EN_your_lang="Your language (Ваш язык): "
MSG_EN_have_fun="Have Fun!"
MSG_EN_script_title="TorrServer install and configuration script for macOS"
MSG_EN_unsupported_arch="Unsupported Arch. Can't continue."
MSG_EN_unsupported_os="It looks like you are running this installer on a system other than macOS."
MSG_EN_downloading="Downloading TorrServer"
MSG_EN_target_version="Target version:"
MSG_EN_installed_version="installed:"
MSG_EN_target_label="target:"
MSG_EN_version_not_found="ERROR: Version %s not found in releases"
MSG_EN_check_versions="Please check available versions at: $REPO_URL/releases"
MSG_EN_already_installed="You already have TorrServer %s installed"
MSG_EN_have_latest="You have latest TorrServer %s"
MSG_EN_update_found="TorrServer update found!"
MSG_EN_will_install="Will install TorrServer version %s"
MSG_EN_installing_packages="Installing missing packages…"
MSG_EN_install_configure="Install and configure TorrServer…"
MSG_EN_starting_service="Starting TorrServer…"
MSG_EN_install_complete="TorrServer %s installed to %s"
MSG_EN_access_web="You can now open your browser at http://%s:%s to access TorrServer web GUI."
MSG_EN_use_auth="Use user \"%s\" with password \"%s\" for authentication"
MSG_EN_want_update="Want to update TorrServer?"
MSG_EN_want_install="Want to install or configure TorrServer? Type Delete to uninstall."
MSG_EN_want_reconfigure="Do you want to reconfigure TorrServer settings?"
MSG_EN_change_port="Change TorrServer web-port?"
MSG_EN_enter_port="Enter port number: "
MSG_EN_enable_auth="Enable server authorization?"
MSG_EN_prompt_user="User: "
MSG_EN_prompt_password="Password: "
MSG_EN_change_auth_credentials="Change authentication username and password?"
MSG_EN_enable_rdb="Start TorrServer in public read-only mode?"
MSG_EN_enable_log="Enable TorrServer log output to file?"
MSG_EN_confirm_delete="Are you sure you want to delete TorrServer?"
MSG_EN_prompt_launchagent="Add autostart for current user (1) or all users (2)?"
MSG_EN_admin_password="System can ask your admin account password"
MSG_EN_install_dir_label="TorrServer install dir -"
MSG_EN_uninstall_warning="This action will delete TorrServer including all it's torrents, settings and files on path above!"
MSG_EN_uninstalled="TorrServer uninstalled!"
MSG_EN_found_in="TorrServer found in"
MSG_EN_not_found="TorrServer not found. It's not installed or have zero size."
MSG_EN_no_version_info="No version information available. Can be server issue."
MSG_EN_config_updated="Configuration updated successfully"
MSG_EN_store_auth="Store %s:%s to %s"
MSG_EN_use_existing_auth="Use existing auth from %s - %s"
MSG_EN_set_readonly="Set database to read-only mode…"
MSG_EN_readonly_hint="To change remove --rdb option from %s or rerun install script without parameters"
MSG_EN_log_location="TorrServer log stored at %s"
MSG_EN_service_added="Autostart service added to %s"
MSG_EN_launchctl_missing="launchctl is not available. Skipping service management commands."
MSG_EN_launchctl_failed="Warning: launchctl %s failed"
MSG_EN_service_start_failed="Warning: TorrServer service failed to start. Check launchctl list for details."
MSG_EN_error_version_required="Error: Version number required for downgrade"
MSG_EN_error_version_example="Example: %s -d 101"
MSG_EN_error_unknown_option="Unknown option: %s"
MSG_EN_installing_specific_version="Installing specific version: %s"
MSG_EN_install_first_required="Please install TorrServer first using: %s --install"

# Russian messages
MSG_RU_lang_choice="Choose Language:"
MSG_RU_lang_english="English"
MSG_RU_lang_russian="Русский"
MSG_RU_your_lang="Your language (Ваш язык): "
MSG_RU_have_fun="Have Fun!"
MSG_RU_script_title="Скрипт установки, удаления и настройки TorrServer для macOS"
MSG_RU_unsupported_arch="Не поддерживаемая архитектура. Продолжение невозможно."
MSG_RU_unsupported_os="Похоже, что вы запускаете этот установщик в системе отличной от macOS."
MSG_RU_downloading="Загружаем TorrServer"
MSG_RU_target_version="Устанавливаемая версия:"
MSG_RU_installed_version="установлен:"
MSG_RU_target_label="устанавливаемая:"
MSG_RU_version_not_found="ОШИБКА: Версия %s не найдена в релизах"
MSG_RU_check_versions="Проверьте доступные версии по адресу: $REPO_URL/releases"
MSG_RU_already_installed="TorrServer %s уже установлен"
MSG_RU_have_latest="Установлен TorrServer последней версии %s"
MSG_RU_update_found="Доступно обновление сервера"
MSG_RU_will_install="Будет установлена версия TorrServer %s"
MSG_RU_installing_packages="Устанавливаем недостающие пакеты…"
MSG_RU_install_configure="Устанавливаем и настраиваем TorrServer…"
MSG_RU_starting_service="Запускаем службу TorrServer…"
MSG_RU_install_complete="TorrServer %s установлен в директории %s"
MSG_RU_access_web="Теперь вы можете открыть браузер по адресу http://%s:%s для доступа к вебу TorrServer"
MSG_RU_use_auth="Для авторизации используйте пользователя «%s» с паролем «%s»"
MSG_RU_want_update="Хотите обновить TorrServer?"
MSG_RU_want_install="Хотите установить, обновить или настроить TorrServer? Для удаления введите «Delete»"
MSG_RU_want_reconfigure="Хотите перенастроить параметры TorrServer?"
MSG_RU_change_port="Хотите изменить порт для TorrServer?"
MSG_RU_enter_port="Введите номер порта: "
MSG_RU_enable_auth="Включить авторизацию на сервере?"
MSG_RU_prompt_user="Пользователь: "
MSG_RU_prompt_password="Пароль: "
MSG_RU_change_auth_credentials="Изменить имя пользователя и пароль для авторизации?"
MSG_RU_enable_rdb="Запускать TorrServer в публичном режиме без возможности изменения настроек через веб сервера?"
MSG_RU_enable_log="Включить запись журнала работы TorrServer в файл?"
MSG_RU_confirm_delete="Вы уверены что хотите удалить программу?"
MSG_RU_prompt_launchagent="Добавить автозагрузку для текущего пользователя (1) или для всех (2)?"
MSG_RU_admin_password="Система может запросить ваш пароль администратора"
MSG_RU_install_dir_label="Директория c TorrServer -"
MSG_RU_uninstall_warning="Это действие удалит все данные TorrServer включая базу данных торрентов и настройки по указанному выше пути!"
MSG_RU_uninstalled="TorrServer удален из системы!"
MSG_RU_found_in="TorrServer найден в директории"
MSG_RU_not_found="TorrServer не найден, возможно он не установлен или размер бинарника равен 0."
MSG_RU_no_version_info="Информация о версии недоступна. Возможно сервер не доступен."
MSG_RU_config_updated="Конфигурация успешно обновлена"
MSG_RU_store_auth="Сохраняем %s:%s в %s"
MSG_RU_use_existing_auth="Используйте реквизиты из %s для авторизации - %s"
MSG_RU_set_readonly="База данных устанавливается в режим «только для чтения»…"
MSG_RU_readonly_hint="Для изменения отредактируйте %s, убрав опцию --rdb или запустите интерактивную установку без параметров повторно"
MSG_RU_log_location="лог TorrServer располагается по пути %s"
MSG_RU_service_added="Сервис автозагрузки записан в %s"
MSG_RU_launchctl_missing="launchctl недоступен. Пропускаем команды управления службой."
MSG_RU_launchctl_failed="Предупреждение: команда launchctl %s завершилась ошибкой"
MSG_RU_service_start_failed="Предупреждение: служба TorrServer не запустилась. Проверьте launchctl list для деталей."
MSG_RU_error_version_required="Ошибка: Требуется номер версии для понижения версии"
MSG_RU_error_version_example="Пример: %s -d 101"
MSG_RU_error_unknown_option="Неизвестная опция: %s"
MSG_RU_installing_specific_version="Установка конкретной версии: %s"
MSG_RU_install_first_required="Пожалуйста, сначала установите TorrServer используя: %s --install"

# Translation function - bash 3.2 compatible
msg() {
  local key="$1"
  shift
  local var_name
  local message=""

  if [[ $lang == "ru" ]]; then
    var_name="MSG_RU_${key}"
  else
    var_name="MSG_EN_${key}"
  fi

  # Use eval to get the variable value (bash 3.2 compatible)
  eval "message=\"\${${var_name}:-${key}}\""

  # Apply printf formatting if additional arguments provided
  if [[ $# -gt 0 ]]; then
    # shellcheck disable=SC2059
    printf "$message" "$@"
  else
    printf '%s\n' "$message"
  fi
}

#############################################
#     UTILITY FUNCTIONS
#############################################

colorize() {
  if [[ $supports_color_output -eq 1 ]]; then
    local color_code
    color_code=$(getColorCode "$1")
    printf "%s%s%s" "$(tput setaf "$color_code")" "$2" "$(tput op)"
  else
    printf "%s" "$2"
  fi
}

# Highlight first letter of a word with specified color
highlightFirstLetter() {
  local color="$1"
  local word="$2"
  local first_char="${word:0:1}"
  local rest="${word:1}"
  printf "%s%s" "$(colorize "$color" "$first_char")" "$rest"
}

getBinaryName() {
  echo "${BINARY_NAME_PREFIX}-${architecture}"
}

getVersionTag() {
  local version="$1"
  echo "${VERSION_PREFIX}.${version}"
}

buildDownloadUrl() {
  local target_version="$1"
  local binary_name="$2"

  if [[ "$target_version" == "latest" ]]; then
    echo "${REPO_URL}/releases/latest/download/${binary_name}"
  else
    echo "${REPO_URL}/releases/download/${target_version}/${binary_name}"
  fi
}

getLang() {
  lang=$(locale | grep LANG | cut -d= -f2 | tr -d '"' | cut -d_ -f1)
  if [[ $lang != "ru" ]]; then
    lang="en"
  fi
}

getIP() {
  local ip="localhost"

  # Try to get local IP address from network interfaces
  # On macOS, try ipconfig first (most reliable)
  if command -v ipconfig >/dev/null 2>&1; then
    # Try common interfaces: en0 (Ethernet/WiFi), en1, etc.
    for interface in en0 en1 eth0; do
      ip=$(ipconfig getifaddr "$interface" 2>/dev/null || echo "")
      if [[ -n "$ip" ]] && [[ "$ip" != "127.0.0.1" ]]; then
        break
      fi
    done
  fi

  # Fallback to ifconfig if ipconfig didn't work
  if [[ -z "$ip" ]] || [[ "$ip" == "127.0.0.1" ]]; then
    if command -v ifconfig >/dev/null 2>&1; then
      # Get the first non-loopback inet address
      ip=$(ifconfig 2>/dev/null | grep -E "inet " | grep -v "127.0.0.1" | head -n1 | awk '{print $2}' | sed 's/addr://' || echo "")
    fi
  fi

  # If still no valid IP, use localhost
  if [[ -z "$ip" ]] || [[ "$ip" == "127.0.0.1" ]]; then
    ip="localhost"
  fi

  serverIP="$ip"
}

promptYesNo() {
  local prompt="$1"
  local default="${2:-n}"
  local recommended="${3:-$default}"

  if [[ $SILENT_MODE -eq 1 ]]; then
    if [[ "$default" == "y" ]]; then
      return 0
    else
      return 1
    fi
  fi

  # Determine colors based on recommendation
  local yes_color no_color
  if [[ "$recommended" == "y" ]]; then
    yes_color="green"
    no_color="red"
  else
    yes_color="red"
    no_color="green"
  fi

  # Define localized Yes/No words
  local yes_word no_word
  if [[ $lang == "ru" ]]; then
    yes_word="Да"
    no_word="Нет"
  else
    yes_word="Yes"
    no_word="No"
  fi

  # Highlight first letter of each word
  local yes_text
  local no_text
  yes_text="$(highlightFirstLetter "$yes_color" "$yes_word")"
  no_text="$(highlightFirstLetter "$no_color" "$no_word")"

  local answer
  IFS= read -r -p " $prompt ($yes_text/$no_text) " answer </dev/tty

  # Support both English (Yy) and Russian (Дд) for Yes
  if [[ "$answer" =~ ^[YyДд] ]]; then
    return 0
  else
    return 1
  fi
}

promptYesNoDelete() {
  local prompt="$1"
  local default="${2:-n}"
  local recommended="${3:-$default}"

  if [[ $SILENT_MODE -eq 1 ]]; then
    if [[ "$default" == "y" ]]; then
      echo "yes"
    else
      echo "no"
    fi
    return
  fi

  # Determine colors based on recommended answer
  local yes_color no_color
  if [[ "$recommended" == "y" ]]; then
    yes_color="green"
    no_color="red"
  else
    yes_color="red"
    no_color="green"
  fi

  # Define localized Yes/No words
  local yes_word no_word
  if [[ $lang == "ru" ]]; then
    yes_word="Да"
    no_word="Нет"
  else
    yes_word="Yes"
    no_word="No"
  fi

  # Highlight first letter of each word
  local yes_text
  local no_text
  yes_text="$(highlightFirstLetter "$yes_color" "$yes_word")"
  no_text="$(highlightFirstLetter "$no_color" "$no_word")"

  local answer
  IFS= read -r -p " $prompt ($yes_text/$no_text) " answer </dev/tty
  answer=$(echo "$answer" | tr '[:upper:]' '[:lower:]' | xargs)

  # Check for Delete (case-insensitive, supports both English and Russian)
  if [[ "$answer" == "delete" ]] || [[ "$answer" == "удалить" ]] || [[ "$answer" == "удаление" ]]; then
    echo "delete"
  elif [[ "$answer" =~ ^[YyДд] ]]; then
    echo "yes"
  else
    echo "no"
  fi
}

promptInput() {
  local prompt="$1"
  local default="$2"

  if [[ $SILENT_MODE -eq 1 ]]; then
    echo "$default"
    return
  fi

  local answer
  IFS= read -r -p " $prompt " answer </dev/tty
  echo "${answer:-$default}"
}

launchctlCmd() {
  local quiet=0
  if [[ ${1:-} == "--quiet" ]]; then
    quiet=1
    shift
  fi

  if ! command -v launchctl >/dev/null 2>&1; then
    if [[ $quiet -eq 0 && $SILENT_MODE -eq 0 ]]; then
      echo " - $(msg launchctl_missing)"
    fi
    return 1
  fi

  local rc
  launchctl "$@" >/dev/null 2>&1
  rc=$?
  if [[ $rc -ne 0 ]]; then
    if [[ $quiet -eq 0 && $SILENT_MODE -eq 0 ]]; then
      printf " - %s\n" "$(msg launchctl_failed "$*")"
    fi
    return $rc
  fi

  return 0
}

killRunning() {
  local self="$(basename "$0")"
  local runningPid
  runningPid=$(ps -ax | grep -i torrserver | grep -v grep | grep -v "$self" | awk '{print $1}' || echo "")
  if [[ -n "$runningPid" ]]; then
    sudo kill -9 "$runningPid" 2>/dev/null || true
  fi
}

#############################################
#     VERSION MANAGEMENT
#############################################

getLatestRelease() {
  curl -s "${REPO_API_URL}/releases/latest" |
  grep -iE '"tag_name":|"version":' |
  sed -E 's/.*"([^"]+)".*/\1/' |
  head -n1
}

getSpecificRelease() {
  local version="$1"
  local tag_name
  tag_name=$(getVersionTag "$version")
  local response
  response=$(curl -s "${REPO_API_URL}/releases/tags/$tag_name")

  if echo "$response" | grep -q '"tag_name"'; then
    echo "$tag_name"
  else
    echo ""
  fi
}

getTargetVersion() {
  if [[ -n "$specificVersion" ]]; then
    local target_release
    target_release=$(getSpecificRelease "$specificVersion")
    if [[ -z "$target_release" ]]; then
      echo " - $(colorize red "$(msg version_not_found "$specificVersion")")"
      echo " - $(msg check_versions)"
      exit 1
    fi
    echo "$target_release"
  else
    getLatestRelease
  fi
}

downloadBinary() {
  local url="$1"
  local destination="$2"
  local version_info="$3"

  local curl_args=(-L)

  if [[ $SILENT_MODE -eq 0 ]]; then
    echo " - $(msg downloading) $version_info..."
    curl_args+=(--progress-bar -#)
  else
    curl_args+=(-s -S)
  fi

  curl "${curl_args[@]}" -o "$destination" "$url"
  chmod +x "$destination"
  xattr -r -d com.apple.quarantine "$destination" 2>/dev/null || true
}

#############################################
#     OS DETECTION & ARCHITECTURE
#############################################

checkOS() {
  if [[ "$(uname)" != "Darwin" ]]; then
    echo " $(msg unsupported_os)"
    exit 1
  fi
}

checkArch() {
  case $(uname -m) in
    i386|i686) architecture="386" ;;
    x86_64) architecture="amd64" ;;
    aarch64|arm64) architecture="arm64" ;;
    *)
      echo " $(msg unsupported_arch)"
      exit 1
      ;;
  esac
}

initialCheck() {
  checkOS
  checkArch
}

#############################################
#     INSTALLATION FUNCTIONS
#############################################

checkInstalled() {
  local binName
  binName=$(getBinaryName)
  if [[ -f "$dirInstall/$binName" ]] && [[ $(stat -f%z "$dirInstall/$binName" 2>/dev/null) -ne 0 ]]; then
    if [[ $SILENT_MODE -eq 0 ]]; then
      echo " - $(msg found_in) $dirInstall"
    fi
    return 0
  else
    if [[ $SILENT_MODE -eq 0 ]]; then
      echo " - $(msg not_found)"
    fi
    return 1
  fi
}

checkInstalledVersion() {
  local binName
  binName=$(getBinaryName)
  local target_version
  target_version=$(getTargetVersion)
  local installed_version
  installed_version="$("$dirInstall/$binName" --version 2>/dev/null | awk '{print $2}')"

  if [[ -z "$target_version" ]]; then
    echo " - $(msg no_version_info)"
    exit 1
  fi

  if [[ "$target_version" == "$installed_version" ]]; then
    if [[ -n "$specificVersion" ]]; then
      if [[ $SILENT_MODE -eq 0 ]]; then
        echo " - $(msg already_installed "$target_version")"
      fi
    else
      if [[ $SILENT_MODE -eq 0 ]]; then
        echo " - $(msg have_latest "$target_version")"
      fi
    fi
    return 0
  else
    if [[ $SILENT_MODE -eq 0 ]]; then
      if [[ -n "$specificVersion" ]]; then
        echo " - $(msg will_install "$target_version")"
      else
        echo " - $(msg update_found)"
      fi
      echo "   $(msg installed_version) \"$installed_version\""
      echo "   $(msg target_label) \"$target_version\""
    fi
    return 1
  fi
}

createPlistFile() {
  local daemon_options="--port $servicePort --path $dirInstall"

  if [[ $isRdb -eq 1 ]]; then
    daemon_options="$daemon_options --rdb"
  fi

  if [[ $isLog -eq 1 ]]; then
    daemon_options="$daemon_options --logpath $dirInstall/$serviceName.log"
  fi

  if [[ $isAuth -eq 1 ]]; then
    daemon_options="$daemon_options --httpauth"
  fi

  # Convert daemon_options to plist array format
  local plist_args=()
  local arg
  for arg in $daemon_options; do
    plist_args+=("    <string>$arg</string>")
  done
  local plist_args_str
  plist_args_str=$(printf '%s\n' "${plist_args[@]}")

  cat << EOF > "$dirInstall/$serviceName.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>${serviceName}</string>
  <key>ServiceDescription</key>
  <string>TorrServer service for macOS</string>
  <key>ProgramArguments</key>
  <array>
    <string>${dirInstall}/$(getBinaryName)</string>
${plist_args_str}
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <dict>
    <key>SuccessfulExit</key>
    <false/>
  </dict>
  <key>ProcessType</key>
  <string>Background</string>
  <key>ThrottleInterval</key>
  <integer>10</integer>
  <key>AbandonProcessGroup</key>
  <true/>
  <key>StandardOutPath</key>
  <string>${dirInstall}/torrserver.log</string>
  <key>StandardErrorPath</key>
  <string>${dirInstall}/torrserver.log</string>
  <key>WorkingDirectory</key>
  <string>${dirInstall}</string>
</dict>
</plist>
EOF
}

readExistingConfig() {
  local plist_file="$dirInstall/$serviceName.plist"

  if [[ -f "$plist_file" ]]; then
    # Extract port
    if grep -q "<string>--port</string>" "$plist_file"; then
      servicePort=$(grep -A1 "<string>--port</string>" "$plist_file" | tail -n1 | sed 's/.*<string>\(.*\)<\/string>.*/\1/')
    fi

    # Check for auth
    if grep -q "<string>--httpauth</string>" "$plist_file"; then
      isAuth=1
    else
      isAuth=0
    fi

    # Check for rdb
    if grep -q "<string>--rdb</string>" "$plist_file"; then
      isRdb=1
    else
      isRdb=0
    fi

    # Check for log
    if grep -q "<string>--logpath</string>" "$plist_file"; then
      isLog=1
    else
      isLog=0
    fi
  fi
}

configureService() {
  # Read existing config if available (for reconfiguration)
  if [[ -f "$dirInstall/$serviceName.plist" ]]; then
    readExistingConfig
  fi

  # Port configuration
  if [[ -z "$servicePort" ]]; then
    local inferred_default="$DEFAULT_PORT"
    if promptYesNo "$(msg change_port)" "n" "y"; then
      servicePort=$(promptInput "$(msg enter_port)" "$inferred_default")
    else
      servicePort="$inferred_default"
    fi
  else
    # Port exists, ask if user wants to change it
    if [[ $SILENT_MODE -eq 0 ]]; then
      if promptYesNo "$(msg change_port)" "n" "y"; then
        servicePort=$(promptInput "$(msg enter_port)" "$servicePort")
      fi
    fi
  fi

  # Auth configuration
  if [[ -z "$isAuth" ]]; then
    if promptYesNo "$(msg enable_auth)" "n" "y"; then
      isAuth=1
    else
      isAuth=0
    fi
  else
    # Auth setting exists, ask if user wants to change it
    if [[ $SILENT_MODE -eq 0 ]]; then
      local current_auth_default
      current_auth_default="$([[ $isAuth -eq 1 ]] && echo 'y' || echo 'n')"
      if promptYesNo "$(msg enable_auth)" "$current_auth_default" "y"; then
        isAuth=1
      else
        isAuth=0
      fi
    fi
  fi

  # Setup auth if enabled
  if [[ $isAuth -eq 1 ]]; then
    if [[ ! -f "$dirInstall/accs.db" ]]; then
      isAuthUser=$(promptInput "$(msg prompt_user)" "admin")
      isAuthPass=$(promptInput "$(msg prompt_password)" "admin")
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf ' %s\n' "$(msg store_auth "$isAuthUser" "$isAuthPass" "${dirInstall}/accs.db")"
      fi
      echo -e "{\n  \"$isAuthUser\": \"$isAuthPass\"\n}" > "$dirInstall/accs.db"
    else
      local auth
      auth=$(cat "$dirInstall/accs.db" | head -2 | tail -1 | tr -d '[:space:]' | tr -d '"')
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf ' - %s\n' "$(msg use_existing_auth "${dirInstall}/accs.db" "$auth")"
        # Ask if user wants to change credentials
        if promptYesNo "$(msg change_auth_credentials)" "n" "n"; then
          isAuthUser=$(promptInput "$(msg prompt_user)" "admin")
          isAuthPass=$(promptInput "$(msg prompt_password)" "admin")
          if [[ $SILENT_MODE -eq 0 ]]; then
            printf ' %s\n' "$(msg store_auth "$isAuthUser" "$isAuthPass" "${dirInstall}/accs.db")"
          fi
          echo -e "{\n  \"$isAuthUser\": \"$isAuthPass\"\n}" > "$dirInstall/accs.db"
        fi
      fi
    fi
  fi

  # Read-only database configuration
  if [[ -z "$isRdb" ]]; then
    if promptYesNo "$(msg enable_rdb)" "n" "n"; then
      isRdb=1
    else
      isRdb=0
    fi
  else
    # RDB setting exists, ask if user wants to change it
    if [[ $SILENT_MODE -eq 0 ]]; then
      local current_rdb_default
      current_rdb_default="$([[ $isRdb -eq 1 ]] && echo 'y' || echo 'n')"
      if promptYesNo "$(msg enable_rdb)" "$current_rdb_default" "n"; then
        isRdb=1
      else
        isRdb=0
      fi
    fi
  fi

  if [[ $isRdb -eq 1 ]] && [[ $SILENT_MODE -eq 0 ]]; then
    echo " $(msg set_readonly)"
    printf ' %s\n' "$(msg readonly_hint "$dirInstall/$serviceName.plist")"
  fi

  # Logging configuration
  if [[ -z "$isLog" ]]; then
    if promptYesNo "$(msg enable_log)" "n" "y"; then
      isLog=1
    else
      isLog=0
    fi
  else
    # Log setting exists, ask if user wants to change it
    if [[ $SILENT_MODE -eq 0 ]]; then
      local current_log_default
      current_log_default="$([[ $isLog -eq 1 ]] && echo 'y' || echo 'n')"
      if promptYesNo "$(msg enable_log)" "$current_log_default" "y"; then
        isLog=1
      else
        isLog=0
      fi
    fi
  fi

  if [[ $isLog -eq 1 ]] && [[ $SILENT_MODE -eq 0 ]]; then
    printf ' - %s\n' "$(msg log_location "$dirInstall/$serviceName.log")"
  fi

  # LaunchAgent/LaunchDaemon selection
  if [[ $SILENT_MODE -eq 0 && $USER_PROMPTED -eq 0 ]]; then
    local answer_cu
    answer_cu=$(promptInput "$(msg prompt_launchagent)" "1")
    if [[ "$answer_cu" == "1" ]]; then
      USE_USER_LAUNCHAGENT=1
      sysPath="${HOME}/Library/LaunchAgents"
    else
      USE_USER_LAUNCHAGENT=0
      sysPath="/Library/LaunchDaemons"
    fi
    USER_PROMPTED=1
  elif [[ $SILENT_MODE -eq 1 ]]; then
    # Silent mode defaults to user LaunchAgent
    USE_USER_LAUNCHAGENT=1
    sysPath="${HOME}/Library/LaunchAgents"
  fi
}

installTorrServer() {
  if [[ $SILENT_MODE -eq 0 ]]; then
    echo " $(msg install_configure)"
  fi

  # Get target version
  local target_version
  target_version=$(getTargetVersion)
  if [[ $SILENT_MODE -eq 0 ]]; then
    echo " - $(msg target_version) $target_version"
  fi

  # Check if already installed and up to date
  if checkInstalled; then
    if ! checkInstalledVersion; then
      if promptYesNo "$(msg want_update)" "y" "y"; then
        UpdateVersion
        return
      fi
    else
      # Already installed and up to date, allow reconfiguration
      if [[ $SILENT_MODE -eq 0 ]]; then
        echo ""
        # Allow user to reconfigure settings
        if promptYesNo "$(msg want_reconfigure)" "n" "n"; then
          # Read existing config first
          if [[ -f "$dirInstall/$serviceName.plist" ]]; then
            readExistingConfig
          fi
          # Reconfigure service
          configureService
          # Update plist file
          createPlistFile
          # Reload and restart service
          cleanup
          installService
          echo ""
          echo " - $(msg config_updated)"
          echo ""
        fi
      fi
      return
    fi
  fi

  # Create directories
  if [[ ! -d "$dirInstall" ]]; then
    mkdir -p "$dirInstall"
    chmod a+rw "$dirInstall"
  fi

  # Download binary if needed
  local binName
  binName=$(getBinaryName)
  if [[ ! -f "$dirInstall/$binName" ]] || [[ ! -x "$dirInstall/$binName" ]] || [[ $(stat -f%z "$dirInstall/$binName" 2>/dev/null) -eq 0 ]]; then
    local urlBin
    if [[ -n "$specificVersion" ]]; then
      urlBin=$(buildDownloadUrl "$target_version" "$binName")
    else
      urlBin=$(buildDownloadUrl "latest" "$binName")
    fi
    downloadBinary "$urlBin" "$dirInstall/$binName" "$target_version"
  fi

  # Create plist and configure service
  configureService
  createPlistFile

  # Install service
  local service_started=0
  if installService; then
    service_started=1
  fi

  # Show completion message
  getIP
  local installed_version="$target_version"

  if [[ $SILENT_MODE -eq 0 ]]; then
    echo ""
    printf ' %s\n' "$(msg install_complete "$installed_version" "$dirInstall")"
    echo ""
    printf ' %s\n' "$(msg access_web "$serverIP" "$servicePort")"
    echo ""

    if [[ $isAuth -eq 1 && -n "$isAuthUser" ]]; then
      printf ' %s\n' "$(msg use_auth "$isAuthUser" "$isAuthPass")"
      echo ""
    fi

    if [[ $service_started -eq 0 ]]; then
      echo " $(colorize yellow "$(msg service_start_failed)")"
    fi
    echo ""
  fi

  if [[ $SILENT_MODE -eq 1 ]]; then
    printf "%s\n" "$(msg install_complete "$installed_version" "$dirInstall")"
    printf "%s\n" "$(msg access_web "$serverIP" "$servicePort")"
    if [[ $isAuth -eq 1 && -n "$isAuthUser" ]]; then
      printf "%s\n" "$(msg use_auth "$isAuthUser" "$isAuthPass")"
    fi
  fi

  return 0
}

installService() {
  # Cleanup existing services first
  cleanup

  if [[ $USE_USER_LAUNCHAGENT -eq 1 ]]; then
    # User LaunchAgent
    sysPath="${HOME}/Library/LaunchAgents"
    [[ ! -d "$sysPath" ]] && mkdir -p "$sysPath"
    cp "$dirInstall/$serviceName.plist" "$sysPath"
    chmod 0644 "$sysPath/$serviceName.plist"
    if launchctlCmd load -w "$sysPath/$serviceName.plist"; then
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf ' %s\n' "$(msg service_added "$sysPath")"
      fi
      return 0
    else
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf ' %s\n' "$(msg service_added "$sysPath")"
      fi
      return 1
    fi
  else
    # System LaunchDaemon
    sysPath="/Library/LaunchDaemons"
    [[ ! -d "$sysPath" ]] && sudo mkdir -p "$sysPath"
    sudo cp "$dirInstall/$serviceName.plist" "$sysPath"
    sudo chown root:wheel "$sysPath/$serviceName.plist"
    sudo chmod 0644 "$sysPath/$serviceName.plist"
    if sudo launchctl load -w "$sysPath/$serviceName.plist" >/dev/null 2>&1; then
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf ' %s\n' "$(msg service_added "$sysPath")"
      fi
      return 0
    else
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf ' %s\n' "$(msg service_added "$sysPath")"
      fi
      return 1
    fi
  fi
}

# Common function to update/downgrade TorrServer version
updateTorrServerVersion() {
  local target_version="$1"
  local cancel_message="$2"
  local use_latest_url="${3:-0}"

  killRunning

  local binName
  binName=$(getBinaryName)
  local urlBin
  if [[ $use_latest_url -eq 1 && -z "$specificVersion" ]]; then
    urlBin=$(buildDownloadUrl "latest" "$binName")
  else
    urlBin=$(buildDownloadUrl "$target_version" "$binName")
  fi

  downloadBinary "$urlBin" "$dirInstall/$binName" "$target_version"

  # Update plist file
  if [[ -f "$dirInstall/$serviceName.plist" ]]; then
    createPlistFile
    installService
  fi

  return 0
}

UpdateVersion() {
  local target_version
  target_version=$(getTargetVersion)
  updateTorrServerVersion "$target_version" "update_cancelled" 1
}

DowngradeVersion() {
  local target_version
  target_version=$(getVersionTag "$downgradeRelease")
  updateTorrServerVersion "$target_version" "downgrade_cancelled" 0
}

#############################################
#     CLEANUP FUNCTIONS
#############################################

cleanup() {
  killRunning
  launchctl unload "$HOME/Library/LaunchAgents/$serviceName.plist" >/dev/null 2>&1 || true
  sudo launchctl unload "/Library/LaunchDaemons/$serviceName.plist" >/dev/null 2>&1 || true
  rm -f "$HOME/Library/LaunchAgents/$serviceName.plist" 2>/dev/null || true
  sudo rm -f "/Library/LaunchDaemons/$serviceName.plist" 2>/dev/null || true
}

uninstall() {
  checkArch
  checkInstalled

  if [[ $SILENT_MODE -eq 1 ]]; then
    cleanup
    sudo rm -rf "$dirInstall"
    echo " - $(msg uninstalled)"
    return
  fi

  echo ""
  echo " $(msg install_dir_label) ${dirInstall}"
  echo ""
  echo " $(msg uninstall_warning)"
  echo ""

  if promptYesNo "$(msg confirm_delete)" "n" "n"; then
    cleanup
    sudo rm -rf "$dirInstall"
    echo " - $(msg uninstalled)"
    echo ""
  else
    echo ""
  fi
}

#############################################
#     RECONFIGURATION
#############################################

reconfigureTorrServer() {
  # Check if TorrServer is installed
  if ! checkInstalled; then
    echo " - $(msg not_found)"
    echo " - $(msg install_first_required "$scriptname")"
    exit 1
  fi

  if [[ $SILENT_MODE -eq 0 ]]; then
    echo ""
  fi

  # Read existing config first
  if [[ -f "$dirInstall/$serviceName.plist" ]]; then
    readExistingConfig
  fi

  # Reconfigure service
  configureService

  # Update plist file
  createPlistFile

  # Reload and restart service
  cleanup
  installService

  if [[ $SILENT_MODE -eq 0 ]]; then
    echo ""
    echo " - $(msg config_updated)"
    echo ""
  else
    echo " - $(msg config_updated)"
  fi
}

#############################################
#     HELP & MAIN
#############################################

helpUsage() {
  cat << EOF
$scriptname - TorrServer Installation Script

Usage: $scriptname [COMMAND] [OPTIONS]

Commands:
  -i, --install [VERSION]           Install latest or specific version
    install [VERSION]
  -u, --update                      Update to latest version
    update
  -c, --check                       Check for updates (version info only)
    check
  -d, --down VERSION                Downgrade to specific version
    down VERSION
  -r, --remove                      Uninstall TorrServer
    remove
  --reconfigure                     Reconfigure TorrServer settings
    reconfigure
  -h, --help                        Show this help message
    help

Options:
  --silent                          Non-interactive mode with defaults

Examples:
  # Install latest version interactively
  $scriptname --install

  # Install specific version silently
  $scriptname --install 135 --silent

  # Update with silent mode
  $scriptname --update --silent

  # Check for updates
  $scriptname --check

  # Uninstall silently
  $scriptname --remove --silent

  # Reconfigure TorrServer settings interactively
  $scriptname --reconfigure

Default Settings (silent mode):
  - Port: ${DEFAULT_PORT}
  - LaunchAgent: current user (not system-wide)
  - Auth: disabled
  - Read-only mode: disabled
  - Logging: disabled

EOF
}

parseArguments() {
  parsedCommand=""

  while [[ $# -gt 0 ]]; do
    case $1 in
      -i|--install|install)
        parsedCommand="install"
        shift
        # Check for version number
        if [[ $# -gt 0 ]]; then
          local next_arg="$1"
          if [[ "$next_arg" =~ ^[0-9]+$ ]]; then
            specificVersion="$next_arg"
            shift
          fi
        fi
        ;;
      -u|--update|update)
        parsedCommand="update"
        shift
        ;;
      -c|--check|check)
        parsedCommand="check"
        shift
        ;;
      -d|--down|down)
        parsedCommand="downgrade"
        shift
        if [[ $# -gt 0 ]]; then
          local next_arg="$1"
          if [[ "$next_arg" =~ ^[0-9]+$ ]]; then
            downgradeRelease="$next_arg"
            shift
          else
            echo " $(msg error_version_required)"
            echo " $(msg error_version_example "$scriptname")"
            exit 1
          fi
        else
          echo " $(msg error_version_required)"
          echo " $(msg error_version_example "$scriptname")"
          exit 1
        fi
        ;;
      -r|--remove|remove)
        parsedCommand="remove"
        shift
        ;;
      --reconfigure|reconfigure)
        parsedCommand="reconfigure"
        shift
        ;;
      -h|--help|help)
        getLang  # Set language before showing help
        helpUsage
        exit 0
        ;;
      --silent)
        SILENT_MODE=1
        shift
        ;;
      *)
        echo " $(msg error_unknown_option "$1")"
        helpUsage
        exit 1
        ;;
    esac
  done
}

#############################################
#     MAIN EXECUTION
#############################################

main() {
  getLang

  parseArguments "$@"

  local command="$parsedCommand"

  case "$command" in
    install)
      if [[ $SILENT_MODE -eq 0 && -n "$specificVersion" ]]; then
        echo " - $(msg installing_specific_version "$specificVersion")"
      fi
      initialCheck

      if [[ $SILENT_MODE -eq 1 ]]; then
        servicePort="$DEFAULT_PORT"
        isAuth=0
        isRdb=0
        isLog=0
        USE_USER_LAUNCHAGENT=1
        USER_PROMPTED=1
      fi

      if ! checkInstalled; then
        installTorrServer
      else
        createPlistFile
        installService
        if [[ $SILENT_MODE -eq 0 ]]; then
          echo " - $(msg config_updated)"
        fi
      fi
      exit 0
      ;;
    update)
      initialCheck
      if checkInstalled; then
        if ! checkInstalledVersion; then
          UpdateVersion
        fi
      fi
      exit 0
      ;;
    check)
      initialCheck
      if checkInstalled; then
        checkInstalledVersion
      fi
      exit 0
      ;;
    downgrade)
      initialCheck
      if checkInstalled; then
        DowngradeVersion
      fi
      exit 0
      ;;
    remove)
      uninstall
      exit 0
      ;;
    reconfigure)
      initialCheck
      reconfigureTorrServer
      exit 0
      ;;
  esac

  # Interactive mode if no command provided and not silent
  if [[ $SILENT_MODE -eq 0 ]]; then
    echo ""
    echo " $(msg lang_choice)"
    echo " [$(colorize green 1)] $(msg lang_english)"
    echo " [$(colorize yellow 2)] $(msg lang_russian)"
    local answer_lang
    answer_lang=$(promptInput "$(msg your_lang)" "1")
    if [[ "$answer_lang" == "2" ]]; then
      lang="ru"
    fi

    echo ""
    echo "============================================================="
    echo " $(msg script_title)"
    echo "============================================================="
    echo ""

    local user_choice
    user_choice=$(promptYesNoDelete "$(msg want_install)" "n" "y")

    if [[ "$user_choice" == "delete" ]]; then
      initialCheck
      uninstall
    elif [[ "$user_choice" == "yes" ]]; then
      initialCheck
      USER_PROMPTED=0
      installTorrServer
    fi
  fi

  echo " $(msg have_fun)"
  echo ""
}

# Run main function
main "$@"
