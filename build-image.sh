docker build -f Dockerfile -t core-deb .
docker tag core-deb kalidux/core-deb
docker push kalidux/core-deb
