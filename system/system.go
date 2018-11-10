package system

import "os"
import "io"
import "fmt"
import "time"
import "math/rand"
import "strconv"
import "path/filepath"

func randomString() string {
	rand.Seed(time.Now().UnixNano() + int64(os.Getpid()))

	str := strconv.Itoa(10000 + rand.Intn(10000))

	return str
}

func TempFileName(pattern string) string {
	for i := 0; i < 1000; i++ {
		filename := pattern + randomString() + ".tmp"
		path := filepath.Join(os.TempDir(), filename)
		if _, err := os.Stat(path); !os.IsExist(err) {
			return path
		}
	}

	return ""
}

func Copy(src, dst string) (int64, error) {
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
