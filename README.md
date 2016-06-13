# jig
json incremental digger

## Installation

### Just use "jig" command

Please download binary from below

https://github.com/simeji/jig/releases

### Build "jig" command by yourself

jig required 'simplejson'

```
go get github.com/bitly/go-simplejson
```

## Usage

Quick start

```
echo '{"aa":"2AA2","bb":{"aaa":[123,"cccc",[1,2]],"c":321}}'| jig
```

then, jig will run.

You can dig JSON data incrementally.

You input `.bb.aaa[2]` and you can see below

## Keymaps

`Ctrl+K` : Change, View mode whole json or keys (only object)

```
[Filter]> .bb.aaa[2]
[
  1,
  2
]
```

Then, you press Enter key and output `[1,2]` and exit.


### Option

-p : Pretty print (output json)

-q : Print query (for jq)
