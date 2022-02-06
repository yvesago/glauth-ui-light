package helpers

import (
	"bufio"
	"bytes"

	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "glauth-ui-light/config"

	"github.com/hydronica/toml"
)

func copyfile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func findnext(file string) (int, error) {
	rxExt := regexp.MustCompile(`\.(\d+)$`)
	root := filepath.Dir(file)

	next := 0
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if strings.Contains(s, file) {
			matches := rxExt.FindStringSubmatch(s)
			if matches != nil {
				i, _ := strconv.Atoi(matches[1])
				if i > next {
					next = i
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return next + 1, nil
}

func WriteDB(cfgw *WebConfig, data Ctmp, username string) error {
	_, head, err := ReadDB(cfgw)
	if err != nil {
		return err
	}
	// fmt.Println(strings.Join(head, "\n"))
	next, e := findnext(cfgw.DBfile)
	if e != nil {
		return e
	}

	_, err = copyfile(cfgw.DBfile, fmt.Sprintf("%s.%d", cfgw.DBfile, next))
	if err != nil {
		return err
	}

	file, err := os.OpenFile(cfgw.DBfile, os.O_RDWR, 0o640)
	if err != nil {
		return err
	}

	out := Ctmp{Users: data.Users, Groups: data.Groups}
	buf := new(bytes.Buffer)
	err = toml.NewEncoder(buf).Encode(out)
	if err != nil {
		return err
	}
	// fmt.Println(buf.String())

	currentTime := time.Now()
	var top []string
	top = append(top, fmt.Sprintf("# Updated by %s on %s", username, currentTime.Format(time.RFC3339)))
	top = append(top, head...)
	top = append(top, buf.String())
	newdata := []byte(strings.Join(top, "\n"))
	if err = os.WriteFile(cfgw.DBfile, newdata, 0o640); err != nil { //nolint:gosec //gosec cant't read octal perm
		file.Close()
		return err
	}
	return file.Close()
}

func ReadDB(cfgw *WebConfig) (Ctmp, []string, error) {
	file, err := os.Open(cfgw.DBfile)
	if err != nil {
		file.Close()
		return Ctmp{}, nil, fmt.Errorf("Non-existent config path: %s", cfgw.DBfile)
	}

	var headfile []string
	// Start reading from the file using a scanner.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[[users]]") {
			break
		}
		headfile = append(headfile, line)
	}

	cfg := Config{}
	// var md toml.MetaData
	_, err = toml.DecodeFile(cfgw.DBfile, &cfg)
	// md, err = toml.DecodeFile(configFileLocation, &cfg)
	if err != nil {
		file.Close()
		return Ctmp{}, nil, err
	}

	/*
	   will need to patch toml for customattributes:
	   add to `decode_meta.go`:

	   ```
	   func (md *MetaData) Mappings() map[string]interface{} {
	     return md.mapping
	   }
	   ```

	   switch users := md.Mappings()["users"].(type) {
	   case []map[string]interface{}:
	        for _, mduser := range users {
	                if mduser["customattributes"] != nil {
	                        for idx, cfguser := range cfg.Users {
	                                if cfguser.Name == mduser["name"].(string) {
	                                        switch attributes := mduser["customattributes"].(type) {
	                                        case []map[string]interface{}:
	                                                cfg.Users[idx].CustomAttrs = attributes[0]
	                                        case map[string]interface{}:
	                                                cfg.Users[idx].CustomAttrs = attributes
	                                        default:
	                                                log.Println("Unknown attribute structure in config file", "attributes", attributes)
	                                        }
	                                        break
	                                }
	                        }
	                }
	        }
	   }
	*/
	// b, _ := json.MarshalIndent(&cfg, "", "  ")
	// log.Print(string(b))
	data := Ctmp{Users: cfg.Users, Groups: cfg.Groups}
	return data, headfile, file.Close()
}
