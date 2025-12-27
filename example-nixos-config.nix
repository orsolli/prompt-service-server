# Example NixOS configuration for Prompt Service Server
# Add this to your /etc/nixos/configuration.nix or import it as a module

{ config, lib, pkgs, ... }:

let
  # Import the prompt-service-server package and module
  promptService = import ./default.nix { inherit pkgs; };
  promptServiceModule = import ./nixos-module.nix;
in
{
  # Import the prompt-service-server module
  imports = [ promptServiceModule ];

  # Enable and configure the service
  services.prompt-service-server = {
    enable = true;
    port = 3000;  # Change from default 8080 if needed
    csrfTokenSecret = "your-csrf-secret-here";     # Generate a secure random key
  };

  # Optional: Configure firewall
  networking.firewall.allowedTCPPorts = [ 3000 ];

  # Optional: Configure nginx reverse proxy with SSL
  services.nginx = {
    enable = true;
    recommendedProxySettings = true;
    virtualHosts."your-domain.com" = {
      forceSSL = true;
      enableACME = true;
      locations."/" = {
        proxyPass = "http://127.0.0.1:3000";
        proxyWebsockets = true;
      };
    };
  };
}