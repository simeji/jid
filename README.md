# jid

[![Circle CI](https://circleci.com/gh/simeji/jid/tree/master.svg?style=shield)](https://circleci.com/gh/simeji/jid/tree/master)

Json Incremental Digger

It's a very simple tool.  
You can drill down JSON interactively by using filtering queries like [jq](https://stedolan.github.io/jq/).

**Suggestion** and **Auto completion** of this tool will provide you a very comfortable JSON drill down.

## Demo

![demo-jid-main](https://github.com/simeji/jid/wiki/images/demo-jid-main-640-colorize.gif)

## Installation

* [With HomeBrew (for macOS)](#with-homebrew-for-macos)  
* [With MacPorts (for macOS)](#with-macports-for-macos)  
* [With pkg (for FreeBSD)](#with-pkg-for-freebsd)
* [With scoop (for Windows)](#with-scoop-for-windows)
* [Other package management system](#other-package-management-systems)
* [Simply use "jid" command](#simply-use-jid-command)  
* [Build](#build)  

### With HomeBrew (for macOS)

```
brew install jid
```

### With MacPorts (for macOS)

```
sudo port install jid
```

### With pkg (for FreeBSD)

```
pkg install jid
```

### With scoop (for Windows)

```
scoop install jid
```

### Other package management systems

Jid can install by package management systems of below OS.

[![Packaging status](https://repology.org/badge/vertical-allrepos/jid.svg)](https://repology.org/metapackage/jid/versions)


### Simply use "jid" command

If you simply want to use `jid` command, please download binary from below.

https://github.com/simeji/jid/releases

## Build

```
go install github.com/simeji/jid/cmd/jid@latest
```

## Usage

### Quick start

* [simple json example](#simple-json-example)  
* [simple json example2](#simple-json-example2)  
* [with initial query](#with-initial-query)  
* [with curl](#with-curl)  

#### simple json example

Please execute the below command.

```
echo '{"aa":"2AA2","bb":{"aaa":[123,"cccc",[1,2]],"c":321}}'| jid
```

then, jid will be running.

You can dig JSON data incrementally.

When you enter `.bb.aaa[2]`, you will see the following.

```
[Filter]> .bb.aaa[2]
[
  1,
  2
]
```

Then, you press Enter key and output `[1,2]` and exit.

#### simple json example2

This json is used by [demo section](https://github.com/simeji/jid#demo).
```
echo '{"info":{"date":"2016-10-23","version":1.0},"users":[{"name":"simeji","uri":"https://github.com/simeji","id":1},{"name":"simeji2","uri":"https://example.com/simeji","id":2},{"name":"simeji3","uri":"https://example.com/simeji3","id":3}],"userCount":3}}'|jid
```

#### With a initial query

First argument of `jid` is initial query.
(Use JSON same as [Demo](#demo))

![demo-jid-with-query](https://github.com/simeji/jid/wiki/images/demo-jid-with-query-640.gif)

#### with curl

Sample for using [RDAP](https://datatracker.ietf.org/wg/weirds/documents/) data.

```
curl -s http://rdg.afilias.info/rdap/domain/example.info | jid
```

#### Load JSON from a file

```
jid < file.json
```

## Keymaps

|key|description|
|:-----------|:----------|
|`TAB` / `CTRL` + `I` |Show available items and choice them|
|`CTRL` + `W` |Delete from the cursor to the start of the word|
|`CTRL` + `U` |Delete whole query|
|`CTRL` + `F` / Right Arrow (:arrow_right:)|Move cursor a character to the right|
|`CTRL` + `B` / Left Arrow (:arrow_left:)|Move cursor a character to the left|
|`CTRL` + `A`|To the first character of the 'Filter'|
|`CTRL` + `E`|To the end of the 'Filter'|
|`CTRL` + `J`|Scroll json buffer 1 line downwards|
|`CTRL` + `K`|Scroll json buffer 1 line upwards|
|`CTRL` + `G`|Scroll json buffer to bottom|
|`CTRL` + `T`|Scroll json buffer to top|
|`CTRL` + `N`|Scroll json buffer 'Page Down'|
|`CTRL` + `P`|Scroll json buffer 'Page Up'|
|`CTRL` + `L`|Change view mode whole json or keys (only object)|
|`ESC`|Hide a candidate box|

### Option

|option|description|
|:-----------|:----------|
|First argument ($1) | Initial query|
|-h | print a help|
|-help | print a help|
|-version | print the version and exit|
|-q | Output query mode (for jq)|
|-M | monochrome output mode|
