# Installation

Download the application for the required platform in the [releases](https://github.com/YouROK/TorrServer/releases) page. After installation, open the link <http://127.0.0.1:8090> in the browser. iOS is shipped as `TorrServer-ios-TorrServerKit.xcframework.zip` for embedding in a host app (see [Build → iOS](development.md#ios)).

Standard binaries are named `TorrServer-<platform>-<arch>` (for example, `TorrServer-linux-amd64`). From release **141.10**, optional **GStreamer (gst)** builds are named `TorrServer-gst-<platform>-<arch>`. Supported gst targets are Windows amd64, Linux amd64/arm64, and macOS amd64/arm64. The install scripts below can install either variant when the matching release asset is available.

## Windows

Run `TorrServer-windows-amd64.exe`.

## Linux

**Interactive install** (download first — preferred):

```bash
curl -fsSL https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh -o installTorrServerLinux.sh && chmod 755 installTorrServerLinux.sh && sudo bash ./installTorrServerLinux.sh
```

Other interactive one-liners (avoid `curl | sudo bash`, which hangs on modern `sudo` with `use_pty`):

```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh)"
```

```bash
sudo bash <(curl -fsSL https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh)
```

**Non-interactive** pipe is fine when there are no prompts:

```bash
curl -fsSL https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh | sudo bash -s -- --install --silent
```

The script supports interactive and non-interactive installation, configuration, updates, and removal. When running the script interactively, you can:

- **Install/Update**: Choose to install or update TorrServer
- **GStreamer build**: For releases 141.10+ on amd64/arm64, choose the gst build with transcoding support (or pass `--gst`)
- **Reconfigure**: If TorrServer is already installed, you'll be prompted to reconfigure settings (port, auth, read-only mode, logging, BBR)
- **Uninstall**: Type `Delete` (or `Удалить` in Russian) to uninstall TorrServer

**Command-line examples:**

- Install a specific version:

  ```bash
  sudo bash ./installTorrServerLinux.sh --install 135 --silent
  ```

- Update to latest version:

  ```bash
  sudo bash ./installTorrServerLinux.sh --update --silent
  ```

- Install GStreamer build (141.10+):

  ```bash
  sudo bash ./installTorrServerLinux.sh --install --gst
  sudo bash ./installTorrServerLinux.sh --update --gst --silent
  ```

- Reconfigure settings interactively:

  ```bash
  sudo bash ./installTorrServerLinux.sh --reconfigure
  ```

- Check for updates:

  ```bash
  sudo bash ./installTorrServerLinux.sh --check
  ```

- Downgrade to a specific version:

  ```bash
  sudo bash ./installTorrServerLinux.sh --down 135
  ```

- Remove/uninstall:

  ```bash
  sudo bash ./installTorrServerLinux.sh --remove --silent
  ```

- Change the systemd service user:

  ```bash
  sudo bash ./installTorrServerLinux.sh --change-user root --silent
  ```

**All available commands:**

- `--install [VERSION]` - Install latest or specific version
- `--update` - Update to latest version
- `--reconfigure` - Reconfigure TorrServer settings (port, auth, read-only mode, logging, BBR)
- `--check` - Check for updates (version info only)
- `--down VERSION` - Downgrade to specific version
- `--remove` - Uninstall TorrServer
- `--change-user USER` - Change service user (root|torrserver)
- `--gst` - Install GStreamer build with transcoding support (141.10+, amd64/arm64 only)
- `--root` - Run service as root user
- `--silent` - Non-interactive mode with defaults
- `--help` - Show help message

## macOS

Run in Terminal.app

```bash
curl -s https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerMac.sh -o installTorrServerMac.sh && chmod 755 installTorrServerMac.sh && bash ./installTorrServerMac.sh
```

The macOS install script supports the same commands as the Linux script, including `--install`, `--update`, `--remove`, `--reconfigure`, and `--gst` for the GStreamer build (141.10+).

**Command-line examples:**

- Install latest version:

  ```bash
  bash ./installTorrServerMac.sh --install
  ```

- Install GStreamer build:

  ```bash
  bash ./installTorrServerMac.sh --install --gst
  ```

- Update silently:

  ```bash
  bash ./installTorrServerMac.sh --update --silent
  ```

Alternative install script for Intel Macs: <https://github.com/dancheskus/TorrServerMacInstaller>

## IOCage Plugin (Unofficial)

On FreeBSD (TrueNAS/FreeNAS) you can use this plugin: <https://github.com/filka96/iocage-plugin-TorrServer>

## NAS Systems (Unofficial)

- Several releases are available through this link: <https://github.com/vladlenas>
- **Synology NAS** packages repo source: <https://grigi.lt>
