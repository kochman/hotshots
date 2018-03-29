# Hotshots 🔥
[![Build Status](https://travis-ci.org/kochman/hotshots.svg?branch=master)](https://travis-ci.org/kochman/hotshots)&nbsp;[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkochman%2Fhotshots.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkochman%2Fhotshots?ref=badge_shield)&nbsp;[![codecov](https://codecov.io/gh/kochman/hotshots/branch/master/graph/badge.svg)](https://codecov.io/gh/kochman/hotshots)

Hotshots automatically uploads photos from remote cameras to a server, allowing news organization editors or event photographers to quickly curate and share content in real-time through social media.

## Development

```
go get github.com/kochman/hotshots
go get -u github.com/kardianos/govendor
cd $GOPATH/src/github.com/kochman/hotshots
govendor sync
go build -i
./hotshots -h
```

## Deployment

Hotshots has a Docker image for easy deployment of the server. It expects a volume to be mounted into the container at `/var/hotshots` so that it can persist its data.

An example command that will run the Docker image and have it listen on the host's port 80:
```
docker run -p 80:8000 -v hotshots:/var/hotshots --name hotshots -d sidney/hotshots
```

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkochman%2Fhotshots.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkochman%2Fhotshots?ref=badge_large)
