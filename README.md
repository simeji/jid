# jig
json incremental digger

## Installation

jig required 'simplejson'

```
go get github.com/bitly/go-simplejson
```

## Usage

Quick start

```
echo '{"aa":"2AA2","bb":{"aaa":[123,"cccc",[1,2]],"c":321}}'| jig
```

You can dig JSON data incrementally.

### Option

-p : Pretty print (output json)
-q : Print query (for jq)
