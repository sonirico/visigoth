#!/usr/bin/env bash
set -e

LATEST=false

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd "$DIR"

if [ ! -z $1 ]; then
    ACTION=$1
fi

if [[ ! -z $2 && $2 == "latest" ]]; then
    LATEST=true
fi

echo "## Build image with version info... "
docker build -t $PROJECT:$VERSION -f Dockerfile .

if [ "$LATEST" == true ]; then
    echo "## Tag additional 'latest'... "
    docker tag $PROJECT:$VERSION $PROJECT:latest
fi

# If its dev mode, only build for ourself
if [ "$ACTION" == "release" ]; then
    echo "## RELEASING IS CURRENTLY NOT FULLY IMPLEMENTED"

    # if [[ -z $DOCKER_USERNAME || -z $DOCKER_PASSWORD ]]; then
    #     echo "## No docker username and password found please provide for login... "
    # else 
    #     echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
    # fi
    
    # echo "## Push version image to registry... "
    # docker push $PROJECT:$VERSION

    # if [ $LATEST ]; then
    #     echo "## Push latest image to registry... "
    #     docker push $PROJECT:latest
    # fi
fi
