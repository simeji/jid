# jig
json interactive digger

It's very simple and pawerful tool.
You can drill down interactively by filtering query like [jq](https://stedolan.github.io/jq/)

## Demo

![demo-jig-main](https://github.com/simeji/jig/wiki/images/demo-jig-main-640.gif)

## Installation

### Just use "jig" command

Please download binary from below

https://github.com/simeji/jig/releases

### Build "jig" command by yourself

Jig require some packages.
Please go get below packages.

[bitly/go-simplejson](https://github.com/bitly/go-simplejson)
[nsf/termbox-go](https://github.com/nsf/termbox-go)
[pkg/erros](https://github.com/pkg/errors)
[stretchr/testify/assert](https://github.com/stretchr/testify/assert)

## Usage

### Quick start

#### simple json example

```
echo '{"aa":"2AA2","bb":{"aaa":[123,"cccc",[1,2]],"c":321}}'| jig
```

then, jig will be running.

You can dig JSON data incrementally.

You input `.bb.aaa[2]` and you can see below.

```
[Filter]> .bb.aaa[2]
[
  1,
  2
]
```

Then, you press Enter key and output `[1,2]` and exit.

#### simple json example2

This json is used by demo section.
```
echo '{"info":{"date":"2016-10-23","version":1.0},"users":[{"name":"simeji","uri":"https://github.com/simeji","id":1},{"name":"simeji2","uri":"https://example.com/simeji","id":2},{"name":"simeji3","uri":"https://example.com/simeji3","id":3}],"userCount":3}}'|jig
```

#### with curl

```
curl -s http://rdg.afilias.info/rdap/domain/example.info | jig
```

## Keymaps

|key|description|
|:-----------|:----------|
|`TAB` / `CTRL` + `I` |Show available items and choice them|
|`CTRL` + `W` |Delete from the cursor to the start of the word|
|`CTRL` + `F` / Right Arrow (:arrow_right:)|To the first character of the 'Filter'|
|`CTRL` + `B` / Left Arrow (:arrow_left:)|To the end of the 'Filter'|
|`CTRL` + `A`|To the first character of the 'Filter'|
|`CTRL` + `E`|To the end of the 'Filter'|
|`CTRL` + `J`|Scroll json buffer 1 line downwards|
|`CTRL` + `K`|Scroll json buffer 1 line upwards|
|`CTRL` + `L`|Change view mode whole json or keys (only object)|

### Option

-p : Pretty print (output json)

-q : Print query (for jq)
