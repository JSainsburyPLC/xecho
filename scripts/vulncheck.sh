#!/usr/bin/env bash
# Runs govulncheck and fails on reachable vulnerabilities unless they are explicitly accepted.
#
# VULNCHECK_ACCEPTED:    long-term accepted risks. Never fail while reported.
# VULNCHECK_UNTIL_FIXED: temporarily accepted risks. Fail as soon as a fixed version is published.
# Any listed ID that govulncheck no longer reports also fails, so the lists stay accurate.
set -euo pipefail

gotool=${GOTOOL:-go tool -modfile=tools/go.mod}
accepted=" ${VULNCHECK_ACCEPTED:-} "
until_fixed=" ${VULNCHECK_UNTIL_FIXED:-} "

annotate() { # level message
	if [[ -n ${GITHUB_ACTIONS:-} ]]; then echo "::$1::$2"; else echo "$1: $2"; fi
}

# One line per reachable advisory: "<id> <fixed version or ->", preferring a known fixed version.
findings=$($gotool govulncheck -format json ./... |
	jq -r 'select(.finding.trace[0].function != null) | "\(.finding.osv) \(.finding.fixed_version // "-")"' |
	sort -k1,1 -k2,2r | awk '!seen[$1]++')

status=0
reported=" "
while read -r id fixed; do
	[[ -z $id ]] && continue
	reported+="$id "
	url="https://pkg.go.dev/vuln/$id"
	if [[ $accepted == *" $id "* ]]; then
		echo "accepted: $id ($url)"
	elif [[ $until_fixed == *" $id "* ]]; then
		if [[ $fixed == "-" ]]; then
			annotate warning "accepted until fixed, no fix released yet: $id ($url)"
		else
			annotate error "fix released in $fixed for $id: upgrade and remove it from VULNCHECK_UNTIL_FIXED ($url)"
			status=1
		fi
	else
		[[ $fixed == "-" ]] && fixed="N/A"
		annotate error "vulnerable: $id, fixed in $fixed ($url)"
		status=1
	fi
done <<<"$findings"

for id in ${VULNCHECK_ACCEPTED:-} ${VULNCHECK_UNTIL_FIXED:-}; do
	if [[ $reported != *" $id "* ]]; then
		annotate error "$id is listed in the Makefile but no longer reported: remove it"
		status=1
	fi
done

if [[ $status -ne 0 ]]; then echo "Run '$gotool govulncheck ./...' for traces."; fi
exit $status
