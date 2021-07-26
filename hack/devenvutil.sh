#!/usr/bin/env bash

#
#   OqtaDrive - Sinclair Microdrive emulator
#   Copyright (c) 2021, Alexander Vollschwitz
#
#   This file is part of OqtaDrive.
#
#   OqtaDrive is free software: you can redistribute it and/or modify
#   it under the terms of the GNU General Public License as published by
#   the Free Software Foundation, either version 3 of the License, or
#   (at your option) any later version.
#
#   OqtaDrive is distributed in the hope that it will be useful,
#   but WITHOUT ANY WARRANTY; without even the implied warranty of
#   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
#   GNU General Public License for more details.
#
#   You should have received a copy of the GNU General Public License
#   along with OqtaDrive. If not, see <http://www.gnu.org/licenses/>.
#

#
# Note: All variables defined in Makefile can be directly accessed here.
#

# shellcheck disable=SC2034
{
# formatting
BLD="\e[1m"
DIM="\e[2m"
ITL="\e[3m"
NRM="\e[0m"
OK="\e[01;32m"
ERR="\e[01;31m"
}

#
#
#
function synopsis {

    files=()

    command -v gawk > /dev/null || echo \
        "Note: proper help display requires gawk! (e.g. sudo apt install gawk)"

    for file in ${MAKEFILE_LIST}; do
        if [[ "$(basename "${file}")" == "Makefile" ]]; then
            files+=( "../${file}" )
        fi
    done

    echo -e "\n${BLD}TARGETS${NRM}"
    print_sorted_help "$(cat "${files[@]}" \
        | gawk '{FS=":"}
            /^[a-zA-Z0-9][-a-zA-Z0-9_\.]+:{1,2}[-a-zA-Z0-9_\. ]*$/{f=1; printf "\n${ITL}${BLD}%s${NRM}\n", $1; next}
            /^[^#].*$/{f=0} f' \
        | tr -d '#')"

    echo -e "\n${BLD}NOTES${NRM}\n"

    # .makerc settings
    print_formatted_help "$(cat "${files[@]}" \
        | gawk '/^## makerc$/{f=1; next} /^[^#].*$/{f=0} /^$/{f=0} f' \
        | tr -d '#')"
    echo

    # env settings
    print_formatted_help "$(cat "${files[@]}" \
        | gawk '/^## env$/{f=1; next} /^[^#].*$/{f=0} /^$/{f=0} f' \
        | tr -d '#')"

    # other notes
    print_formatted_help "$(cat "${files[@]}" \
        | gawk '/^##$/{f=1; printf "%s", $0; next} /^[^#].*$/{f=0} /^$/{f=0} f' \
        | tr -d '#')"
    echo
}

#
# $1    help text
#
function print_sorted_help {
    print_formatted_help "$1" \
        | gawk 'BEGIN{print "\0"}
            /^$/{printf "\0"} {print $0}' \
        | sort -z \
        | tr -d '\000' \
        | tail -n+2
}

#
# $1    help text
#
function print_formatted_help {
    echo -e "$(apply_shell_expansion "$1")" | uniq
}

#
# $1    string to expand
#
function apply_shell_expansion {
    declare data="$1"
    declare delimiter="__apply_shell_expansion_delimiter__"
    declare command="cat <<${delimiter}"$'\n'"${data}"$'\n'"${delimiter}"
    eval "${command}"
}

#
# build command binary
#
# $1    command
# $2    target OS
# $3    target architecture; omit for `amd64`
# $4    `keep` for keeping the binary, not just the archive; requires $3
#
function build_binary {

    local arch="amd64"
    [[ -z "$3" ]] || arch="$3"

    local suffix
    [[ "$2" != "windows" ]] || suffix=".exe"

    local binary="${BINARIES}/$1"

    local extra_env
    [[ "${arch}" != "arm" ]] || extra_env="-e GOARM=6"

    echo -e "\nbuilding ${binary} for $2/${arch}"

    # shellcheck disable=SC2086
    docker run --rm --user "$(id -u):$(id -g)" \
        -v "${ROOT}/${BINARIES}:/go/bin" ${CACHE_VOLS} \
        -v "${ROOT}:/go/src/${REPO}" -w "/go/src/${REPO}" \
        -e CGO_ENABLED=0 -e GOOS="$2" -e GOARCH="${arch}" ${extra_env} \
        "${GO_IMAGE}" go build -v -tags netgo -installsuffix netgo \
        -ldflags "-w -X main.OqtaDriveVersion=${OQTADRIVE_VERSION}" \
        -o "${binary}" "./cmd/$1/"

    local specifier="_${OQTADRIVE_RELEASE}_$2_${arch}${suffix}"
    zip -j "../${binary}${specifier}.zip" "../${binary}"

    if [[ "$4" == "keep" ]]; then
    	mv "../${binary}" "../${binary}${specifier}"
    else
    	rm -f "../${binary}"
    fi
}

#
#
#
function download_oqtactl {

    local arch
    arch="$(get_architecture)" || return 1

    local os
    os="$(get_os "${arch}")" || return 1

    local marker="${os,,}_${arch}"
    local url

    if [[ -z "${BUILD_URL}" ]]; then # get form GitHub release page
        url="$(get_asset_url "${marker}.zip")"

    else # get from custom build page
        url="$(curl -fsSL "${BUILD_URL}" \
            | grep "${marker}\.zip" \
            | cut -d '"' -f 8)" && url="${BUILD_URL}/${url}"
    fi

    if [[ -z "${url}" || ! ${url} =~ ${marker} ]]; then
        echo -e \
            "\nNo download available for architecture '${arch}' on OS '${os}', in version '${VERSION}'.\n" >&2
        return 1
    fi

    echo "  from ${url}"
    curl -fsSL "${url}" -o oqtactl.zip
    unzip -o oqtactl.zip
    rm oqtactl.zip
    chmod +x oqtactl
    mv oqtactl "${OQTACTL}"
}

#
#
#
function download_firmware {

    local url

    if [[ -z "${BUILD_URL}" ]]; then # get form GitHub repo
        [[ -n "${VERSION}" ]] || VERSION="$(get_latest_release)"
        url="${BASE_URL}/${VERSION}/arduino/oqtadrive.ino"

    else # get from custom build page
        url="${BUILD_URL}/oqtadrive.ino"
    fi

    echo "  from ${url}"
    curl -fsSL  -o "${SKETCH}.org" "${url}"
}

#
# $1    filter
#
function get_asset_url {

    local path="releases/latest"
    [[ -z "${VERSION}" || "${VERSION}" == "latest" ]] \
        || path="releases/tags/${VERSION}"

    github_api_call "${path}" 2>/dev/null \
        | jq -r ".assets[]
            | select(.name | contains(\"$1\"))
            | .browser_download_url"
}

#
#
#
function get_latest_release {
    github_api_call "releases/latest" 2>/dev/null \
        | jq -r ".name"
}

#
# $1    path
#
function github_api_call {
    curl -fsSL -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/xelalexv/oqtadrive/$1"
}

#
# $...  actions, `install|remove|start|stop|enable|disable`
#
function manage_service {

    for a in "$@"; do

        case "${a}" in

            install)
                curl -fsSL "${BASE_URL}/${BRANCH}/hack/oqtadrive.service" \
                    | sed -E -e "s;^ExecStart=.*$;ExecStart=${OQTACTL} serve -d ${PORT};g" \
                          -e "s;^WorkingDirectory=.*$;WorkingDirectory=${ROOT};g" \
                          -e "s;^User=.*$;User=${USER};g" \
                    | sudo tee /etc/systemd/system/oqtadrive.service > /dev/null
                sudo systemctl daemon-reload
                ;;

            remove)
                sudo rm --force /etc/systemd/system/oqtadrive.service
                ;;

            start|stop|status|enable|disable)
                [[ ! -f /etc/systemd/system/oqtadrive.service ]] || {
                    echo "running daemon service action: ${a}"
                    sudo systemctl "${a}" "${OQTADRIVE_SERVICE}"
                }
                ;;

            *)
                echo -e "\nUnknown service command: ${a}\n" >&2
                return 1
                ;;
        esac

        sleep 1
    done
}

#
#
#
function patch_avrdude {

    local avrdude
    avrdude="$(find ~/.arduino15/ -type f -name avrdude.org)"
    [[ -z "${avrdude}" ]] || {
        echo "avrdude already patched!"
        return 1
    }

    avrdude="$(find ~/.arduino15/ -type f -name avrdude)"

    [[ -n "${avrdude}" ]] || {
        echo "avrdude not installed!"
        return 1
    }

    mv "${avrdude}" "${avrdude}.org"

    local dir
    dir="$(dirname "${avrdude}")"

    local autoreset="${dir}/autoreset"
    curl -fsSL -o "${autoreset}" "${BASE_URL}/${BRANCH}/hack/autoreset"
    sed -Ei "s/^pin[[:space:]]*=[[:space:]]*[0-9]+$/pin = ${RESET_PIN}/g" \
        "${autoreset}"
    chmod +x "${autoreset}"

    cat <<EOF > "${avrdude}"
#!/bin/sh
sudo strace -o "|${autoreset}" -eioctl "${dir}/avrdude.org" \$@
EOF
    chmod +x "${avrdude}"
}

#
#
#
function unpatch_avrdude {

    local avrdude
    avrdude="$(find ~/.arduino15/ -type f -name avrdude.org)"
    [[ -n "${avrdude}" ]] || {
        echo "avrdude not patched!"
        return 1
    }

    local dir
    dir="$(dirname "${avrdude}")"

    mv --force "${avrdude}" "${dir}/avrdude"
    rm --force "${dir}/autoreset"
}

#
# $...  prerequisites
#
function prereqs {

    local missing

    for p in "$@"; do
        type "${p}" >/dev/null 2>&1 || {
            missing+=" ${p}"
        }
    done

    [[ -z "${missing}" ]] || {
        echo -e "\nYou need to install these dependencies:${missing}\n"
        return 1
    }
}

#
#
#
function sanity {
    local arch
    arch="$(get_architecture)" || return 1
    get_os "${arch}" > /dev/null || return 1
}

#
#
#
function get_architecture {

    local arch
    arch="$(uname -m)"

    case ${arch} in
        x86_64)
            echo -n "amd64"
            ;;
        x86|i386|i686)
            echo -n "386"
            ;;
        armv5*|armv6*|armv7*)
            echo -n "arm"
            ;;
        aarch64)
            echo -n "arm64"
            ;;
        *)
            echo -e "\nUnsupported architecture: ${arch}\n" >&2
            return 1
            ;;
    esac
}

#
# $1    architecture
#
function get_os {

    local os
    os="$(uname -s)"

    local err=1

    case ${os} in

        Linux)
            case $1 in
                amd64|amd|arm|386)
                    err=0
                    ;;
            esac
            ;;

        Darwin)
            case $1 in
                amd64|arm64)
                    err=0
                    ;;
            esac
            echo -e "\nWARNING: This has not been tested on MacOS yet!\n" >&2
            ;;

        *)
            echo -e "\n${os} not supported.\n" >&2
            return 1
            ;;
    esac

    [[ ${err} -eq 0 ]] || \
        {
            echo -e "\nArchitecture $1 not supported on ${os}.\n" >&2
            return 1
        }

    echo -n "${os}"
}

#
#
#

cd "$(dirname "$0")" || exit 1
"$@"
