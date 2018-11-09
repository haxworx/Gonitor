package system

import "os"
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
	filename := pattern + randomString() + ".tmp"

	for i := 0; i < 1000; i++ {
		path := filepath.Join(os.TempDir(), filename)
		if _, err := os.Stat(path); !os.IsExist(err) {
			return path
		}
	}

	return ""
}