{
  description = "wtf - Terraform version manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "wtf";
          version = "1.0.0";
          src = ./.;
          vendorHash = "sha256-JgC0pv3Wy+jZqgF6LqZ5OujCDW0CkBYoQvjxERJIbEA=";
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
          ];
        };
      }
    );
}
