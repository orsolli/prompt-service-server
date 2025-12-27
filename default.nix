{ pkgs ? import <nixpkgs> {} }:

let
  # Build the Go application
  prompt-service-server = pkgs.buildGoModule {
    pname = "prompt-service-server";
    version = "0.1.0";
    src = ./.;

    vendorHash = "sha256-j4cWJ+VaVENawhdLBMjDKMyttJX+aNHO80LW127zBjI=";

    # Build flags for security
    ldflags = [
      "-s" # Strip debug info
      "-w" # Strip DWARF info
    ];

    meta = with pkgs.lib; {
      description = "A secure web service for managing prompts";
      license = licenses.mit;
      maintainers = [ ];
    };
  };

in {
  # The main package
  inherit prompt-service-server;

  # Development shell
  devShell = pkgs.mkShell {
    buildInputs = with pkgs; [
      go
      gopls
      golangci-lint
    ];

    shellHook = ''
      echo "Prompt Service Server development environment"
      echo "Run 'go run main.go' to start the server"
    '';
  };
}