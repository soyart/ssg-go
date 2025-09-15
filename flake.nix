rec {
  description = "Static site generator ssg-go";
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    ssg-testdata = {
      url = "github:soyart/ssg-testdata";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;

      # The set of systems to provide outputs for
      allSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      # A function that provides a system-specific Nixpkgs for the desired systems
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        pkgs = import nixpkgs { inherit system; };
      });
    in

    {
      packages = forAllSystems ({ pkgs }: {
        default = pkgs.buildGoModule {
          inherit version;
          pname = "ssg-go";
          preCheck = ''
            ln -sf ${inputs.ssg-testdata} ./ssg-testdata;
          '';
          src = ./.;
          vendorHash = "sha256-P8vB0khyNGjYNmYwn/AfzKyB+CaK/lhcMPsO9UmDNSQ=";
          meta = {
            homepage = "https://github.com/soyart/ssg";
            description = "${description} (go implementation)";
          };
        };
      });

      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            nixd
            nixpkgs-fmt

            bash-language-server
            shellcheck
            shfmt

            coreutils
            lowdown

            go
            gopls
            gotools
            go-tools
          ];
        };
      });
    };
}
