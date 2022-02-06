//nolint
package helpers

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/stretchr/testify/assert"

	. "glauth-ui-light/config"
)

// tools

func deleteFile(file string) {
	// delete file
	var err = os.Remove(file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}

func copyTmpFile(source, dest string) {
	content, _ := ioutil.ReadFile(source)
	ioutil.WriteFile(dest, content, 0640)
}

func clean(file string) {
	deleteFile(file + ".1")
	copyTmpFile(file+".orig", file)
	deleteFile(file)
}

func lazyContains(s []string, e string) bool {
	for _, a := range s {
		if strings.Contains(strings.ReplaceAll(a, " ", ""), strings.ReplaceAll(e, " ", "")) {
			return true
		}
	}
	return false
}

func readFile(f string) []string {
	var origfile []string
	file, _ := os.Open(f)
	// Start reading from the file using a scanner.
	scanner := bufio.NewScanner(file)
	read := false
	for scanner.Scan() {
		line := scanner.Text()

		if !read && strings.HasPrefix(line, "[[users]]") {
			read = true
		}
		if read {
			origfile = append(origfile, line)
		}
	}
	file.Close()
	return origfile
}

// Tests

func TestDB(t *testing.T) {

	cfg := WebConfig{
		DBfile:  "_sample-simple.cfg",
                Locale: Locale{
                        Lang: "en",
                        Path: "../routes/",
                },
		Debug:   true,
		Verbose: false,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6500,
		},
		PassPolicy: PassPolicy{
			AllowReadSSHA256: true,
		},
	}
	copyTmpFile(cfg.DBfile+".orig", cfg.DBfile)

	defer clean(cfg.DBfile)

	origfile := readFile(cfg.DBfile)
	//fmt.Printf("=== orig data ===\n%s\n", strings.Join(origfile, "\n"))

	data, head, _ := ReadDB(&cfg)
	WriteDB(&cfg, data, "test")
	newfile := readFile(cfg.DBfile)
	//fmt.Printf("\n=== new data ===\n%s\n============\n", strings.Join(newfile, "\n"))

	// Missing User CustomAttrs management
	except := []string{
		`[[users.customattributes]]`,
		`employeetype = ["Intern", "Temp"]`,
		`employeenumber = [12345, 54321]`,
	}

	for _, ol := range origfile {
		if !lazyContains(newfile, ol) {
			if !strings.Contains(ol, "#") {
				fmt.Printf(">%s\n", strings.ReplaceAll(ol, " ", ""))
				assert.Equal(t, true, lazyContains(except, ol), ol)
			}
		}
	}

	data2, head2, _ := ReadDB(&cfg)
	assert.Equal(t, true, strings.Contains(head2[0], "# Updated by test on "), "head2 updated header")
	assert.Equal(t, true, cmp.Equal(head, head2[1:]), "same heads")
	assert.Equal(t, true, cmp.Equal(data, data2), "same data")
}
