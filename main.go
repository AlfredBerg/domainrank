package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const TRANCO_FILE_PREFIX = "tranco-"

type Tranco struct {
	Domains map[string]int
}

func NewTranco() (Tranco, error) {
	path, err := getTrancoPath()
	if err != nil {
		return Tranco{}, err
	}

	f, err := os.Open(path)
	if err != nil {
		return Tranco{}, err
	}

	domains := make(map[string]int)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		l := strings.Split(line, ",")
		if len(l) != 2 {
			return Tranco{}, fmt.Errorf("unable to parse tranco row %s", line)
		}
		domain := l[1]
		rank, err := strconv.Atoi(l[0])
		if err != nil {
			return Tranco{}, err
		}
		domains[domain] = rank
	}
	if err := sc.Err(); err != nil {
		return Tranco{}, err
	}

	t := Tranco{}
	t.Domains = domains

	return t, nil
}

func (t *Tranco) Rank(domain string) int {
	r := t.Domains[domain]
	if r == 0 {
		return -1
	}
	return r
}

func getTrancoPath() (string, error) {
	tmpDir := path.Join(os.TempDir(), "domainrank")

	err := os.MkdirAll(tmpDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", err
	}

	mostRecentTrancofile := 0
	for _, f := range files {
		if !strings.HasPrefix(f.Name(), TRANCO_FILE_PREFIX) {
			continue
		}
		s := strings.Split(f.Name(), "-")
		if !(len(s) == 2) {
			continue
		}

		tEpoch, err := strconv.Atoi(s[1])
		if err != nil {
			return "", err
		}
		if tEpoch > mostRecentTrancofile {
			mostRecentTrancofile = tEpoch
		}
	}

	trancoPath := ""
	if mostRecentTrancofile == 0 || time.Since(time.Unix(int64(mostRecentTrancofile), 0)) > time.Hour*24*7 {
		//Get new tranco list
		r, err := http.Get("https://tranco-list.eu/top-1m-id")
		if err != nil {
			return "", err
		}
		if r.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed getting tranco list id, unepected status code %d", r.StatusCode)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return "", err
		}

		r, err = http.Get(fmt.Sprintf("https://tranco-list.eu/download/%s/full", string(b)))
		if err != nil {
			return "", err
		}
		if r.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed getting tranco list, unepected status code %d", r.StatusCode)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		epoch := time.Now().Unix()
		err = os.WriteFile(path.Join(tmpDir, fmt.Sprintf("%s%d", TRANCO_FILE_PREFIX, epoch)), body, os.ModePerm)
		if err != nil {
			return "", err
		}
		trancoPath = path.Join(tmpDir, fmt.Sprintf("%s%d", TRANCO_FILE_PREFIX, epoch))

		//Delete old tranco list
		if mostRecentTrancofile != 0 {
			err := os.Remove(path.Join(tmpDir, fmt.Sprintf("%s%d", TRANCO_FILE_PREFIX, mostRecentTrancofile)))
			if err != nil {
				return "", err
			}
		}
	} else {
		trancoPath = path.Join(tmpDir, fmt.Sprintf("%s%d", TRANCO_FILE_PREFIX, mostRecentTrancofile))
	}

	return trancoPath, nil
}

func main() {
	inFile := ""
	flag.StringVar(&inFile, "i", "", "File to read input from, STDIN is not set")
	flag.Parse()

	t, err := NewTranco()
	if err != nil {
		log.Panicln(err)
	}

	var sc *bufio.Scanner
	if inFile == "" {
		sc = bufio.NewScanner(os.Stdin)
	} else {
		f, err := os.Open(inFile)
		if err != nil {
			log.Panicln(err)
		}
		sc = bufio.NewScanner(f)
	}

	for sc.Scan() {
		rawDomain := sc.Text()
		domain := rawDomain

		apex, err := publicsuffix.EffectiveTLDPlusOne(domain)
		if err != nil {
			log.Printf("failed to get apex for %s due to: %s", domain, err)
			continue
		}
		domain = apex

		fmt.Printf("%s %s %d\n", rawDomain, domain, t.Rank(domain))

	}
	if sc.Err() != nil {
		log.Panicln(sc.Err())
	}
}
