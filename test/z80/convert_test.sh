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

OQTACTL=../../_build/bin/oqtactl

#
# $1	path to Z80 snapshot file
# $2 	test output folder
#
function convert_test {

	local out
	out="$2/$(basename "$1")"
	out="${out%.z80}"

	echo -e "\nC version:"
	local out_c="${out}.c.mdr"

	start=$(date +%s)
	./z80lite "$1"
	end=$(date +%s)
	runtime_c=$((runtime_c + end-start))
	echo
	mv "${1%.z80}.mdr" "${out_c}"

	echo -e "\nGo version:"
	local out_go="${out}.go.mdr"

	start=$(date +%s)
	if ! "${OQTACTL}" load -i "$1" -n Z80onMDR; then
		echo -e "\nLOAD CRASHED"
		return 2
	fi
	if ! "${OQTACTL}" save -o "${out_go}"; then
		echo -e "\nSAVE CRASHED"
		return 3
	fi

	end=$(date +%s)
	runtime_go=$((runtime_go + end-start))

	local ret=0

	if ! cmp "${out_go}" "${out_c}"; then
		echo -e "\n\nFAILED:"
		"${OQTACTL}" dump -i "${out_go}" > "${out_go}.dump"
		"${OQTACTL}" dump -i "${out_c}" > "${out_c}.dump"
		#meld "${out_go}.dump" "${out_c}.dump" &
		echo
		ret=1
	fi

	rm -f "${out_go}" "${out_c}"
	echo -e "\nPASSED"

	return ${ret}
}

#
#
#

if [[ $# -ne 1 ]]; then
	echo '
synopsis: convert_test.sh {Z80 snapshot folder}

This runs conversions of the Z80 files contained in the provided folder with the
original C version and with OqtadDrive, and compares the results. This is not a
strict test, but rather a plausibility check.

If there is a difference, this may not necessarily be a failure. The compression
done for a particular snapshot file by the two differing implementations may
produce different results. Overall, most of the time (for more than 90% of the
snapshot files) identical results should be produced. Those that fail could be
loaded into a Spectrum to check on validity. Hex dumps of all failures are also
placed in the test output folder (out) for easy diffing.

To run the test, the daemon needs to be started separately.
'
	exit 1
fi

cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1

# compile Z80onMDR Lite
if [[ ! -f z80lite ]]; then
	echo "building Z80onMDR Lite..."
	gcc -Os z80lite.c -o z80lite
fi

# prep
test_folder="out"
mkdir -p "${test_folder}"
rm -rf "${test_folder:?}"/*

echo -e "\nunpacking snapshots..."
unzip -o -j "$1/*.zip" -d "$1"
rm -f "$1"/*.zip

# counters
runtime_go=0
runtime_c=0
success=0
failure=0
max=100

echo -e "\ntesting snapshot conversion..."

for f in "$1"/*.z80; do
	echo -n "$(basename "${f}") "
	if convert_test "${f}" "${test_folder}" > /dev/null 2>&1; then
		(( success++ ))
		echo " OK"
	else
		(( failure++ ))
		echo " FAIL"
	fi
	[[ $(( success + failure )) -lt ${max} ]] || break
done

echo -e "\nsuccess: ${success}, failure: ${failure}"
echo -e "\nGo: ${runtime_go}, C: ${runtime_c}"
