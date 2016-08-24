package jig

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"io"
	"strings"
)

const (
	PROMPT              = "[jig]>> "
	DefaultY     int    = 1
	FilterPrompt string = "[Filter]> "
)

type Engine struct {
	manager *JsonManager
	jq      bool
	pretty  bool
	query   *Query
}

func NewEngine(s io.Reader, q bool, p bool) *Engine {
	j, err := NewJsonManager(s)
	if err != nil {
		return &Engine{}
	}
	e := &Engine{
		manager: j,
		jq:      q,
		pretty:  p,
		query:   NewQuery([]rune("")),
	}
	return e
}

func (e Engine) Run() int {

	if !e.render() {
		return 2
	}
	if e.jq {
		fmt.Printf("%s", e.query.StringGet())
	} else if e.pretty {
		s, _, err := e.manager.GetPretty(e.query)
		if err != nil {
			return 1
		}
		fmt.Printf("%s", s)
	} else {
		s, _, err := e.manager.Get(e.query)
		if err != nil {
			return 1
		}
		fmt.Printf("%s", s)
	}
	return 0
}

// fix:me
func (e *Engine) render() bool {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	contents, _, _ := e.manager.GetPretty(e.query)
	keymode := false

	var complete string

	for {
		//var flgFilter bool
		contents, complete, _ = e.manager.GetPretty(e.query)
		if keymode {
			//ckeys := []string{}
			//kws := e.query.StringGetKeywords()
			//if lkw := kws[len(kws)-1]; lkw != "" && !flgFilter {
			//for k, _ := range e.getFilteredCurrentKeys(e.json, lkw) {
			//ckeys = append(ckeys, e.currentKeys[k])
			//}
			//sort.Strings(ckeys)
			//contents = ckeys
			//} else {
			//contents = e.currentKeys
			//}
		} else {
			//contents, complete, _ = e.manager.GetPretty(e.query)
		}
		e.draw(e.query.StringGet(), complete, strings.Split(contents, "\n"))
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return false
			case termbox.KeyCtrlK:
				keymode = !keymode
			case termbox.KeySpace:
				_ = e.query.StringAdd(" ")
			case termbox.KeyCtrlW:
				//delete whole word to period
				_, _ = e.query.PopKeyword()
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				_ = e.query.Delete(1)
			case termbox.KeyTab:
				_ = e.query.StringAdd(complete)
			case termbox.KeyEnter:
				return true
			case 0:
				_ = e.query.StringAdd(string(ev.Ch))
			default:
			}
		case termbox.EventError:
			panic(ev.Err)
			break
		default:
		}
	}
}

//
//func (e *Engine) suggest() bool {
//	s := e.query.StringGet()
//	if arr, _ := e.json.Array(); arr != nil {
//		if l := len(s); l < 1 {
//			*complete = []rune("")
//			return false
//		}
//		le := s[len(s)-1:]
//		if le == "." {
//			*complete = []rune("")
//			return false
//		}
//		var rs string
//		ds := regexp.MustCompile("\\[([0-9]*)?\\]?$").FindString(s)
//		if len(arr) > 1 {
//			if ds == "" {
//				rs = "["
//			} else if le != "]" {
//				rs = "]"
//			}
//		} else {
//			rs = "[0]"
//		}
//		cs := strings.Replace(rs, ds, "", -1)
//		*complete = []rune(cs)
//		return true
//	}
//	kws := strings.Split(s, ".")
//	lki := len(kws) - 1
//	if lki == 0 {
//		return false
//	}
//	lkw, tkws := kws[lki], kws[:lki]
//
//	re, err := regexp.Compile("(?i)^" + lkw)
//	if err != nil {
//		return false
//	}
//	m := e.getFilteredCurrentKeys(e.json, lkw)
//
//	if len(m) == 1 {
//		for k, v := range m {
//			kw := re.ReplaceAllString(e.currentKeys[k], "")
//			*complete = []rune(kw)
//			s = strings.Join(tkws, ".") + "." + v
//		}
//		_ = e.query.StringSet(s)
//		return true
//	} else {
//		var sw []rune
//		cnt := 0
//		for k, _ := range m {
//			tsw := []rune{}
//			v := []rune(e.currentKeys[k])
//			if cnt == 0 {
//				sw = v
//				cnt = cnt + 1
//				continue
//			}
//			swl := len(sw) - 1
//			for i, s := range v {
//				if i > swl {
//					break
//				}
//				if sw[i] != s {
//					break
//				}
//				tsw = append(tsw, s)
//			}
//			sw = tsw
//			cnt = cnt + 1
//		}
//		if len(sw) >= 0 {
//			kw := re.ReplaceAllString(string(sw), "")
//			*complete = []rune(kw)
//			s = strings.Join(tkws, ".") + "." + lkw
//		}
//		_ = e.query.StringSet(s)
//		return true
//	}
//	*complete = []rune("")
//	return false
//}

// func (e *Engine) getFilteredCurrentKeys(json *simplejson.Json, kw string) map[int]string {
// 	m := map[int]string{}
//
// 	re, err := regexp.Compile("(?i)^" + kw)
// 	if err != nil {
// 		return m
// 	}
//
// 	currentKeys := e.getCurrentKeys(json)
// 	for i, k := range currentKeys {
// 		if str := re.FindString(k); str != "" {
// 			m[i] = str
// 		}
// 	}
// 	return m
// }
//
//func (e *Engine) prettyContents() []string {
//	s, _ := e.json.EncodePretty()
//	return strings.Split(string(s), "\n")
//}

//func (e *Engine) filterJson(json *simplejson.Json, q string) (*simplejson.Json, bool) {
//	if len(q) < 1 {
//		return json, false
//	}
//	keywords := strings.Split(q, ".")
//
//	// check start "."
//	if keywords[0] != "" {
//		return &simplejson.Json{}, false
//	}
//
//	keywords = keywords[1:]
//
//	re := regexp.MustCompile("\\[[0-9]*\\]")
//	delre := regexp.MustCompile("\\[([0-9]+)?")
//
//	lastIdx := len(keywords) - 1
//
//	flgMatchLastKw := false
//
//	//eachFlg := false
//	for ki, keyword := range keywords {
//		if len(keyword) == 0 {
//			if ki != lastIdx {
//				json = &simplejson.Json{}
//			}
//			break
//		}
//		// abc[0]
//		if keyword[:1] == "[" {
//			matchIndexes := re.FindAllStringIndex(keyword, -1)
//			lmi := len(matchIndexes) - 1
//			for idx, m := range matchIndexes {
//				i, _ := strconv.Atoi(keyword[m[0]+1 : m[1]-1])
//				if idx == lmi && m[1]-m[0] == 2 {
//					//eachFlg = true
//				} else if tj := json.GetIndex(i); !isEmptyJson(tj) {
//					json = tj
//				}
//			}
//		} else if keyword[len(keyword)-1:] == "]" {
//			matchIndexes := re.FindAllStringIndex(keyword, -1)
//			kw := re.ReplaceAllString(keyword, "")
//
//			tj := json.Get(kw)
//			if ki != lastIdx {
//				json = tj
//			} else if !isEmptyJson(tj) {
//				json = tj
//				flgMatchLastKw = true
//			}
//			lmi := len(matchIndexes) - 1
//			for idx, m := range matchIndexes {
//				i, _ := strconv.Atoi(keyword[m[0]+1 : m[1]-1])
//				if idx == lmi && m[1]-m[0] == 2 {
//					//eachFlg = true
//				} else if tj := json.GetIndex(i); !isEmptyJson(tj) {
//					json = tj
//					flgMatchLastKw = true
//				}
//			}
//		} else {
//			kw := delre.ReplaceAllString(keyword, "")
//			tj := json.Get(kw)
//			if ki != lastIdx {
//				json = tj
//			} else if len(e.getFilteredCurrentKeys(json, kw)) < 1 {
//				json = tj
//			} else if !isEmptyJson(tj) {
//				json = tj
//				flgMatchLastKw = true
//			}
//		}
//	}
//	return json, flgMatchLastKw
//}

//func (e *Engine) getCurrentKeys(json *simplejson.Json) []string {
//
//	keys := []string{}
//	m, err := json.Map()
//
//	if err != nil {
//		return keys
//	}
//	for k := range m {
//		keys = append(keys, k)
//	}
//	sort.Strings(keys)
//	return keys
//}

func (e *Engine) draw(query string, complete string, rows []string) {

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	fs := FilterPrompt + query
	cs := complete
	drawln(0, 0, fs+cs, []([]int){[]int{len(fs), len(fs) + len(cs)}})
	termbox.SetCursor(len(fs), 0)

	for idx, row := range rows {
		drawln(0, idx+DefaultY, row, nil)
	}

	termbox.Flush()
}

func drawln(x int, y int, str string, matches [][]int) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault

	var c termbox.Attribute
	for i, s := range str {
		c = color
		for _, match := range matches {
			if i >= match[0] && i < match[1] {
				c = termbox.ColorGreen
			}
		}
		termbox.SetCell(x+i, y, s, c, backgroundColor)
	}
}
