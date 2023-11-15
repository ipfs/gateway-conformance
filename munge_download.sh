#!/usr/bin/env bash
# This script downloads the latest gateway-conformance.json artifacts from the
# given repos and saves them to the output folder.
# it requires GH_TOKEN to be set to a valid GitHub token with repo access.

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 output_folder repo1 [repo2 ...]"
    exit 1
fi

# if GH_TOKEN is empty, show a message to the user
if [ -z "$GH_TOKEN" ]; then
    echo "::warning::GH_TOKEN is required to download the build artifact."
    echo "You can also download the artifacts file from the latest CI run:"
    echo "https://github.com/ipfs/gateway-conformance/actions/workflows/deploy-pages.yml"

    # if the output folder contains some json files, we'll assume that the user has already downloaded them and move on.
    if [ -z "$(ls -A $1/*.json 2>/dev/null)" ]; then
        echo "::error::GH_TOKEN is required to download the build artifact."
        exit 1
    fi
fi

if ! command -v jq &> /dev/null || ! command -v unzip &> /dev/null; then
    echo "::error::Required command(s) 'jq' and/or 'unzip' not found."
    exit 1
fi

output_folder=$1
mkdir -p "$output_folder"

shift
REPOS=("$@")

GH_TOKEN="${GH_TOKEN:-}"
GH_API_BASE_URL="https://api.github.com"

gh_api_request() {
    local endpoint="${@: -1}"

    if [[ $endpoint =~ ^https?:// ]]; then
        local url="$endpoint"
    else
        local url="$GH_API_BASE_URL$endpoint"
    fi

    curl -H "Authorization: token $GH_TOKEN" "${@:1:$#-1}" "$url"
}

for repo in "${REPOS[@]}"; do
    default_branch=$(gh_api_request "/repos/$repo" | jq -r '.default_branch')
    [ -z "$default_branch" ] && echo "::warning::Failed to fetch default branch for $repo" && continue

    run_id=$(gh_api_request "/repos/$repo/actions/workflows/gateway-conformance.yml/runs?branch=$default_branch&status=success&per_page=1" | jq -r '.workflow_runs[0].id')
    [ -z "$run_id" ] && echo "::warning::Failed to fetch workflow run ID for $repo" && continue

    artifact_url=$(gh_api_request "/repos/$repo/actions/runs/$run_id/artifacts" | jq -r '.artifacts[] | select(.name=="gateway-conformance.json") | .archive_download_url')
    [ -z "$artifact_url" ] && echo "::warning::Failed to fetch artifact URL for $repo" && continue

    temp_dir=$(mktemp -d)
    zip_file="$temp_dir/${repo##*/}-artifact.zip"
    gh_api_request -L -o "${zip_file}" "$artifact_url" || echo "::warning::Failed to download artifact for $repo"
    unzip -j "$zip_file" "output.json" -d "$temp_dir" && mv "$temp_dir/output.json" "$output_folder/${repo##*/}.json" || echo "::warning::Failed to extract output.json for $repo" && continue

    rm -rf "$temp_dir"
done
