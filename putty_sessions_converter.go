package main

import (
	"io/ioutil"
	"bytes"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"fmt"
	"os"
	"strings"
	"strconv"
	"gopkg.in/ini.v1"
	"net/url"
	"path"
	"runtime"
	"path/filepath"
)

const (
	CR             = "\r"
	LF             = "\n"
	CRLF           = "\r\n"
	SessionHiveKey = `HKEY_CURRENT_USER\Software\SimonTatham\PuTTY\Sessions`
)

func ReadFileUTF16(filename string) ([]byte, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	win16be := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	utf16bom := unicode.BOMOverride(win16be.NewDecoder())
	unicodeReader := transform.NewReader(bytes.NewReader(raw), utf16bom)
	decoded, err := ioutil.ReadAll(unicodeReader)
	return decoded, err
}

func main()  {
	sep := LF
	if runtime.GOOS == "windows" {
		sep = CRLF
	}

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, fmt.Sprintf("Usage: %s infile [outdir]%s", os.Args[0], sep))
		os.Exit(1)
	}

	filename := os.Args[1]

	sessionDir := "sessions"
	if len(os.Args) > 2 {
		sessionDir = os.Args[2]
	}

	raw, err := ReadFileUTF16(filename)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	lines := strings.Split(strings.Replace(string(raw), CRLF, LF, -1), LF)
	raw = []byte(strings.Join(lines[1:len(lines)-1], LF))

	cfg, err := ini.LooseLoad(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fail to read file: %v%s", err, sep)
		os.Exit(1)
	}

	if _, err := os.Stat(sessionDir); err != nil {
		if err = os.Mkdir(sessionDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Fail to create directory: %v%s", sessionDir, sep)
			os.Exit(1)
		}
	}

	for _, section := range cfg.Sections() {
		if strings.Index(section.Name(), SessionHiveKey) < 0 {
			continue
		}

		kh := strings.Replace(section.Name(), SessionHiveKey+`\`, "", -1)
		if strings.TrimSpace(kh) == SessionHiveKey {
			continue
		}

		sh := make([]string, 0)

		for _, key := range section.Keys() {
			value := key.Value()
			if strings.Index(value, "dword:") == 0 {
				dwordValue := strings.Split(value, "dword:")[1]
				i, err := strconv.ParseInt(dwordValue, 16, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr,"Can't parse dword value '%s', error: %s%s", value, err.Error(), sep)
					sh = append(sh, fmt.Sprintf("%s=%s", key.Name(), value))
					continue
				}

				sh = append(sh, fmt.Sprintf("%s=%d", key.Name(), i))
				continue
			}

			sh = append(sh, fmt.Sprintf("%s=%s", key.Name(), value))
		}

		sh = append(sh, "FontName=server:-misc-fixed-medium-r-normal--15-140-75-75-c-90-iso10646-1")
		sh = append(sh, "")

		if ph, err := url.PathUnescape(kh); err == nil {
			kh = ph
		}
		kh = strings.TrimSpace(kh)

		sessionFile := path.Join(sessionDir, url.PathEscape(kh))
		sessionFile, _ = filepath.Abs(sessionFile)

		ioutil.WriteFile(sessionFile, []byte(strings.Join(sh, LF)), 0644)
	}
}
