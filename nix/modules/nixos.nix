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
      default = "/var/lib/torrserver";
      description = "Directory for TorrServer data and cache";
    };

    logDir = mkOption {
      type = types.path;
      default = "/var/log/torrserver";
      description = "Directory for TorrServer logs";
    };

    user = mkOption {
      type = types.str;
      default = "torrserver";
      description = "User to run TorrServer as";
    };

    group = mkOption {
      type = types.str;
      default = "torrserver";
      description = "Group to run TorrServer as";
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
    users.users = mkIf (cfg.user == "torrserver") {
      torrserver = {
        isSystemUser = true;
        group = cfg.group;
        home = cfg.dataDir;
        createHome = true;
      };
    };

    users.groups = mkIf (cfg.group == "torrserver") {
      torrserver = { };
    };

    systemd.tmpfiles.rules = [
      "d ${cfg.logDir} 0755 ${cfg.user} ${cfg.group} - -"
    ];

    systemd.services.torrserver = {
      description = "TorrServer - Stream torrents online";
      after = [ "network-online.target" ];
      wants = [ "network-online.target" ];
      wantedBy = [ "multi-user.target" ];

      serviceConfig = {
        Type = "simple";
        User = cfg.user;
        Group = cfg.group;

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

        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [
          cfg.dataDir
          cfg.logDir
        ];

        LimitNOFILE = 65536;
        LimitNPROC = 512;
      };
    };

    networking.firewall.allowedTCPPorts = [ cfg.port ] ++ lib.optionals cfg.ssl [ cfg.sslPort ];
  };
}
