#!/bin/bash
if [ -z ${TAG_NAME+x} ]; then
if [ -z ${BRANCH_NAME+x} ]; then
BRANCH_NAME=$(echo $(git branch --show-current) || \
	echo $(git name-rev --name-only HEAD))
fi
GIT_VERSION=$(echo ${BRANCH_NAME} | grep -q 'release/' \
	&& echo ${BRANCH_NAME} | sed -e 's|release/|v|' -e 's/$/-RC/' || \
	echo $(git describe --always --tags --dirty 2>/dev/null) || echo v0)
else
GIT_VERSION=${TAG_NAME}
fi

if [ -z ${VERSION+x} ]; then
VERSION=$(echo ${GIT_VERSION} | sed -e 's|^origin/||')
fi

if [ -z $1 ]; then
echo "${VERSION}"
else
echo ${VERSION} | sed -e 's/^v//'
fi
