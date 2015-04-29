package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/russross/blackfriday"
)

type HtmlMore struct {
	*blackfriday.Html
}

func HtmlMoreRenderer(flags int, title string, css string) blackfriday.Renderer {
	r := blackfriday.HtmlRendererWithParameters(flags, title, css, blackfriday.HtmlRendererParameters{})
	return &HtmlMore{
		r.(*blackfriday.Html),
	}
}

func (options *HtmlMore) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	doubleSpace(out)

	langinfo := strings.Fields(lang)
	if len(langinfo) < 1 {
		return
	}
	switch langinfo[0] {
	case "uml", "plantuml":
		//如何管理图片目录？
		//如何管理图片反复生成？
		//如何保证文件名稳定友好？
		//如何保证所有转换编程非阻塞？
		out.WriteString("binggo\n")
		wd, _ := os.Getwd()
		tmpdir, _ := ioutil.TempDir(os.TempDir(), "hugo")
		tmpuml := path.Join(tmpdir, "umlname.uml")
		// tmpuml := path.Join(wd, "umlname.uml")
		tmpimgdir := path.Join(wd, "uml")
		fmt.Println(tmpuml)
		f, err := os.Create(tmpuml)
		if err == nil {
			f.Write(text)
			f.Close()
		}

		outinfo, _ := exec.Command("plantuml", "-nbthread auto", "-tsvg", tmpuml, "-o", tmpimgdir).Output()
		fmt.Println(outinfo)

		// out.WriteString()
		attrEscape(out, text)
	default:
		count := 0
		for _, elt := range strings.Fields(lang) {
			if elt[0] == '.' {
				elt = elt[1:]
			}
			if len(elt) == 0 {
				continue
			}
			if count == 0 {
				out.WriteString("<pre><code class=\"language-")
			} else {
				out.WriteByte(' ')
			}
			attrEscape(out, []byte(elt))
			count++
		}

		if count == 0 {
			out.WriteString("<pre><code>")
		} else {
			out.WriteString("\">")
		}

		attrEscape(out, text)
		out.WriteString("</code></pre>\n")
	}
}

//copy code from blackfriday
func doubleSpace(out *bytes.Buffer) {
	if out.Len() > 0 {
		out.WriteByte('\n')
	}
}

// Using if statements is a bit faster than a switch statement. As the compiler
// improves, this should be unnecessary this is only worthwhile because
// attrEscape is the single largest CPU user in normal use.
// Also tried using map, but that gave a ~3x slowdown.
func escapeSingleChar(char byte) (string, bool) {
	if char == '"' {
		return "&quot;", true
	}
	if char == '&' {
		return "&amp;", true
	}
	if char == '<' {
		return "&lt;", true
	}
	if char == '>' {
		return "&gt;", true
	}
	return "", false
}

func attrEscape(out *bytes.Buffer, src []byte) {
	org := 0
	for i, ch := range src {
		if entity, ok := escapeSingleChar(ch); ok {
			if i > org {
				// copy all the normal characters since the last escape
				out.Write(src[org:i])
			}
			org = i + 1
			out.WriteString(entity)
		}
	}
	if org < len(src) {
		out.Write(src[org:])
	}
}
