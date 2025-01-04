{
  description = "go-event-bus-flake";

  inputs = { nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable"; };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        system = system;
        config.allowUnfree = true;
      };
    in {
      devShells.${system}.default =
        pkgs.mkShell { nativeBuildInputs = with pkgs; [ go gopls ]; };
    };
}
