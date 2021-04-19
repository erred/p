# p

basic paste service,
hosted on [p.seankhliao.com](https://p.seankhliao.com/)

[![License](https://img.shields.io/github/license/seankhliao/p.svg?style=flat-square)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/go.seankhliao.com/p.svg)](https://pkg.go.dev/go.seankhliao.com/p)
![Version](https://img.shields.io/github/v/tag/seankhliao/p?sort=semver&style=flat-square)

![screenshot](https://user-images.githubusercontent.com/11343221/115290683-211ff780-a154-11eb-9420-9981f77cb31f.png)

## build / install

```sh
go install go.seankhliao.com/p@latest
```

## run

```sh
p
```

## config

```sh
Usage of p:
  -addr string
            address to listen on (default "127.0.0.1:28002")
  -data string
            path to data dir (default "data")
```

## todo

- [ ] Docker packaging
- [ ] Arch packaging
- [ ] systemd service
- [ ] Content Security Policy
- [ ] Metrics / stats / logs
