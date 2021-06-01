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

    command -v gawk > /dev/null || echo "Note: proper help display requires gawk!"

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
        | gawk '/^##$/{f=1; printf "-%s", $0; next} /^[^#].*$/{f=0} /^$/{f=0} f' \
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

    echo -e "\nbuilding ${binary} for $2/${arch}"

    # shellcheck disable=SC2086
    docker run --rm --user "$(id -u):$(id -g)" \
        -v "${ROOT}/${BINARIES}:/go/bin" ${CACHE_VOLS} \
        -v "${ROOT}:/go/src/${REPO}" -w "/go/src/${REPO}" \
        -e CGO_ENABLED=0 -e GOOS="$2" -e GOARCH="${arch}" \
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

cd "$(dirname "$0")" || exit 1
"$@"
