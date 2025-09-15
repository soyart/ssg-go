{
  description = "Static site generator library - Go implementation";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    ssg-testdata = {
      url = "github:soyart/ssg-testdata";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, ssg-testdata }:
    let
      homepage = "https://github.com/soyart/ssg-go";

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
          
          src = ./.;
          
          # Go module configuration
          vendorHash = "sha256-P8vB0khyNGjYNmYwn/AfzKyB+CaK/lhcMPsO9UmDNSQ="; # This will be updated by nix
          
          # Build configuration
          buildPhase = ''
            runHook preBuild
            
            # Ensure test data is available during build
            ln -sf ${ssg-testdata} ./testdata
            
            # Build the main binary
            go build -o $out/bin/ssg ./cmd/ssg
            
            runHook postBuild
          '';
          
          # Install configuration
          installPhase = ''
            runHook preInstall
            
            mkdir -p $out/bin
            cp ssg $out/bin/ssg
            
            runHook postInstall
          '';
          
          # Test configuration
          checkPhase = ''
            runHook preCheck
            
            # Ensure test data is available for tests
            ln -sf ${ssg-testdata} ./testdata
            
            # Run tests
            go test -v ./...
            
            runHook postCheck
          '';
          
          meta = {
            inherit homepage;
            description = "Static site generator library - Go implementation";
            license = pkgs.lib.licenses.mit;
            maintainers = [ ];
            platforms = pkgs.lib.platforms.unix;
          };
        };

      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            # Nix development tools
            nixd
            nixpkgs-fmt
            
            # Go development tools
            go
            gopls
            gotools
            go-tools
            
            # Testing tools
            ginkgo
            gomega
          ];
          
          # Environment variables
          shellHook = ''
            echo "ssg-go development environment"
            echo "Go version: $(go version)"
            echo "Test data will be available at ./testdata (submodule)"
            
            # Check if testdata submodule exists
            if [ ! -d "./testdata" ]; then
              echo "Warning: testdata submodule not found. Run:"
              echo "  git submodule add https://github.com/soyart/ssg-testdata.git testdata"
            fi
          '';
        };
      });

      # Nix flake checks
      checks = forAllSystems ({ pkgs }: {
        # Build check
        build = self.packages.${pkgs.system}.default;
        
        # Test check
        test = pkgs.runCommand "ssg-go-tests" {
          nativeBuildInputs = with pkgs; [ go ];
        } ''
          cp -r ${self}/* .
          chmod -R +w .
          
          # Ensure test data is available
          ln -sf ${ssg-testdata} ./testdata
          
          # Run tests
          go test -v -count=1 -race ./...
          
          touch $out
        '';
        
        # Lint check
        lint = pkgs.runCommand "ssg-go-lint" {
          nativeBuildInputs = with pkgs; [ go-tools ];
        } ''
          cp -r ${self}/* .
          chmod -R +w .
          
          # Run linter
          golangci-lint run
          
          touch $out
        '';
      });
    };
}
