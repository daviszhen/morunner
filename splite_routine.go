package main

import (
	"os"
	"regexp"
	"strings"
)

var rtstart = "^goroutine.*\\[|\ngoroutine.*\\["

func split(ipath string) error {
	//
	rtCmp, err := regexp.Compile(rtstart)
	if err != nil {
		return err
	}

	fdata, err := os.ReadFile(ipath)
	if err != nil {
		return err
	}
	fcontent := string(fdata)

	seps := rtCmp.FindAllStringIndex(fcontent, -1)

	//output
	opath := ipath + "_split.txt"
	ofile, err := os.OpenFile(opath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer ofile.Close()
	for idx, sep := range seps {
		start := sep[0]
		end := 0
		if idx == len(seps)-1 {
			end = len(fcontent)
		} else {
			end = seps[idx+1][0]
		}
		s := fcontent[start:end]
		s = strings.TrimSpace(s)
		s = strings.ReplaceAll(s, "\n", "\\n")
		_, err = ofile.WriteString(s)
		if err != nil {
			return err
		}
		_, err = ofile.WriteString("\n")
		if err != nil {
			return err
		}
	}
	return nil
}
