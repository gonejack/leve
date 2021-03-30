# LEVE

Command line tool to save RSS articles as .eml files.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gonejack/leve)
![Build](https://github.com/gonejack/leve/actions/workflows/go.yml/badge.svg)
[![GitHub license](https://img.shields.io/github/license/gonejack/leve.svg?color=red)](LICENSE)

### Install
```shell
> go get github.com/gonejack/leve
```

### Usage

```shell
> leve [-c ~/.leve] [urls...]
```

### Config
To avoid passing urls everytime, put your feed urls inside `~/.leve/feeds.txt`
```shell
> leve -v
> vi ~/.leve/feeds.txt
```

### Files
- Images would be saved into dir `~/.leve/cache`
- Article records would be saved into `~/.leve/records.txt`
