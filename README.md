# LEVE

Command line tool to save RSS articles as .eml files.

### Install
```shell
> go get github.com/gonejack/leve
```

### Config
Put your feed urls inside `~/.leve/feeds.txt`.
```shell
> leve -v
> vi ~/.leve/feeds.txt
```

### Usage

```shell
> leve [-c ~/.leve]
```

### Files
- Images would be saved into dir `~/.leve/cache`
- Article records would be saved into `~/.leve/records.txt`
