{
  description = "Logic gates";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-23.11";
  };

  outputs = { self, nixpkgs }:
  let 
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages."${system}";
  in {
    devShells."${system}".default = pkgs.mkShell {
      nativeBuildInputs = with pkgs; [
        go
      ];
      buildInputs = with pkgs; [
        libGL
        xorg.libXi
        wayland
        libxkbcommon
        xorg.libX11
        xorg.libXcursor
        xorg.libXrandr
        xorg.libXinerama
      ];
    };

    packages."${system}" = {
      default = pkgs.buildGoModule {
        pname = "Raylib-go";
        version = "0.0.1";
        src = self;
        vendorHash = null;
      };
    };
  };
}
