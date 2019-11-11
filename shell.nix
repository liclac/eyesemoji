{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  propagatedBuildInputs = [ pkgs.bluezFull pkgs.dbus ];
  nativeBuildInputs = [ pkgs.pkg-config ];
}
