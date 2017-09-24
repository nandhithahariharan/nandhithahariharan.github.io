package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/lunny/html2md"
)

const (
	fileName = "/Users/sudarshanreddy/dev/blog/nan/mickey/start.txt"
)

var (
	skipRegexBuildStrings = []string{"https://nandhithahariharan1.wordpress.com",
		"# Menu",
		"\\[",
		"\\[Skip to content\\]\\(#content\\)",
		"Kaleidoscope of Life",
		"#comments",
		"\\[CONNECT\\]\\(#\\)",
		"\\[Twitter\\]\\(https://twitter.com/nandhitha\\)",
		"\\[Facebook\\]\\(https://www.facebook.com/nandhithahariharan\\)",
		"\\[Linkedin\\]\\(https://www.linkedin.com/profile/view\\?id=137570705\\)",
		"(January|February|March|April|May|June?|July|August|September|October|November|December)\\s(\\d\\d?).+?(\\d\\d\\d\\d)",
	}
)

func main() {

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	links, err := getPosts(file)
	if err != nil {
		log.Fatal(err)
	}

	re := regexBuilder(skipRegexBuildStrings)
	datecapture := regexp.MustCompile("[0-9]{4}/(0[1-9]|1[0-2])/(0[1-9]|[1-2][0-9]|3[0-1])")

	for _, link := range *links {
		<-time.NewTicker(500 * time.Millisecond).C
		html, err := getHTML(link)
		if err != nil {
			log.Fatal(err)
		}
		outputFile := path.Base(link)
		date := strings.Replace(datecapture.FindString(link), "/", "-", -1)
		err = writeMarkDownToFile(html, outputFile, re, date)
		if err != nil {
			log.Fatal("writing to file failed", outputFile, err)
		}
	}

}

func writeMarkDownToFile(content []byte, fileName string, re *regexp.Regexp, date string) (err error) {
	rawmd := html2md.Convert(string(content))
	t, err := os.Create(date + "-" + fileName + ".markdown")
	if err != nil {
		return err
	}
	var data bytes.Buffer
	reader := strings.NewReader(rawmd)
	bufReader := bufio.NewReader(reader)
	var title string
	excerpt := ""

	defer func() {
		if err == io.EOF || err == nil {
			if len(excerpt) > 35 {
				excerpt = excerpt[:35]
			}
			header := getHeader(title, date, date, excerpt, "unclassified", "unclassified", "black.jpg")
			fullFile := append(header, data.Bytes()...)
			_, err = t.Write(fullFile)
		}

		t.Close()
	}()

	arrivedAtTitle := false
	for {
		line, _, err := bufReader.ReadLine()
		if err != nil {
			return err
		}
		//Trims Blank Spaced lines
		if !arrivedAtTitle && len(strings.Trim(string(line), " ")) == 0 {
			continue
		}

		//Trims Blank Tabbed lines
		if !arrivedAtTitle && len(strings.Trim(string(line), "\t")) == 0 {
			continue
		}

		//Skips any lines that match the regex supplied
		if re != nil {
			if re.MatchString(string(line)) {
				continue
			}
		}

		//The Blog usually ends here
		if string(line) == "Advertisements" {
			return nil
		}

		if !arrivedAtTitle {
			title = strings.TrimPrefix(string(line), "#")
			arrivedAtTitle = true
		} else if len(excerpt) < 15 {
			excerpt = string(line)
		}

		data.Write(line)
		data.Write([]byte("\n"))
	}
}

func regexBuilder(skipStrings []string) *regexp.Regexp {
	var regStr string
	for _, each := range skipStrings {
		regStr += each
		regStr += "|"
	}
	return regexp.MustCompile(strings.TrimSuffix(regStr, "|"))
}

func getHeader(vals ...string) []byte {
	var fileHeader = `---
layout: post
title: "%s"
date: %s 00:00:00
last_modified_at: %s 00:00:00
excerpt: "%s..." 
categories: %s
tags: %s
image: 
  feature: %s
  topPosition: -100px
bgContrast: dark
bgGradientOpacity: darker
syntaxHighlighter: no
---
`
	if len(vals) < 7 {
		return []byte(fileHeader)
	}

	header := fmt.Sprintf(fileHeader, vals[0], vals[1], vals[2], vals[3], vals[4], vals[5], vals[6])
	return []byte(header)

}

func getPosts(file io.Reader) (*[]string, error) {
	var links []string
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return &links, nil
			}
			return nil, err
		}
		links = append(links, string(line))
	}
}

func getHTML(link string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return bdy, nil
}
