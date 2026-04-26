{
  self,
}:
{
  config,
  lib,
  pkgs,
  ...
}:
let
  inherit (lib.modules) mkIf mkMerge;
  inherit (lib.options) mkOption mkEnableOption mkPackageOption;
  inherit (lib.types) nullOr bool;
  inherit (lib)
    optionalString
    types
    mapAttrs'
    mapAttrsToList
    nameValuePair
    mkDefault
    literalExpression
    ;

  cfg = config.services.torrserver;

in
{
  options.services.torrserver = {
    enable = mkEnableOption "TorrServer service";

    package = mkOption {
      type = types.package;
      default = self.packages.${pkgs.system}.torrserver;
      description = "TorrServer package to use";
    };

    port = mkOption {
      type = types.int;
      default = 8090;
      description = "Web server port";
    };

    address = mkOption {
      type = types.str;
      default = "0.0.0.0";
      description = "Web server address to bind to";
    };

    dataDir = mkOption {
      type = types.path;
      default = "${config.home.homeDirectory}/.local/share/torrserver";
      description = "Directory for TorrServer data and cache";
    };

    logDir = mkOption {
      type = types.path;
      default = "${config.home.homeDirectory}/.local/log/torrserver";
      description = "Directory for TorrServer logs";
    };

    enableAuth = mkOption {
      type = types.bool;
      default = false;
      description = "Enable HTTP authentication";
    };

    enableWebDAV = mkOption {
      type = types.bool;
      default = false;
      description = "Enable WebDAV server";
    };

    enableDLNA = mkOption {
      type = types.bool;
      default = false;
      description = "Enable DLNA server";
    };

    maxCacheSize = mkOption {
      type = types.nullOr types.str;
      default = null;
      example = "10GB";
      description = "Maximum cache size (e.g., 10GB, 5000MB)";
    };

    torrentsDir = mkOption {
      type = types.nullOr types.path;
      default = null;
      description = "Directory for auto-loading torrents";
    };

    ssl = mkOption {
      type = types.bool;
      default = false;
      description = "Enable HTTPS";
    };

    sslPort = mkOption {
      type = types.int;
      default = 8091;
      description = "HTTPS port (requires ssl = true)";
    };

    sslCert = mkOption {
      type = types.nullOr types.path;
      default = null;
      description = "Path to SSL certificate file";
    };

    sslKey = mkOption {
      type = types.nullOr types.path;
      default = null;
      description = "Path to SSL key file";
    };

    telegramToken = mkOption {
      type = types.nullOr types.str;
      default = null;
      description = "Telegram bot token";
    };

    readOnly = mkOption {
      type = types.bool;
      default = false;
      description = "Run in read-only database mode";
    };

    extraArgs = mkOption {
      type = types.listOf types.str;
      default = [ ];
      description = "Additional command-line arguments";
    };
  };

  config = mkIf cfg.enable {
    home.activation.torrserverDirs = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      mkdir -p ${cfg.dataDir} ${cfg.logDir}
    '';

    systemd.user.services.torrserver = {
      Unit = {
        Description = "TorrServer - Stream torrents online";
        After = [ "network-online.target" ];
        Wants = [ "network-online.target" ];
      };

      Service = {
        Type = "simple";
        ExecStart =
          let
            args = [
              "-p ${toString cfg.port}"
              "-i ${cfg.address}"
              "-d ${cfg.dataDir}"
              "-l ${cfg.logDir}/server.log"
              "-w ${cfg.logDir}/web.log"
            ]
            ++ lib.optionals cfg.enableAuth [ "-a" ]
            ++ lib.optionals cfg.readOnly [ "-r" ]
            ++ lib.optionals cfg.enableWebDAV [ "--webdav" ]
            ++ lib.optionals cfg.ssl [
              "--ssl"
              "--ssl-port ${toString cfg.sslPort}"
            ]
            ++ lib.optionals (cfg.sslCert != null) [ "--ssl-cert ${cfg.sslCert}" ]
            ++ lib.optionals (cfg.sslKey != null) [ "--ssl-key ${cfg.sslKey}" ]
            ++ lib.optionals (cfg.maxCacheSize != null) [ "-m ${cfg.maxCacheSize}" ]
            ++ lib.optionals (cfg.torrentsDir != null) [ "-t ${cfg.torrentsDir}" ]
            ++ lib.optionals (cfg.telegramToken != null) [ "-T ${cfg.telegramToken}" ]
            ++ cfg.extraArgs;
          in
          "${cfg.package}/bin/torrserver ${lib.strings.concatStringsSep " " args}";

        Restart = "always";
        RestartSec = "10s";
      };

      Install = {
        WantedBy = [ "default.target" ];
      };
    };
  };
}
