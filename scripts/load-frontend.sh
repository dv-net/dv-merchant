#!/usr/bin/env bash
FRONTEND_PROJECT_NAME=dv-frontend
GITHUB_OWNER=dv-net
DV_FRONTEND_TAG=$(curl -s https://api.github.com/repos/${GITHUB_OWNER}/${FRONTEND_PROJECT_NAME}/releases/latest | jq -r '.tag_name')
DV_FRONTEND_ARCHIVE_FILENAME="${FRONTEND_PROJECT_NAME}.${DV_FRONTEND_TAG}.tar.gz"

echo "Downloading frontend ${DV_FRONTEND_TAG} ..."
curl -sSL -o ${DV_FRONTEND_ARCHIVE_FILENAME} \
"https://github.com/${GITHUB_OWNER}/${FRONTEND_PROJECT_NAME}/releases/download/${DV_FRONTEND_TAG}/${DV_FRONTEND_ARCHIVE_FILENAME}"

mkdir -p frontend/dist
echo "Extracting archive ${DV_FRONTEND_ARCHIVE_FILENAME} to frontend/dist ..."
tar -xf ${DV_FRONTEND_ARCHIVE_FILENAME} -C frontend/dist
rm ${DV_FRONTEND_ARCHIVE_FILENAME}