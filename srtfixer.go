package main

import (
	"fmt"
	"os"
	"io"
	"log"
	"bufio"
	"bytes"
	"strings"
	"path/filepath"
	"errors"
	"unicode/utf8"
	"regexp"
	
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	
	var fileName = filepath.Clean(os.Args[1])
	
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fatal("File Not Found")
	}
	
	if strings.ToLower(filepath.Ext(fileName)) != ".srt" {
		fatal("File Not Supported")
	}
	
	var newName = fileName[:len(fileName) - len(filepath.Ext(fileName))] + "_fixed.srt"
	
	file, err := os.Open(fileName)
	cheker(err)
	defer file.Close()
	
	out, err := os.Create(newName)
	cheker(err)
	defer out.Close()
	
	var scanner = bufio.NewScanner(file)
	
	for scanner.Scan() {
		buf := scanner.Bytes()
		if !isUTF8(buf) {
			decodeWindows1256(&buf)		
		}
		
		str := string(buf);
		
		if isTextLine(buf) {
			str = fixText(str)
		}
		
		out.WriteString(fmt.Sprintln(str))
		out.Sync()
	}
}

func usage() {
	fatal(fmt.Sprintf("Usage:\n%s subtitle.srt\n", filepath.Base(os.Args[0])))
}

func fatal(err string) {
	log.Fatalln(err)
}

func cheker(err error) {
	if err != nil {
		fatal(err.Error())
	}
}

func isTextLine(str []byte) bool {
	var reTime = regexp.MustCompile(`^\d\d:\d\d:\d\d,\d\d\d\s-->\s\d\d:\d\d:\d\d,\d\d\d$`)
	var isNum = regexp.MustCompile(`^\d+$`)
	
	if strings.Trim(string(str), " ") != "" && !isNum.Match(str) && !isNum.Match(str[:len(str)-1]) && !reTime.Match(str) && !reTime.Match(str[:len(str)-1]) {
		return true
	}
	
	return false
}

func isUTF8(str []byte) bool {
	if utf8.Valid(str) {
		return true
	}
	return false
}

func decodeWindows1256(str *[]byte) error {
	reader := bytes.NewReader(*str)
	transformer := transform.NewReader(reader, charmap.Windows1256.NewDecoder())
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, transformer); err == nil {
		*str = buf.Bytes()
		return nil
	}
	return errors.New("Convert Faild")
}

func fixText(str string) string {
	// clean rtl character
	str = strings.Replace(str, "\u202b", "", -1)
	
	// remove italic style which persian languge does not support
	str = strings.Replace(str, "<i>", "", -1)
	str = strings.Replace(str, "</i>", "", -1)
	
	// remove some arabic characters
	str = strings.Replace(str, "ي", "ی", -1)
	str = strings.Replace(str, "ك", "ک", -1)
	
	// replace persian question mark
	str = strings.Replace(str, "?", "؟", -1)
	
	// replace persian number
	pnums := map[string]string{"0": "۰", "1": "۱", "2": "۲", "3": "۳", "4": "۴", "5": "۵", "6": "۶", "7": "۷", "8": "۸", "9": "۹",}
	for k,v := range pnums {
		str = strings.Replace(str, k, v, -1)
	}
	
	if len(str) != 0 && str[len(str)-1] == '-' {
		str = "-" + str[:len(str)-1]
	}
	
	// add rtl character
	// str = "\u202B" + str
	
	return str
}
