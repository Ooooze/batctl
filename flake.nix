{
  description = "batctl - battery control TUI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.1.0";
      in {
        packages.default = pkgs.buildGoModule {
          pname = "batctl";
          inherit version;

          src = ./.;

          vendorHash = "sha256-irJksXupZGHzZ5vbFeI9laKi5+LyATc1lMxpMLLl69w=";

          ldflags = [ "-s" "-w" "-X main.version=${version}" ];

          meta = with pkgs.lib; {
            description = "Battery control TUI";
            homepage = "https://github.com/Ooooze/batctl";
            license = licenses.mit;
            mainProgram = "batctl";
            platforms = platforms.linux;
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
          ];
        };
      });
}
