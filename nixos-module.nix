{ config, lib, pkgs, ... }:

# NixOS module for Prompt Service Server
# Note: Static files (HTML, CSS, JS) are embedded in the binary at build time.
# The service does not require filesystem access to serve static content.

with lib;
let
  cfg = config.services.prompt-service-server;
  pkg = (import ./default.nix { inherit pkgs; }).prompt-service-server;
in {
  options.services.prompt-service-server = {
    enable = mkEnableOption "Prompt Service Server";

    port = mkOption {
      type = types.port;
      default = 8080;
      description = "Port to listen on";
    };

    csrfTokenSecret = mkOption {
      type = types.str;
      description = ''
        CSRF token secret.

        WARNING: Do not hardcode secrets in your configuration!
        Use a secret management tool such as agenix or sops-nix to securely provide this value.
      '';
    };

    allowedOrigins = mkOption {
      type = types.str;
      default = "";
      description = ''
        Comma-separated list of allowed origins for CORS.
        
        If empty, CORS will only be enabled for POST /api/prompts (unrestricted).
        If set, these origins will be allowed for all other endpoints.
        Use "*" to allow all origins (not recommended for production).
        
        Example: "https://example.com,https://app.example.com"
      '';
    };

    user = mkOption {
      type = types.str;
      default = "prompt-service";
      description = "User to run the service as";
    };

    group = mkOption {
      type = types.str;
      default = "prompt-service";
      description = "Group to run the service as";
    };
  };

  config = mkIf cfg.enable {
    # Create the user and group
    users.users.${cfg.user} = {
      isSystemUser = true;
      group = cfg.group;
      createHome = false;
      description = "Prompt Service Server user";
    };

    users.groups.${cfg.group} = {};

    # Systemd service with hardened security
    systemd.services.prompt-service-server = {
      description = "Prompt Service Server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        ExecStart = "${pkg}/bin/prompt-service-server";
        User = cfg.user;
        Group = cfg.group;

        # Security hardening
        NoNewPrivileges = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectHome = true;
        ProtectSystem = "strict";
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [ "AF_INET" "AF_INET6" ];
        RestrictNamespaces = true;
        MemoryDenyWriteExecute = true;
        LockPersonality = true;

        # Environment variables
        Environment = [
          "PORT=${toString cfg.port}"
          "CSRF_TOKEN_SECRET=${cfg.csrfTokenSecret}"
          "ALLOWED_ORIGINS=${cfg.allowedOrigins}"
        ];

        # Restart on failure
        Restart = "on-failure";
        RestartSec = 5;

        # Limits
        LimitNOFILE = 1024;
      };
    };
  };
}