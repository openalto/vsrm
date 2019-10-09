package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// Replica metadata including all candidates
type Replica struct {
	candidates []string
}

// Fingerprint of a replica
type Fingerprint struct {
	MD5  string
	Size int64
}

func main() {
	replicas := make(map[Fingerprint]Replica)
	root := os.Args[1]
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			var md5sum []byte
			f, fErr := os.Open(path)
			if fErr == nil {
				defer f.Close()
				h := md5.New()
				if _, hErr := io.Copy(h, f); hErr == nil {
					md5sum = h.Sum(nil)
				}
			}
			if md5sum != nil {
				// Cache file replica info
				fmt.Println(path, hex.EncodeToString(md5sum), info.Size())
				key := Fingerprint{hex.EncodeToString(md5sum), info.Size()}
				r, ok := replicas[key]
				if ok {
					r.candidates = append(r.candidates, path)
				} else {
					replicas[key] = Replica{[]string{path}}
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	dumpToDB(replicas)
}

func arrayToStr(array []string) string {
	result := "{"
	for _, s := range array {
		result += fmt.Sprintf("%q", s)
	}
	result += "}"
	return result
}

func dumpToDB(replicas map[Fingerprint]Replica) {
	w := csv.NewWriter(os.Stdout)
	w.Comma = '\t'
	for key, r := range replicas {
		if err := w.Write([]string{key.MD5, strconv.FormatInt(key.Size, 10), arrayToStr(r.candidates)}); err != nil {
			// Skip this entry
			continue
		}
		// fmt.Println([]string{key.MD5, strconv.FormatInt(key.Size, 10), arrayToStr(r.candidates)})
	}
	w.Flush()
}
