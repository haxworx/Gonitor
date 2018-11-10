package monitor

import "container/list"
import "io/ioutil"
import "bufio"
import "log"
import "path/filepath"
import "fmt"
import "os"
import "time"
import "errors"
import "../system"

type File struct {
	Path  string
	Mtime int64
	Size  int64
}

type statefile struct {
	directory string
	fullpath  string
	enabled  bool
}

type Monitor struct {
	workingDirectory string

	pollInterval time.Duration

	stateFile statefile

	filesPrevious *list.List

	OnAdd func(string)
	OnDel func(string)
	OnMod func(string)
}

func New() *Monitor {
	m := new(Monitor)

	return m
}

func (m *Monitor) readFiles(list *list.List, dirpath string) *list.List {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		// log.Fatal(err)
	}

	const ModeDiscard = os.ModeDir | os.ModeSymlink | os.ModeNamedPipe | os.ModeSocket | os.ModeDevice

	for _, f := range files {
		base := f.Name()
		if base[0] == '.' {
			continue
		}

		path := filepath.Join(dirpath, base)
		if f.Mode()&ModeDiscard == 0 {
			file := File{Path: path, Mtime: f.ModTime().Unix(), Size: f.Size()}
			list.PushBack(file)
			continue
		}
		m.readFiles(list, path)
	}

	return list
}

func (m *Monitor) Scan() *list.List {
	if m.workingDirectory == "" {
		log.Fatal("No directory set.")
	}

	list := list.New()
	list = m.readFiles(list, m.workingDirectory)

	return list
}

func fileExists(list *list.List, path string) bool {
	for el := list.Front(); el != nil; el = el.Next() {
		if el.Value.(File).Path == path {
			return true
		}
	}

	return false
}

func (m *Monitor) findModFiles(ch chan bool, first *list.List, second *list.List) {
	for l2 := second.Front(); l2 != nil; l2 = l2.Next() {
		filename := l2.Value.(File).Path
		for l1 := first.Front(); l1 != nil; l1 = l1.Next() {
			if filename == l1.Value.(File).Path &&
				l1.Value.(File).Mtime != l2.Value.(File).Mtime {
				if m.OnMod != nil {
					m.OnMod(filename)
				}
			}
		}
	}

	ch <- true
}

func (m *Monitor) findAddFiles(ch chan bool, first *list.List, second *list.List) {
	for l2 := second.Front(); l2 != nil; l2 = l2.Next() {
		filename := l2.Value.(File).Path
		if !fileExists(first, filename) {
			if m.OnAdd != nil {
				m.OnAdd(filename)
			}
		}
	}

	ch <- true
}

func (m *Monitor) findDelFiles(ch chan bool, first *list.List, second *list.List) {
	for l1 := first.Front(); l1 != nil; l1 = l1.Next() {
		filename := l1.Value.(File).Path
		if !fileExists(second, filename) {
			if m.OnDel != nil {
				m.OnDel(filename)
			}
		}
	}

	ch <- true
}

func (m *Monitor) Compare(filesPrevious *list.List, filesCurrent *list.List) {
	ch := make(chan bool, 3)

	go m.findDelFiles(ch, filesPrevious, filesCurrent)
	go m.findModFiles(ch, filesPrevious, filesCurrent)
	go m.findAddFiles(ch, filesPrevious, filesCurrent)

	for i := 0; i < cap(ch); i++ {
		<-ch
	}
}

func (m *Monitor) SetDirectory(directory string) {
	m.workingDirectory = directory
}

func (m *Monitor) SetStateFile(directory string, filename string) {
	var err error

	if m.workingDirectory == "" {
		log.Printf("No working directory set.")
		os.Exit(1 << 0)
	}

	m.stateFile.directory = filepath.Join(m.workingDirectory, directory)
	if _, err = os.Stat(m.stateFile.directory); os.IsNotExist(err) {
		err = os.Mkdir(m.stateFile.directory, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	m.stateFile.fullpath = filepath.Join(m.stateFile.directory, filename)

	m.stateFile.enabled = true;
}

func (m *Monitor) ClearStateFiles() bool {
	if !m.stateFile.enabled {
		return false
	}
	os.Remove(m.stateFile.fullpath)
	err := os.RemoveAll(m.stateFile.directory)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (m *Monitor) LoadSavedState() {
	m.filesPrevious = list.New()

	if !m.stateFile.enabled {
		return
	}

	if _, err := os.Stat(m.stateFile.fullpath); os.IsNotExist(err) {
		return
	}

	f, err := os.Open(m.stateFile.fullpath)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scan := bufio.NewScanner(f)

	for scan.Scan() {
		file := File{}
		_, err = fmt.Sscanf(scan.Text(), "%d\t%d\t%s\n", &file.Mtime, &file.Size, &file.Path)
		if err != nil {
			log.Fatal(err)
		}
		m.filesPrevious.PushBack(file)
	}
}

func (m *Monitor) SaveState(files *list.List) {
	if !m.stateFile.enabled {
		return
	}

	tmpPath := system.TempFileName("monitor")
	if tmpPath == "" {
		log.Fatal(errors.New("system.TempFileName"))
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		log.Fatal(err)
	}

	for el := files.Front(); el != nil; el = el.Next() {
		file := el.Value.(File)
		_, err = fmt.Fprintf(f, "%d\t%d\t%s\n", file.Mtime, file.Size, file.Path)
		if err != nil {
			log.Fatal(err)
		}
	}

	f.Close()
	system.Copy(tmpPath, m.stateFile.fullpath)
	os.Remove(tmpPath)
}


func (m *Monitor) Watch() {
	m.LoadSavedState()

	for {
		filesCurrent := m.Scan()
		m.Compare(m.filesPrevious, filesCurrent)
		m.SaveState(filesCurrent)
		m.filesPrevious = filesCurrent
		time.Sleep(m.pollInterval)
	}
}

func (m *Monitor) SetPollInterval(seconds int) {
	m.pollInterval = time.Duration(seconds) * time.Second
}

func (m *Monitor) SetOnAddFunc(fn func(string)) {
	m.OnAdd = fn
}

func (m *Monitor) SetOnDelFunc(fn func(string)) {
	m.OnDel = fn
}

func (m *Monitor) SetOnModFunc(fn func(string)) {
	m.OnMod = fn
}
