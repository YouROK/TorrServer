#!/usr/bin/env bash

set -euo pipefail

#############################################
#           GLOBAL VARIABLES
#############################################

# Installation settings
DEFAULT_USERNAME="torrserver"
DEFAULT_INSTALL_DIR="/opt/torrserver"
DEFAULT_SERVICE_NAME="torrserver"
DEFAULT_PORT="8090"

# Runtime variables
username="${DEFAULT_USERNAME}"
dirInstall="${DEFAULT_INSTALL_DIR}"
serviceName="${DEFAULT_SERVICE_NAME}"
scriptname=$(basename "$(test -L "$0" && readlink "$0" || echo "$0")")

# Flags
SILENT_MODE=0
USE_ROOT_USER=0
ROOT_PROMPTED=0

# Command-line state
parsedCommand=""
specificVersion=""
downgradeRelease=""
changeUserName=""

# Service configuration
servicePort=""
isAuth=""
isRdb=""
isLog=""
isBbr=""
isAuthUser=""
isAuthPass=""

# Constants
readonly REPO_URL="https://github.com/YouROK/TorrServer"
readonly REPO_API_URL="https://api.github.com/repos/YouROK/TorrServer"
readonly VERSION_PREFIX="MatriX"
readonly BINARY_NAME_PREFIX="TorrServer-linux"
readonly MIN_GLIBC_VERSION="2.32"
readonly MIN_VERSION_REQUIRING_GLIBC=136
readonly SYSCTL_BBR_FILE="/etc/sysctl.d/90-torrserver.conf"

# Color support
declare -A colors=([black]=0 [red]=1 [green]=2 [yellow]=3 [blue]=4 [magenta]=5 [cyan]=6 [white]=7)
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
declare -A MSG_EN=(
  # General
  [lang_choice]="Choose Language:"
  [lang_english]="English"
  [lang_russian]="Русский"
  [your_lang]="Your language (Ваш язык): "
  [have_fun]="Have Fun!"

  # Script info
  [script_title]="TorrServer install and configuration script for Linux"

  # Checks
  [need_root]="Script must run as root or user with sudo privileges. Example: sudo $scriptname"
  [unsupported_arch]="Unsupported Arch. Can't continue."
  [unsupported_os]="It looks like you are running this installer on a system other than Debian, Ubuntu, Fedora, CentOS, Amazon Linux, Oracle Linux or Arch Linux."

  # User management
  [user_exists]="User %s exists!"
  [user_added]="User %s has been added to system!"
  [user_add_failed]="Failed to add %s user!"
  [user_removed]="User %s has been removed from system!"
  [user_remove_failed]="Failed to remove %s user!"
  [user_not_found]="%s - no such user!"

  # Version
  [downloading]="Downloading TorrServer"
  [target_version]="Target version:"
  [installed_version]="installed:"
  [target_label]="target:"
  [version_not_found]="ERROR: Version %s not found in releases"
  [check_versions]="Please check available versions at: $REPO_URL/releases"
  [already_installed]="You already have TorrServer %s installed"
  [have_latest]="You have latest TorrServer %s"
  [update_found]="TorrServer update found!"
  [will_install]="Will install TorrServer version %s"

  # Installation
  [installing_packages]="Installing missing packages…"
  [install_configure]="Install and configure TorrServer…"
  [starting_service]="Starting TorrServer…"
  [install_complete]="TorrServer %s installed to %s"
  [access_web]="You can now open your browser at http://%s:%s to access TorrServer web GUI."
  [use_auth]="Use user \"%s\" with password \"%s\" for authentication"

  # Prompts
  [want_update]="Want to update TorrServer? (Yes/No) "
  [want_install]="Want to install or configure TorrServer? (Yes|No) Type Delete to uninstall. "
  [want_reconfigure]="Do you want to reconfigure TorrServer settings? (Yes/No) "
  [change_port]="Change TorrServer web-port? (Yes/No) "
  [enter_port]="Enter port number: "
  [enable_auth]="Enable server authorization? (Yes/No) "
  [prompt_user]="User: "
  [prompt_password]="Password: "
  [change_auth_credentials]="Change authentication username and password? (Yes/No) "
  [enable_rdb]="Start TorrServer in public read-only mode? (Yes/No) "
  [enable_log]="Enable TorrServer log output to file? (Yes/No) "
  [enable_bbr]="Enable BBR (recommended for better download speed)? (Yes/No) "
  [confirm_delete]="Are you sure you want to delete TorrServer? (Yes/No) "
  [prompt_run_as_root]="Run service as root user? (Yes/No) "

  # Uninstall
  [install_dir_label]="TorrServer install dir -"
  [uninstall_warning]="This action will delete TorrServer including all it's torrents, settings and files on path above!"
  [uninstalled]="TorrServer uninstalled!"

  # Status
  [found_in]="TorrServer found in"
  [not_found]="TorrServer not found. It's not installed or have zero size."
  [no_version_info]="No version information available. Can be server issue."
  [config_updated]="Configuration updated successfully"
  [store_auth]="Store %s:%s to %s"
  [use_existing_auth]="Use existing auth from %s - %s"
  [set_readonly]="Set database to read-only mode…"
  [readonly_hint]="To change remove --rdb option from %s or rerun install script without parameters"
  [log_location]="TorrServer log stored at %s"
  [bbr_enabled]="BBR TCP congestion control enabled"
  [bbr_already_configured]="BBR is already configured"
  [bbr_configured_not_available]="BBR is configured but not available in this kernel"
  [bbr_config_failed]="Warning: Failed to configure BBR"
  [bbr_not_available]="BBR is not available in this kernel"
  [bbr_requires_kernel]="BBR requires Linux kernel 4.9+ with tcp_bbr module"
  [bbr_write_failed]="Failed to write to %s"
  [bbr_current_values]="Current: qdisc=%s, congestion_control=%s"
  [bbr_settings_will_apply]="BBR settings are in %s and will apply after reboot"
  [bbr_settings_not_added]="BBR settings were not added to %s"
  [bbr_activate_failed]="Warning: Could not activate BBR - module not available"
  [bbr_no_optimization]="Service will start without BBR optimization"
  [bbr_activate_failed_cc]="Warning: Could not activate BBR (currently: %s)"
  [systemctl_missing]="systemctl is not available. Skipping service management commands."
  [systemctl_failed]="Warning: systemctl %s failed"
  [service_start_failed]="Warning: TorrServer service failed to start. Check systemctl status for details."
  [user_change_success]="Service user changed to %s"
  [user_change_permissions]="Updated ownership of %s to %s:%s"
  [user_change_missing]="TorrServer installation not found in %s. Run install first."
  [user_change_invalid]="Only %s or %s are allowed for --change-user"

  # glibc
  [glibc_error]="ERROR: TorrServer version %s requires glibc >= $MIN_GLIBC_VERSION"
  [glibc_current]="Your system has glibc %s"
  [glibc_upgrade]="Please install a version < $MIN_VERSION_REQUIRING_GLIBC or upgrade your system"
  [glibc_detected]="Detected glibc version: %s"
  [glibc_ok]="OK: glibc version meets requirements for TorrServer %s"
  [glibc_no_requirement]="TorrServer version %s: no special glibc requirements"
  [glibc_warning]="Warning: Could not detect glibc version"
  [glibc_may_fail]="TorrServer version %s requires glibc >= $MIN_GLIBC_VERSION"
  [glibc_install_may_fail]="Installation may fail if your system doesn't meet this requirement"
  [update_cancelled]="Update cancelled due to glibc incompatibility"
  [downgrade_cancelled]="Downgrade cancelled due to glibc incompatibility"

  # OS version errors
  [os_not_supported]="Your %s version is not supported."
  [os_script_supports]="Script supports only %s %s"

  # User mode
  [running_as_root]="Service will run as root user"
  [running_as_user]="Service will run as %s user"

  # Error messages
  [error_username_required]="Error: Username required for --change-user"
  [error_version_required]="Error: Version number required for downgrade"
  [error_version_example]="Example: %s -d 101"
  [error_unknown_option]="Unknown option: %s"
  [installing_specific_version]="Installing specific version: %s"
  [service_reconfigured_user]="Service reconfigured for user: %s"
  [install_first_required]="Please install TorrServer first using: %s --install"
)

declare -A MSG_RU=(
  # General
  [lang_choice]="Choose Language:"
  [lang_english]="English"
  [lang_russian]="Русский"
  [your_lang]="Your language (Ваш язык): "
  [have_fun]="Have Fun!"

  # Script info
  [script_title]="Скрипт установки, удаления и настройки TorrServer для Linux"

  # Checks
  [need_root]="Вам нужно запустить скрипт от root или пользователя с правами sudo. Пример: sudo $scriptname"
  [unsupported_arch]="Не поддерживаемая архитектура. Продолжение невозможно."
  [unsupported_os]="Похоже, что вы запускаете этот установщик в системе отличной от Debian, Ubuntu, Fedora, CentOS, Amazon Linux, Oracle Linux или Arch Linux."

  # User management
  [user_exists]="пользователь %s найден!"
  [user_added]="пользователь %s добавлен!"
  [user_add_failed]="не удалось добавить пользователя %s!"
  [user_removed]="Пользователь %s удален!"
  [user_remove_failed]="не удалось удалить пользователя %s!"
  [user_not_found]="пользователь %s не найден!"

  # Version
  [downloading]="Загружаем TorrServer"
  [target_version]="Устанавливаемая версия:"
  [installed_version]="установлен:"
  [target_label]="устанавливаемая:"
  [version_not_found]="ОШИБКА: Версия %s не найдена в релизах"
  [check_versions]="Проверьте доступные версии по адресу: $REPO_URL/releases"
  [already_installed]="TorrServer %s уже установлен"
  [have_latest]="Установлен TorrServer последней версии %s"
  [update_found]="Доступно обновление сервера"
  [will_install]="Будет установлена версия TorrServer %s"

  # Installation
  [installing_packages]="Устанавливаем недостающие пакеты…"
  [install_configure]="Устанавливаем и настраиваем TorrServer…"
  [starting_service]="Запускаем службу TorrServer…"
  [install_complete]="TorrServer %s установлен в директории %s"
  [access_web]="Теперь вы можете открыть браузер по адресу http://%s:%s для доступа к вебу TorrServer"
  [use_auth]="Для авторизации используйте пользователя «%s» с паролем «%s»"

  # Prompts
  [want_update]="Хотите обновить TorrServer? (Yes/No) "
  [want_install]="Хотите установить, обновить или настроить TorrServer? (Yes|No) Для удаления введите «Delete» "
  [want_reconfigure]="Хотите перенастроить параметры TorrServer? (Yes/No) "
  [change_port]="Хотите изменить порт для TorrServer? (Yes/No) "
  [enter_port]="Введите номер порта: "
  [enable_auth]="Включить авторизацию на сервере? (Yes/No) "
  [prompt_user]="Пользователь: "
  [prompt_password]="Пароль: "
  [change_auth_credentials]="Изменить имя пользователя и пароль для авторизации? (Yes/No) "
  [enable_rdb]="Запускать TorrServer в публичном режиме без возможности изменения настроек через веб сервера? (Yes/No) "
  [enable_log]="Включить запись журнала работы TorrServer в файл? (Yes/No) "
  [enable_bbr]="Включить BBR (рекомендуется для лучшей скорости загрузки)? (Yes/No) "
  [confirm_delete]="Вы уверены что хотите удалить программу? (Yes/No) "
  [prompt_run_as_root]="Запускать службу от пользователя root? (Yes/No) "

  # Uninstall
  [install_dir_label]="Директория c TorrServer -"
  [uninstall_warning]="Это действие удалит все данные TorrServer включая базу данных торрентов и настройки по указанному выше пути!"
  [uninstalled]="TorrServer удален из системы!"

  # Status
  [found_in]="TorrServer найден в директории"
  [not_found]="TorrServer не найден, возможно он не установлен или размер бинарника равен 0."
  [no_version_info]="Информация о версии недоступна. Возможно сервер не доступен."
  [config_updated]="Конфигурация успешно обновлена"
  [store_auth]="Сохраняем %s:%s в %s"
  [use_existing_auth]="Используйте реквизиты из %s для авторизации - %s"
  [set_readonly]="База данных устанавливается в режим «только для чтения»…"
  [readonly_hint]="Для изменения отредактируйте %s, убрав опцию --rdb или запустите интерактивную установку без параметров повторно"
  [log_location]="лог TorrServer располагается по пути %s"
  [bbr_enabled]="Включено управление перегрузкой TCP BBR"
  [bbr_already_configured]="BBR уже настроен"
  [bbr_configured_not_available]="BBR настроен, но недоступен в этом ядре"
  [bbr_config_failed]="Предупреждение: Не удалось настроить BBR"
  [bbr_not_available]="BBR недоступен в этом ядре"
  [bbr_requires_kernel]="BBR требует Linux kernel 4.9+ с модулем tcp_bbr"
  [bbr_write_failed]="Не удалось записать в %s"
  [bbr_current_values]="Текущие значения: qdisc=%s, congestion_control=%s"
  [bbr_settings_will_apply]="Настройки BBR находятся в %s и вступят в силу после перезагрузки"
  [bbr_settings_not_added]="Настройки BBR не были добавлены в %s"
  [bbr_activate_failed]="Предупреждение: Не удалось активировать BBR - модуль недоступен"
  [bbr_no_optimization]="Служба запустится без оптимизации BBR"
  [bbr_activate_failed_cc]="Предупреждение: Не удалось активировать BBR (текущее: %s)"
  [systemctl_missing]="systemctl недоступен. Пропускаем команды управления службой."
  [systemctl_failed]="Предупреждение: команда systemctl %s завершилась ошибкой"
  [service_start_failed]="Предупреждение: служба TorrServer не запустилась. Проверьте systemctl status для деталей."
  [user_change_success]="Сервис TorrServer теперь запускается от пользователя %s"
  [user_change_permissions]="Обновлены права на %s: %s:%s"
  [user_change_missing]="Установка TorrServer не найдена в %s. Сначала выполните установку."
  [user_change_invalid]="Параметр --change-user принимает только %s или %s"

  # glibc
  [glibc_error]="ОШИБКА: TorrServer версии %s требует glibc >= $MIN_GLIBC_VERSION"
  [glibc_current]="В вашей системе установлена glibc %s"
  [glibc_upgrade]="Пожалуйста, установите версию < $MIN_VERSION_REQUIRING_GLIBC или обновите систему"
  [glibc_detected]="Обнаружена версия glibc: %s"
  [glibc_ok]="OK: версия glibc соответствует требованиям для TorrServer %s"
  [glibc_no_requirement]="TorrServer версии %s: нет особых требований к glibc"
  [glibc_warning]="Предупреждение: Не удалось определить версию glibc"
  [glibc_may_fail]="TorrServer версии %s требует glibc >= $MIN_GLIBC_VERSION"
  [glibc_install_may_fail]="Установка может завершиться неудачей, если система не соответствует требованиям"
  [update_cancelled]="Обновление отменено из-за несовместимости glibc"
  [downgrade_cancelled]="Понижение версии отменено из-за несовместимости glibc"

  # OS version errors
  [os_not_supported]="Ваша версия %s не поддерживается."
  [os_script_supports]="Скрипт поддерживает только %s %s"

  # User mode
  [running_as_root]="Служба будет запущена от пользователя root"
  [running_as_user]="Служба будет запущена от пользователя %s"

  # Error messages
  [error_username_required]="Ошибка: Требуется имя пользователя для --change-user"
  [error_version_required]="Ошибка: Требуется номер версии для понижения версии"
  [error_version_example]="Пример: %s -d 101"
  [error_unknown_option]="Неизвестная опция: %s"
  [installing_specific_version]="Установка конкретной версии: %s"
  [service_reconfigured_user]="Служба перенастроена для пользователя: %s"
  [install_first_required]="Пожалуйста, сначала установите TorrServer используя: %s --install"
)

# Translation function
msg() {
  local key="$1"
  shift
  local message=""

  if [[ $lang == "ru" ]]; then
    message="${MSG_RU[$key]:-$key}"
  else
    message="${MSG_EN[$key]:-$key}"
  fi

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
    printf "%s%s%s" "$(tput setaf "${colors[$1]:-7}")" "$2" "$(tput op)"
  else
    printf "%s" "$2"
  fi
}

isRoot() {
  [[ $EUID -eq 0 ]]
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

  if command -v dig >/dev/null 2>&1; then
    ip=$(dig +short myip.opendns.com @resolver1.opendns.com 2>/dev/null || echo "")
    if [[ -z "$ip" ]]; then
      ip="localhost"
    fi
  elif command -v host >/dev/null 2>&1; then
    local host_output=""
    host_output=$(host myip.opendns.com resolver1.opendns.com 2>/dev/null || true)
    ip=$(printf "%s\n" "$host_output" | tail -n1 | awk '{print $NF}')
    if [[ -z "$ip" ]]; then
      ip="localhost"
    fi
  fi

  serverIP="$ip"
}

promptYesNo() {
  local prompt="$1"
  local default="${2:-n}"

  if [[ $SILENT_MODE -eq 1 ]]; then
    if [[ "$default" == "y" ]]; then
      return 0
    else
      return 1
    fi
  fi

  local answer
  IFS= read -r -p " $prompt " answer </dev/tty

  if [[ "$answer" =~ ^[YyДд] ]]; then
    return 0
  else
    return 1
  fi
}

promptYesNoDelete() {
  local prompt="$1"
  local default="${2:-n}"

  if [[ $SILENT_MODE -eq 1 ]]; then
    if [[ "$default" == "y" ]]; then
      echo "yes"
    else
      echo "no"
    fi
    return
  fi

  local answer
  IFS= read -r -p " $prompt " answer </dev/tty
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

systemctlCmd() {
  local quiet=0
  if [[ ${1:-} == "--quiet" ]]; then
    quiet=1
    shift
  fi

  if ! command -v systemctl >/dev/null 2>&1; then
    if [[ $quiet -eq 0 && $SILENT_MODE -eq 0 ]]; then
      echo " - $(msg systemctl_missing)"
    fi
    return 1
  fi

  local rc
  systemctl "$@" >/dev/null 2>&1
  rc=$?
  if [[ $rc -ne 0 ]]; then
    if [[ $quiet -eq 0 && $SILENT_MODE -eq 0 ]]; then
      printf " - %s\n" "$(msg systemctl_failed "$*")"
    fi
    return $rc
  fi

  return 0
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
}

#############################################
#     GLIBC COMPATIBILITY
#############################################

getGlibcVersion() {
  local glibc_version

  # Try ldd --version (most reliable)
  if command -v ldd >/dev/null 2>&1; then
    glibc_version=$(ldd --version 2>/dev/null | head -n1 | grep -oE '[0-9]+\.[0-9]+' | head -n1)
    if [[ -n "$glibc_version" ]]; then
      echo "$glibc_version"
      return 0
    fi
  fi

  # Try getconf GNU_LIBC_VERSION
  if command -v getconf >/dev/null 2>&1; then
    glibc_version=$(getconf GNU_LIBC_VERSION 2>/dev/null | grep -oE '[0-9]+\.[0-9]+')
    if [[ -n "$glibc_version" ]]; then
      echo "$glibc_version"
      return 0
    fi
  fi

  # Try rpm package manager
  if command -v rpm >/dev/null 2>&1; then
    glibc_version=$(rpm -q glibc 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -n1)
    if [[ -n "$glibc_version" ]]; then
      echo "$glibc_version"
      return 0
    fi
  fi

  # Try dpkg package manager
  if command -v dpkg >/dev/null 2>&1; then
    glibc_version=$(dpkg -l libc6 2>/dev/null | awk '/^ii/ {print $3}' | grep -oE '[0-9]+\.[0-9]+' | head -n1)
    if [[ -n "$glibc_version" ]]; then
      echo "$glibc_version"
      return 0
    fi
  fi

  return 1
}

compareVersions() {
  local ver1="$1"
  local ver2="$2"
  local sorted_first
  sorted_first=$(printf '%s\n' "$ver1" "$ver2" | sort -V | head -n1)
  [[ "$sorted_first" == "$ver2" ]]
}

checkGlibcCompatibility() {
  local target_version="$1"
  local version_number

  # Extract numeric version
  if [[ "$target_version" =~ ${VERSION_PREFIX}\.([0-9]+) ]]; then
    version_number="${BASH_REMATCH[1]}"
  elif [[ "$target_version" =~ ^[0-9]+$ ]]; then
    version_number="$target_version"
  else
    if [[ $SILENT_MODE -eq 0 ]]; then
      echo " - $(msg glibc_warning)"
    fi
    return 0
  fi

  # Check if version requires glibc 2.32+
  if [[ $version_number -ge $MIN_VERSION_REQUIRING_GLIBC ]]; then
  local current_glibc
  current_glibc=$(getGlibcVersion)

    if [[ -z "$current_glibc" ]]; then
      if [[ $SILENT_MODE -eq 0 ]]; then
        echo " - $(msg glibc_warning)"
        echo " - $(msg glibc_may_fail "$target_version")"
        echo " - $(msg glibc_install_may_fail)"
      fi
      return 0
    fi

    if [[ $SILENT_MODE -eq 0 ]]; then
      echo " - $(msg glibc_detected "$current_glibc")"
    fi

    if ! compareVersions "$current_glibc" "$MIN_GLIBC_VERSION"; then
      echo " - $(colorize red "$(msg glibc_error "$target_version")")"
      echo " - $(msg glibc_current "$current_glibc")"
      echo " - $(msg glibc_upgrade)"
      return 1
    fi

    if [[ $SILENT_MODE -eq 0 ]]; then
      echo " - $(colorize green "$(msg glibc_ok "$target_version")")"
    fi
  else
    if [[ $SILENT_MODE -eq 0 ]]; then
      echo " - $(msg glibc_no_requirement "$target_version")"
    fi
  fi

  return 0
}

#############################################
#     USER MANAGEMENT
#############################################

addUser() {
  if ! isRoot; then
    return 1
  fi

  if [[ $username == "root" ]]; then
    return 0
  fi

  if id "$username" >/dev/null 2>&1; then
    if [[ $SILENT_MODE -eq 0 ]]; then
      printf " - %s\n" "$(msg user_exists "$username")"
    fi
    return 0
  else
    if useradd --home-dir "$dirInstall" --create-home --shell /bin/false -c "TorrServer" "$username" 2>/dev/null; then
      chmod 755 "$dirInstall"
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf " - %s\n" "$(msg user_added "$username")"
      fi
      return 0
    else
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf " - %s\n" "$(msg user_add_failed "$username")"
      fi
      return 1
    fi
  fi
}

delUser() {
  if ! isRoot; then
    return 1
  fi

  if [[ $username == "root" ]]; then
    return 0
  fi

  if id "$username" >/dev/null 2>&1; then
    if userdel --remove "$username" 2>/dev/null; then
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf " - %s\n" "$(msg user_removed "$username")"
      fi
      return 0
    else
      if [[ $SILENT_MODE -eq 0 ]]; then
        printf " - %s\n" "$(msg user_remove_failed "$username")"
      fi
      return 1
    fi
  else
    if [[ $SILENT_MODE -eq 0 ]]; then
      printf " - %s\n" "$(msg user_not_found "$username")"
    fi
    return 1
  fi
}

#############################################
#     OS DETECTION & PACKAGES
#############################################

installPackages() {
  local pkg_type="$1"
  shift
  local packages=("$@")

  case "$pkg_type" in
    deb)
      local missing=()
      for pkg in "${packages[@]}"; do
        if ! dpkg -s "$pkg" >/dev/null 2>&1; then
          missing+=("$pkg")
        fi
      done
      if [[ ${#missing[@]} -gt 0 ]]; then
        if [[ $SILENT_MODE -eq 0 ]]; then
          echo " $(msg installing_packages)"
        fi
        apt update >/dev/null 2>&1
        apt -y install "${missing[@]}"
      fi
      ;;
    rpm)
      local pkg_manager="$1"
      shift
      packages=("$@")
      local needs_update=0
      for pkg in "${packages[@]}"; do
        if [[ -z "$(rpm -qa "$pkg" 2>/dev/null)" ]]; then
          needs_update=1
          break
        fi
      done
      if [[ $needs_update -eq 1 ]]; then
        if [[ "$pkg_manager" == "dnf" ]]; then
          dnf makecache -q >/dev/null 2>&1 || true
        elif [[ "$pkg_manager" == "yum" ]]; then
          yum makecache fast -q >/dev/null 2>&1 || true
        fi
        for pkg in "${packages[@]}"; do
          if [[ -z "$(rpm -qa "$pkg" 2>/dev/null)" ]]; then
            $pkg_manager -y install "$pkg"
          fi
        done
      fi
      ;;
    arch)
      local missing=()
      for pkg in "${packages[@]}"; do
        if ! pacman -Q "$pkg" >/dev/null 2>&1; then
          missing+=("$pkg")
        fi
      done
      if [[ ${#missing[@]} -gt 0 ]]; then
        pacman -Sy --noconfirm >/dev/null 2>&1
        pacman -S --noconfirm "${missing[@]}"
      fi
      ;;
  esac
}

getRpmPackageManager() {
  local version_id="$1"

  if [[ "$version_id" =~ ^[0-9]+$ ]] && [[ $version_id -ge 8 ]] && command -v dnf >/dev/null 2>&1; then
    echo "dnf"
  elif command -v dnf >/dev/null 2>&1; then
    echo "dnf"
  else
    echo "yum"
  fi
}

validateOSVersion() {
  local os_name="$1"
  local supported_versions="$2"
  local version_id="$3"

  local major_version
  major_version=$(echo "$version_id" | cut -d '.' -f1)

  if [[ ! $major_version =~ ^($supported_versions)$ ]]; then
    echo ""
    echo " $(msg os_not_supported "$os_name")"
    echo ""
    echo " $(msg os_script_supports "$os_name" "$supported_versions")"
    echo ""
    exit 1
  fi
}

checkOS() {
  if [[ -e /etc/debian_version ]]; then
    # shellcheck source=/dev/null
    source /etc/os-release

    if [[ $ID == "debian" || $ID == "raspbian" ]]; then
      local current_version_id
      current_version_id="${VERSION_ID:-}"
      if [[ -n "$current_version_id" && $current_version_id -lt 6 ]]; then
        validateOSVersion "Debian" ">=6" "$current_version_id"
      fi
    elif [[ $ID == "ubuntu" ]]; then
      local current_version_id
      current_version_id="${VERSION_ID:-}"
      local major
      major=$(echo "$current_version_id" | cut -d '.' -f1)
      if [[ -n "$current_version_id" && $major -lt 10 ]]; then
        validateOSVersion "Ubuntu" ">=10" "$current_version_id"
      fi
    fi

    installPackages deb curl iputils-ping dnsutils

  elif [[ -e /etc/system-release ]]; then
    # shellcheck source=/dev/null
    source /etc/os-release
    local pkg_manager

    case "$ID" in
      fedora)
        pkg_manager=$(getRpmPackageManager "${VERSION_ID%%.*}")
        installPackages rpm "$pkg_manager" curl iputils bind-utils
        ;;
      centos|redhat)
        validateOSVersion "CentOS/RedHat" "7|8|9|10" "$VERSION_ID"
        pkg_manager=$(getRpmPackageManager "${VERSION_ID%%.*}")
        installPackages rpm "$pkg_manager" curl iputils bind-utils
        ;;
      rocky)
        validateOSVersion "RockyLinux" "8|9|10" "$VERSION_ID"
        pkg_manager=$(getRpmPackageManager "${VERSION_ID%%.*}")
        installPackages rpm "$pkg_manager" curl iputils bind-utils
        ;;
      almalinux)
        validateOSVersion "AlmaLinux" "8|9|10" "$VERSION_ID"
        pkg_manager=$(getRpmPackageManager "${VERSION_ID%%.*}")
        installPackages rpm "$pkg_manager" curl iputils bind-utils
        ;;
      ol)
        validateOSVersion "Oracle Linux" "8|9|10" "$VERSION_ID"
        pkg_manager=$(getRpmPackageManager "${VERSION_ID%%.*}")
        installPackages rpm "$pkg_manager" curl iputils bind-utils
        ;;
      amzn)
        if [[ $VERSION_ID != "2" ]]; then
          validateOSVersion "Amazon Linux" "2" "$VERSION_ID"
        fi
        installPackages rpm yum curl iputils bind-utils
        ;;
    esac

  elif [[ -e /etc/arch-release ]]; then
    installPackages arch curl iputils bind-tools

  else
    echo " $(msg unsupported_os)"
    exit 1
  fi
}

checkArch() {
  case $(uname -m) in
    i386|i686) architecture="386" ;;
    x86_64) architecture="amd64" ;;
    aarch64) architecture="arm64" ;;
    armv7|armv7l) architecture="arm7" ;;
    armv6|armv6l) architecture="arm5" ;;
    *)
      echo " $(msg unsupported_arch)"
      exit 1
      ;;
  esac
}

initialCheck() {
  if ! isRoot; then
    echo " $(msg need_root)"
    exit 1
  fi

  checkOS
  checkArch
}

#############################################
#     INSTALLATION FUNCTIONS
#############################################

checkInstalled() {
  # Set username based on USE_ROOT_USER flag
  if [[ $USE_ROOT_USER -eq 1 ]]; then
    username="root"
  else
    username="${DEFAULT_USERNAME}"
    if ! addUser; then
      username="root"
    fi
  fi

  local binName
  binName=$(getBinaryName)
  if [[ -f "$dirInstall/$binName" ]] && [[ $(stat -c%s "$dirInstall/$binName" 2>/dev/null) -ne 0 ]]; then
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

createServiceFile() {
  cat << EOF > "$dirInstall/$serviceName.service"
[Unit]
Description=TorrServer - stream torrent to http
Wants=network-online.target
After=network.target

[Service]
User=$username
Group=$username
Type=simple
NonBlocking=true
EnvironmentFile=$dirInstall/$serviceName.config
ExecStart=${dirInstall}/$(getBinaryName) \$DAEMON_OPTIONS
ExecReload=/bin/kill -HUP \${MAINPID}
ExecStop=/bin/kill -INT \${MAINPID}
TimeoutSec=30
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
}

# Check if BBR is available in the kernel
isBBRAvailable() {
  sysctl net.ipv4.tcp_available_congestion_control 2>/dev/null | grep -qw "bbr"
}

# Load BBR kernel module
loadBBRModule() {
  local max_attempts=3
  local attempt=0

  while [[ $attempt -lt $max_attempts ]]; do
    if modprobe tcp_bbr >/dev/null 2>&1; then
      sleep 0.3
      if isBBRAvailable; then
        return 0
      fi
    fi
    ((attempt++))
    sleep 0.2
  done

  return 1
}

# Ensure BBR module loads at boot
ensureBBRModuleAtBoot() {
  local modules_dir="/etc/modules-load.d"
  local modules_file="$modules_dir/bbr.conf"

  [[ -d "$modules_dir" ]] || mkdir -p "$modules_dir" 2>/dev/null || return 1
  if [[ ! -f "$modules_file" ]] || ! grep -q "^tcp_bbr$" "$modules_file" 2>/dev/null; then
    echo "tcp_bbr" >> "$modules_file" 2>/dev/null || return 1
  fi
  return 0
}

# Check if BBR is configured in sysctl.d file
isBBRConfiguredInFile() {
  [[ -f "$SYSCTL_BBR_FILE" ]] && \
    grep -q "^net.core.default_qdisc=fq" "$SYSCTL_BBR_FILE" 2>/dev/null && \
    grep -q "^net.ipv4.tcp_congestion_control=bbr" "$SYSCTL_BBR_FILE" 2>/dev/null
}

# Add BBR settings to sysctl.d file
addBBRToSysctl() {
  if ! grep -q "^net.core.default_qdisc=fq" "$SYSCTL_BBR_FILE" 2>/dev/null; then
    echo "net.core.default_qdisc=fq" >> "$SYSCTL_BBR_FILE" 2>/dev/null || return 1
  fi
  if ! grep -q "^net.ipv4.tcp_congestion_control=bbr" "$SYSCTL_BBR_FILE" 2>/dev/null; then
    echo "net.ipv4.tcp_congestion_control=bbr" >> "$SYSCTL_BBR_FILE" 2>/dev/null || return 1
  fi
  return 0
}

# Check if BBR is currently active in the kernel
isBBRActive() {
  local current_cc
  current_cc=$(sysctl -n net.ipv4.tcp_congestion_control 2>/dev/null || echo "")
  [[ "$current_cc" == "bbr" ]]
}

# Apply BBR settings to kernel (non-critical - returns status for logging only)
applyBBRSettings() {
  sysctl -w net.core.default_qdisc=fq >/dev/null 2>&1 || true
  if sysctl -w net.ipv4.tcp_congestion_control=bbr >/dev/null 2>&1; then
    sleep 0.2
    isBBRActive && return 0
  fi
  return 1
}

# Ensure BBR is active (non-critical - always returns success)
ensureBBRActive() {
  ! isBBRConfiguredInFile && return 0
  isBBRActive && return 0

  if ! isBBRAvailable && ! loadBBRModule; then
    [[ $SILENT_MODE -eq 0 ]] && {
      echo " - $(colorize yellow "$(msg bbr_activate_failed)")"
      echo "   $(colorize yellow "$(msg bbr_no_optimization)")"
    }
    return 0
  fi

  if applyBBRSettings; then
    [[ $SILENT_MODE -eq 0 ]] && echo " - $(msg bbr_enabled)"
  else
    local current_cc
    current_cc=$(sysctl -n net.ipv4.tcp_congestion_control 2>/dev/null || echo "unknown")
    [[ $SILENT_MODE -eq 0 ]] && {
      echo " - $(colorize yellow "$(msg bbr_activate_failed_cc "$current_cc")")"
      echo "   $(colorize yellow "$(msg bbr_no_optimization)")"
      echo "   $(colorize yellow "$(msg bbr_settings_will_apply "$SYSCTL_BBR_FILE")")"
    }
  fi
  return 0
}

configureBBR() {
  [[ $isBbr -ne 1 ]] && return 0

  # Check if BBR is available or can be loaded first
  if ! isBBRAvailable && ! loadBBRModule; then
    # BBR not available - check if it's already in config
    if isBBRConfiguredInFile; then
      [[ $SILENT_MODE -eq 0 ]] && {
        echo " - $(colorize yellow "$(msg bbr_configured_not_available)")"
        echo "   $(colorize yellow "$(msg bbr_requires_kernel)")"
        echo "   $(colorize yellow "$(msg bbr_settings_will_apply "$SYSCTL_BBR_FILE")")"
        echo "   $(colorize yellow "$(msg bbr_no_optimization)")"
      }
    else
      # Not in config and not available - don't add it
      [[ $SILENT_MODE -eq 0 ]] && {
        echo " - $(colorize yellow "$(msg bbr_not_available)")"
        echo "   $(colorize yellow "$(msg bbr_requires_kernel)")"
        echo "   $(colorize yellow "$(msg bbr_settings_not_added "$SYSCTL_BBR_FILE")")"
        echo "   $(colorize yellow "$(msg bbr_no_optimization)")"
      }
    fi
    return 0
  fi

  # BBR is available - now configure it
  if isBBRConfiguredInFile; then
    [[ $SILENT_MODE -eq 0 ]] && echo " - $(msg bbr_already_configured)"
  else
    ! addBBRToSysctl && {
      [[ $SILENT_MODE -eq 0 ]] && {
        echo " - $(colorize yellow "$(msg bbr_config_failed)")"
        echo "   $(colorize yellow "$(msg bbr_write_failed "$SYSCTL_BBR_FILE")")"
        echo "   $(colorize yellow "$(msg bbr_no_optimization)")"
      }
      return 0
    }
  fi

  ensureBBRModuleAtBoot || true

  if applyBBRSettings; then
    [[ $SILENT_MODE -eq 0 ]] && echo " - $(msg bbr_enabled)"
    return 0
  fi

  local current_cc current_qdisc
  current_cc=$(sysctl -n net.ipv4.tcp_congestion_control 2>/dev/null || echo "unknown")
  current_qdisc=$(sysctl -n net.core.default_qdisc 2>/dev/null || echo "unknown")
  [[ $SILENT_MODE -eq 0 ]] && {
    echo " - $(colorize yellow "$(msg bbr_config_failed)")"
    echo "   $(colorize yellow "$(msg bbr_current_values "$current_qdisc" "$current_cc")")"
    echo "   $(colorize yellow "$(msg bbr_settings_will_apply "$SYSCTL_BBR_FILE")")"
    echo "   $(colorize yellow "$(msg bbr_no_optimization)")"
  }
  return 0
}

readExistingConfig() {
  local config_file="$dirInstall/$serviceName.config"

  if [[ -f "$config_file" ]]; then
    local daemon_options
    daemon_options=$(grep "^DAEMON_OPTIONS=" "$config_file" | cut -d'"' -f2)

    # Extract port - use -- to prevent grep from interpreting pattern as option
    if echo "$daemon_options" | grep -qE -- "--port[[:space:]]+[0-9]+"; then
      servicePort=$(echo "$daemon_options" | grep -oE -- "--port[[:space:]]+[0-9]+" | awk '{print $2}')
    fi

    # Check for auth
    if echo "$daemon_options" | grep -qE -- "--httpauth"; then
      isAuth=1
    else
      isAuth=0
    fi

    # Check for rdb
    if echo "$daemon_options" | grep -qE -- "--rdb"; then
      isRdb=1
    else
      isRdb=0
    fi

    # Check for log
    if echo "$daemon_options" | grep -qE -- "--logpath"; then
      isLog=1
    else
      isLog=0
    fi
  fi
}

configureService() {
  # Read existing config if available (for reconfiguration)
  if [[ -f "$dirInstall/$serviceName.config" ]]; then
    readExistingConfig
  fi

  # Port configuration
  if [[ -z "$servicePort" ]]; then
    local inferred_default="$DEFAULT_PORT"
    if promptYesNo "$(msg change_port)" "n"; then
      servicePort=$(promptInput "$(msg enter_port)" "$inferred_default")
    else
      servicePort="$inferred_default"
    fi
  else
    # Port exists, ask if user wants to change it
    if [[ $SILENT_MODE -eq 0 ]]; then
      if promptYesNo "$(msg change_port)" "n"; then
        servicePort=$(promptInput "$(msg enter_port)" "$servicePort")
      fi
    fi
  fi

  # Auth configuration
  if [[ -z "$isAuth" ]]; then
    if promptYesNo "$(msg enable_auth)" "n"; then
      isAuth=1
    else
      isAuth=0
    fi
  else
    # Auth setting exists, ask if user wants to change it
    if [[ $SILENT_MODE -eq 0 ]]; then
      if promptYesNo "$(msg enable_auth)" "$([[ $isAuth -eq 1 ]] && echo 'y' || echo 'n')"; then
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
        if promptYesNo "$(msg change_auth_credentials)" "n"; then
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
    if promptYesNo "$(msg enable_rdb)" "n"; then
      isRdb=1
    else
      isRdb=0
    fi
  else
    # RDB setting exists, ask if user wants to change it
    if [[ $SILENT_MODE -eq 0 ]]; then
      if promptYesNo "$(msg enable_rdb)" "$([[ $isRdb -eq 1 ]] && echo 'y' || echo 'n')"; then
        isRdb=1
      else
        isRdb=0
      fi
    fi
  fi

  if [[ $isRdb -eq 1 ]] && [[ $SILENT_MODE -eq 0 ]]; then
    echo " $(msg set_readonly)"
    printf ' %s\n' "$(msg readonly_hint "$dirInstall/$serviceName.config")"
  fi

  # Logging configuration
  if [[ -z "$isLog" ]]; then
    if promptYesNo "$(msg enable_log)" "n"; then
      isLog=1
    else
      isLog=0
    fi
  else
    # Log setting exists, ask if user wants to change it
    if [[ $SILENT_MODE -eq 0 ]]; then
      if promptYesNo "$(msg enable_log)" "$([[ $isLog -eq 1 ]] && echo 'y' || echo 'n')"; then
        isLog=1
      else
        isLog=0
      fi
    fi
  fi

  if [[ $isLog -eq 1 ]] && [[ $SILENT_MODE -eq 0 ]]; then
    printf ' - %s\n' "$(msg log_location "$dirInstall/$serviceName.log")"
  fi

  # BBR configuration
  if [[ -z "$isBbr" ]] && ! isBBRConfiguredInFile; then
    if promptYesNo "$(msg enable_bbr)" "n"; then
      isBbr=1
    else
      isBbr=0
    fi
  fi

  # Build the complete config file with all options
  local daemon_options="--port $servicePort"

  if [[ $isRdb -eq 1 ]]; then
    daemon_options="$daemon_options --rdb"
  fi

  if [[ $isLog -eq 1 ]]; then
    daemon_options="$daemon_options --logpath $dirInstall/$serviceName.log"
  fi

  daemon_options="$daemon_options --path $dirInstall"

  if [[ $isAuth -eq 1 ]]; then
    daemon_options="$daemon_options --httpauth"
  fi

  cat << EOF > "$dirInstall/$serviceName.config"
DAEMON_OPTIONS="$daemon_options"
EOF
}

changeServiceUser() {
  local target_user="$1"
  if [[ -z "$target_user" ]]; then
    echo " $(msg error_username_required)"
    exit 1
  fi

  if [[ ! -d "$dirInstall" ]] || [[ ! -f "$dirInstall/$serviceName.config" ]]; then
    echo " - $(msg user_change_missing "$dirInstall")"
    exit 1
  fi

  checkArch

  local normalized_target
  normalized_target=$(echo "$target_user" | tr '[:upper:]' '[:lower:]')
  local default_lower
  default_lower=$(echo "$DEFAULT_USERNAME" | tr '[:upper:]' '[:lower:]')

  if [[ "$normalized_target" == "root" ]]; then
    target_user="root"
    USE_ROOT_USER=1
    username="root"
  elif [[ "$normalized_target" == "$default_lower" ]]; then
    target_user="$DEFAULT_USERNAME"
    USE_ROOT_USER=0
    username="$DEFAULT_USERNAME"
    if ! id "$username" >/dev/null 2>&1; then
      if ! addUser; then
        printf " - %s\n" "$(msg user_add_failed "$username")"
        exit 1
      fi
    fi
  else
    printf " - %s\n" "$(msg user_change_invalid "root" "$DEFAULT_USERNAME")"
    exit 1
  fi

  local owner="$username"
  local group
  if [[ "$username" == "root" ]]; then
    group="root"
  else
    group="$(id -gn "$username" 2>/dev/null || echo "$username")"
  fi

  createServiceFile
  sed -i 's/^[ \t]*//' "$dirInstall/$serviceName.service"
  ln -sf "$dirInstall/$serviceName.service" /usr/local/lib/systemd/system/

  if [[ -d "$dirInstall" ]]; then
    chown -R "$owner":"$group" "$dirInstall"
    if [[ $SILENT_MODE -eq 0 ]]; then
      printf ' - %s\n' "$(msg user_change_permissions "$dirInstall" "$owner" "$group")"
    fi
  fi

  local restart_rc=0
  if ! systemctlCmd daemon-reload; then
    restart_rc=1
  fi
  if ! systemctlCmd restart "$serviceName.service"; then
    restart_rc=1
  fi

  if [[ $SILENT_MODE -eq 0 ]]; then
    printf ' - %s\n' "$(msg user_change_success "$username")"
    if [[ "$username" == "root" ]]; then
      echo " - $(msg running_as_root)"
    else
      printf ' - %s\n' "$(msg running_as_user "$username")"
    fi
    if [[ $restart_rc -eq 1 ]]; then
      echo " - $(colorize yellow "$(msg service_start_failed)")"
    fi
  else
    printf "%s\n" "$(msg user_change_success "$username")"
    if [[ $restart_rc -eq 1 ]]; then
      printf "%s\n" "$(msg service_start_failed)"
    fi
  fi
}

installTorrServer() {
  if [[ $SILENT_MODE -eq 0 && $ROOT_PROMPTED -eq 0 ]]; then
    if [[ $USE_ROOT_USER -ne 1 ]]; then
      if promptYesNo "$(msg prompt_run_as_root)" "n"; then
        USE_ROOT_USER=1
        username="root"
      fi
    fi
    ROOT_PROMPTED=1
  fi

  if [[ $SILENT_MODE -eq 0 ]]; then
    echo " $(msg install_configure)"
  fi

  # Get target version and check glibc compatibility
  local target_version
  target_version=$(getTargetVersion)
  if [[ $SILENT_MODE -eq 0 ]]; then
    echo " - $(msg target_version) $target_version"
  fi

  if ! checkGlibcCompatibility "$target_version"; then
    exit 1
  fi

  # Check if already installed and up to date
  if checkInstalled; then
    if ! checkInstalledVersion; then
      if promptYesNo "$(msg want_update)" "y"; then
        UpdateVersion
        return
      fi
    else
      # Already installed and up to date, allow reconfiguration
      if [[ $SILENT_MODE -eq 0 ]]; then
        echo " - $(msg running_as_user "$username")"
        echo ""
        # Allow user to reconfigure settings
        if promptYesNo "$(msg want_reconfigure)" "n"; then
          # Read existing config first
          if [[ -f "$dirInstall/$serviceName.config" ]]; then
            readExistingConfig
          fi
          # Reconfigure service
          configureService
          # Configure BBR if enabled
          configureBBR
          # Update service file
          createServiceFile
          sed -i 's/^[ \t]*//' "$dirInstall/$serviceName.service"
          sed -i 's/^[ \t]*//' "$dirInstall/$serviceName.config"
          # Reload and restart service
          if ! systemctlCmd daemon-reload; then
            :
          fi
          if ! systemctlCmd restart "$serviceName.service"; then
            :
          fi
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
  fi
  if [[ ! -d "/usr/local/lib/systemd/system" ]]; then
    mkdir -p "/usr/local/lib/systemd/system"
  fi

  # Download binary if needed
  local binName
  binName=$(getBinaryName)
  if [[ ! -f "$dirInstall/$binName" ]] || [[ ! -x "$dirInstall/$binName" ]] || [[ $(stat -c%s "$dirInstall/$binName" 2>/dev/null) -eq 0 ]]; then
    local urlBin
    if [[ -n "$specificVersion" ]]; then
      urlBin=$(buildDownloadUrl "$target_version" "$binName")
    else
      urlBin=$(buildDownloadUrl "latest" "$binName")
    fi
    downloadBinary "$urlBin" "$dirInstall/$binName" "$target_version"
  fi

  # Create service and config files
  createServiceFile
  configureService

  # Configure BBR if enabled (non-critical - always succeeds)
  configureBBR

  # Set up systemd service
  ln -sf "$dirInstall/$serviceName.service" /usr/local/lib/systemd/system/
  sed -i 's/^[ \t]*//' "$dirInstall/$serviceName.service"
  sed -i 's/^[ \t]*//' "$dirInstall/$serviceName.config"

  local service_started=0

  # Start service
  if [[ $SILENT_MODE -eq 0 ]]; then
    echo " $(msg starting_service)"
  fi
  if ! systemctlCmd daemon-reload; then
    :
  fi
  if ! systemctlCmd enable "$serviceName.service"; then
    :
  fi
  if systemctlCmd restart "$serviceName.service"; then
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

    if [[ $username == "root" ]]; then
      echo " $(colorize yellow "$(msg running_as_root)")"
    else
      printf ' %s\n' "$(msg running_as_user "$username")"
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

# Common function to update/downgrade TorrServer version
updateTorrServerVersion() {
  local target_version="$1"
  local cancel_message="$2"
  local use_latest_url="${3:-0}"

  if ! checkGlibcCompatibility "$target_version"; then
    if [[ $SILENT_MODE -eq 0 ]]; then
      echo " - $(msg "$cancel_message")"
    fi
    return 1
  fi

  if ! systemctlCmd stop "$serviceName.service"; then
    :
  fi

  local binName
  binName=$(getBinaryName)
  local urlBin
  if [[ $use_latest_url -eq 1 && -z "$specificVersion" ]]; then
    urlBin=$(buildDownloadUrl "latest" "$binName")
  else
    urlBin=$(buildDownloadUrl "$target_version" "$binName")
  fi

  downloadBinary "$urlBin" "$dirInstall/$binName" "$target_version"

  # Update service file to reflect user change
  if [[ -f "$dirInstall/$serviceName.service" ]]; then
    createServiceFile
    if ! systemctlCmd daemon-reload; then
      :
    fi
  fi

  # Ensure BBR is active before starting service (if previously configured)
  ensureBBRActive

  if ! systemctlCmd start "$serviceName.service"; then
    :
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
  if ! systemctlCmd --quiet stop "$serviceName"; then
    :
  fi
  if ! systemctlCmd --quiet disable "$serviceName"; then
    :
  fi
  rm -rf /usr/local/lib/systemd/system/"$serviceName.service" "$dirInstall" 2>/dev/null
  delUser
}

cleanAll() {
  if ! systemctlCmd --quiet stop torr; then
    :
  fi
  if ! systemctlCmd --quiet stop torrserver; then
    :
  fi
  if ! systemctlCmd --quiet disable torr; then
    :
  fi
  if ! systemctlCmd --quiet disable torrserver; then
    :
  fi
  rm -rf /home/torrserver 2>/dev/null
  rm -rf /usr/local/torr 2>/dev/null
  rm -rf /opt/torr* 2>/dev/null
  rm -f /{,etc,usr/local/lib}/systemd/system/tor{,r,rserver}.service 2>/dev/null
}

uninstall() {
  checkArch
  checkInstalled

  if [[ $SILENT_MODE -eq 1 ]]; then
    cleanup
    cleanAll
    echo " - $(msg uninstalled)"
    return
  fi

  echo ""
  echo " $(msg install_dir_label) ${dirInstall}"
  echo ""
  echo " $(msg uninstall_warning)"
  echo ""

  if promptYesNo "$(msg confirm_delete)" "n"; then
    cleanup
    cleanAll
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

  # Set username based on USE_ROOT_USER flag
  if [[ $USE_ROOT_USER -eq 1 ]]; then
    username="root"
  fi

  if [[ $SILENT_MODE -eq 0 ]]; then
    echo " - $(msg running_as_user "$username")"
    echo ""
  fi

  # Read existing config first
  if [[ -f "$dirInstall/$serviceName.config" ]]; then
    readExistingConfig
  fi

  # Reconfigure service
  configureService

  # Configure BBR if enabled
  configureBBR

  # Update service file
  createServiceFile
  sed -i 's/^[ \t]*//' "$dirInstall/$serviceName.service"
  sed -i 's/^[ \t]*//' "$dirInstall/$serviceName.config"

  # Reload and restart service
  if ! systemctlCmd daemon-reload; then
    :
  fi
  if ! systemctlCmd restart "$serviceName.service"; then
    :
  fi

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
  -C, --change-user USER            Change TorrServer service user (root|torrserver)
    change-user USER
  -h, --help                        Show this help message
    help

Options:
  --root                            Run service as root user
  --silent                          Non-interactive mode with defaults

Examples:
  # Install latest version interactively
  sudo $scriptname --install

  # Install specific version as root user silently
  sudo $scriptname --install 135 --root --silent

  # Update with silent mode
  sudo $scriptname --update --silent

  # Check for updates
  sudo $scriptname --check

  # Uninstall silently
  sudo $scriptname --remove --silent

  # Reconfigure TorrServer settings interactively
  sudo $scriptname --reconfigure

  # Switch service to run as root
  sudo $scriptname -C root

  # Switch service back to torrserver user
  sudo $scriptname --change-user torrserver

Default Settings (silent mode):
  - Port: ${portOverride:-$DEFAULT_PORT}
  - User: torrserver (or root with --root flag)
  - Auth: disabled
  - Read-only mode: disabled
  - Logging: disabled
  - BBR: enabled

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
      -C|--change-user|change-user)
        parsedCommand="change_user"
        shift
        if [[ $# -gt 0 ]]; then
          changeUserName="$1"
          shift
        else
          echo " $(msg error_username_required)"
          exit 1
        fi
        ;;
      -h|--help|help)
        getLang  # Set language before showing help
        helpUsage
        exit 0
        ;;
      --root)
        USE_ROOT_USER=1
        shift
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
        isBbr=1
      fi

      if [[ $USE_ROOT_USER -eq 1 ]]; then
        username="root"
        if [[ $SILENT_MODE -eq 0 ]]; then
          echo " - $(msg running_as_root)"
        fi
      fi

      if ! checkInstalled; then
        installTorrServer
      else
        createServiceFile
        if ! systemctlCmd daemon-reload; then
          :
        fi
        if ! systemctlCmd stop "$serviceName.service"; then
          :
        fi
        if ! systemctlCmd start "$serviceName.service"; then
          :
        fi
        if [[ $SILENT_MODE -eq 0 ]]; then
          echo " - $(msg service_reconfigured_user "$username")"
        fi
      fi
      exit 0
      ;;
    update)
      initialCheck
      if [[ $USE_ROOT_USER -eq 1 ]]; then
        username="root"
      fi
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
      if [[ $USE_ROOT_USER -eq 1 ]]; then
        username="root"
      fi
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
      if [[ $USE_ROOT_USER -eq 1 ]]; then
        username="root"
      fi
      reconfigureTorrServer
      exit 0
      ;;
    change_user)
      if [[ -z "$changeUserName" ]]; then
        echo " $(msg error_username_required)"
        exit 1
      fi
      if ! isRoot; then
        echo " $(msg need_root)"
        exit 1
      fi
      changeServiceUser "$changeUserName"
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
    user_choice=$(promptYesNoDelete "$(msg want_install)" "n")

    if [[ "$user_choice" == "delete" ]]; then
      initialCheck
      uninstall
    elif [[ "$user_choice" == "yes" ]]; then
      initialCheck

      if promptYesNo "$(msg prompt_run_as_root)" "n"; then
        USE_ROOT_USER=1
        username="root"
      fi
      ROOT_PROMPTED=1

      installTorrServer
    fi
  fi

  echo " $(msg have_fun)"
  echo ""
}

# Run main function
main "$@"
