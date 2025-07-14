{
  description = "Wait4X allows you to wait for a port or a service to enter the requested state.";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      packageName = "wait4x";
      version = "${self.shortRev or self.dirtyShortRev or "dirty"}";
    in {
      formatter = pkgs.alejandra;
      devShells.default = pkgs.mkShell {
        name = packageName;
        buildInputs = with pkgs; [
          go
          gotools
          golint
          gopls
          revive
          golangci-lint
          delve
          git
          gh
          gnumake
        ];
      };
      packages.default = pkgs.buildGoModule {
        pname = packageName;
        inherit version;
        src = self;
        vendorHash = "sha256-ODcHrmmHHeZbi1HVDkYPCyHs7mcs2UGdBzicP1+eOSI=";
        doCheck = false;
        nativeBuildInputs = with pkgs; [git];
        GOCACHE = "$(mktemp -d)";
      };
    });
}
