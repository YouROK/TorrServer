{
  description = "TorrServer - Simple and powerful tool for streaming torrents";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/default-linux";
  };

  outputs =
    {
      self,
      nixpkgs,
      systems,
      ...
    }:
    let
      inherit (nixpkgs) lib;
      eachSystem = f: lib.genAttrs (import systems) (system: f nixpkgs.legacyPackages.${system});
    in
    {
      formatter = eachSystem (pkgs: pkgs.alejandra);

      devShells = eachSystem (pkgs: {
        default = pkgs.mkShell {
          name = "torrserver";
          inputsFrom = [ self.packages.${pkgs.stdenv.system}.torrserver ];
        };
      });

      packages = eachSystem (pkgs: {
        default = self.packages.${pkgs.stdenv.system}.torrserver;
        torrserver = pkgs.callPackage ./nix/packages/torrserver.nix { };
      });

      homeModules = {
        default = self.homeModules.torrserver;
        torrserver = import ./nix/modules/home-manager.nix { inherit self; };
      };

      nixosModules = {
        default = self.nixosModules.torrserver;
        torrserver = import ./nix/modules/nixos.nix { inherit self; };
      };
    };
}
