package system

import "os"
import "io"
import "fmt"
import "time"
import "math/rand"
import "strings"
import "path/filepath"

func randomString(length int) string {
	var sb strings.Builder
	symbols := []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
		'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R',
		'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}

	nanoseconds := time.Now().UnixNano()

	rand.Seed(nanoseconds + int64(os.Getpid()))

	for i := 0; i < length; i++ {
		ch := byte(symbols[rand.Intn(len(symbols))])
		nanoseconds = time.Now().UnixNano()
		if nanoseconds%2 == 0 {
			ch += 32
		}
		sb.WriteByte(ch)
	}
	return sb.String()
}

func TempFileName(pattern string) string {
	for i := 0; i < 1000; i++ {
		rand.Seed(time.Now().UnixNano())
		rng := 8 + rand.Intn(16-8)
		filename := pattern + "-" + randomString(rng) + ".tmp"
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
