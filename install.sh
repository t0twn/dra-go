#!/usr/bin/env bash

set -euo pipefail

readonly GITHUB_REPOSITORY="t0twn/dra-go"

info(){
  echo -e "$1" >&2
}

error(){
  echo -e "$1" >&2
  exit 1
}

check_command(){
  if ! command -v "$1" &>/dev/null; then
    return 1
  fi
}

check_dependencies(){
  local os=$1

  if ! check_command curl && ! check_command wget; then
    error "Missing 'curl' and 'wget'"
  fi

  check_command grep || error "Missing 'grep'"
  check_command cut || error "Missing 'cut'"
  check_command tar || error "Missing 'tar'"
}

http_get(){
  local url="$1"
  local output_path="$2"

  if command -v curl &>/dev/null; then
    curl --proto =https --tlsv1.2 -sSfL -o "$output_path" "$url"
  else
    wget --https-only --secure-protocol=TLSv1_2 --quiet -O "$output_path" "$url"
  fi
}

load_latest_release(){
  http_get "https://api.github.com/repos/$GITHUB_REPOSITORY/releases/latest" /dev/stdout |
    grep tag_name |
    cut -d'"' -f4
}

asset_version(){
  local tag=$1
  # Strip leading 'v' if present, then append '-go' suffix (Go port disambiguator)
  tag="${tag#v}"
  case "$tag" in
    *-go) echo "$tag" ;;
    *)    echo "${tag}-go" ;;
  esac
}

get_os(){
  local os
  os=$(uname -s)
  case "$os" in
    Darwin*) echo "darwin" ;;
    Linux*) echo "linux" ;;
    *) error "Unsupported operating system: $os" ;;
  esac
}

get_arch(){
  local arch
  arch=$(uname -m)
  case "$arch" in
    x86_64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    armv6l|armv7l) echo "arm" ;;
    *) error "Unsupported architecture: $arch" ;;
  esac
}

get_asset_name(){
  local version=$1
  local os=$2
  local arch=$3

  echo "dra_${version}_${os}_${arch}.tar.gz"
}

download_asset(){
  local version=$1
  local asset=$2
  local temp_dir=$3
  local output_path="$temp_dir/$asset"

  http_get "https://github.com/$GITHUB_REPOSITORY/releases/download/$version/$asset" "$output_path"
  echo "$output_path"
}

extract_archive(){
  local asset_path=$1
  local output_dir=$2

  case "$asset_path" in
    *tar.gz) tar xzf "$asset_path" --strip-components=1 -C "$output_dir" ;;
    *) error "Unknown archive $asset_path" ;;
  esac
}

copy_executable(){
  local asset_dir=$1
  local destination=$2

  cp "$asset_dir"/dra "$destination"
  chmod +x "$destination/dra"
}

help(){
  cat <<'EOF'
Install latest release of dra (Go version) from GitHub Releases

USAGE:
    install.sh [options]

FLAGS:
    -h, --help      Display this message

OPTIONS:
    --to <DESTINATION>   Save dra to custom path [default: current working directory]
EOF
}

installation_completed(){
  cat <<EOF

Thanks for installing dra (Go version)!

You can run \`dra --help\` to get started and see useful examples.

More examples can be found in the documentation:
- https://github.com/$GITHUB_REPOSITORY#usage
- https://github.com/$GITHUB_REPOSITORY#examples
EOF
}

main(){
  local destination="$PWD"

  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        help
        exit 0
        ;;
      --to)
        destination="$2"
        shift
        shift
        ;;
      *)
        error "Unknown option $1"
        ;;
    esac
  done

  local os arch
  os=$(get_os)
  arch=$(get_arch)
  check_dependencies "$os"

  local tag version asset
  tag=$(load_latest_release)
  version=$(asset_version "$tag")
  asset=$(get_asset_name "$version" "$os" "$arch")

  info "OS:           $os/$arch"
  info "Repository:   $GITHUB_REPOSITORY"
  info "Release:      $tag"
  info "Version:      $version"
  info "Asset:        $asset"

  info "\nDownloading $asset"
  local temp_dir asset_path
  temp_dir=$(mktemp -d)
  asset_path=$(download_asset "$tag" "$asset" "$temp_dir")

  info "Extracting archive $asset_path"
  extract_archive "$asset_path" "$temp_dir"
  copy_executable "$temp_dir" "$destination"

  info "dra saved to $destination"
  installation_completed
}

main "$@"
