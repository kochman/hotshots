# Hotshots 🔥
[![Build Status](https://travis-ci.org/kochman/hotshots.svg?branch=master)](https://travis-ci.org/kochman/hotshots)&nbsp;[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkochman%2Fhotshots.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkochman%2Fhotshots?ref=badge_shield)&nbsp;[![codecov](https://codecov.io/gh/kochman/hotshots/branch/master/graph/badge.svg)](https://codecov.io/gh/kochman/hotshots)

Hotshots automatically uploads photos from remote cameras to a server, allowing news organization editors or event photographers to quickly curate and share content in real-time through social media.

## Development

You will need to install libgphoto2. On macOS with Homebrew, this can be done with `brew install libgphoto2`.

```
go get github.com/kochman/hotshots
go get -u github.com/kardianos/govendor
cd $GOPATH/src/github.com/kochman/hotshots
govendor sync
go build -i
./hotshots -h
```


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkochman%2Fhotshots.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkochman%2Fhotshots?ref=badge_large)
