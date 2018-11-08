package main;

import "container/list"
import "io/ioutil"
import "log"
import "path"
import "fmt"
import "os"
import "time"

const (
	DefaultStateDir = ".monitor"
	DefaultStateFile = "statefile"
)

const ModeDiscard = os.ModeDir | os.ModeSymlink | os.ModeNamedPipe | os.ModeSocket | os.ModeDevice

type File struct {
	Path string
	Mtime int64
	Size int64
}

type Monitor struct {
	Root string

	Cwd string

	stateDir string
	starePath string

	Previous *list.List
	OnAdd func(string)
	OnDel func(string)
	OnMod func(string)
}

func (m *Monitor) SetDirectory(dirpath string) {
	m.Root = dirpath;
}

func (m *Monitor) ReadFiles(list *list.List, dirpath string) *list.List {
	files, err := ioutil.ReadDir(dirpath); if err != nil {
		// log.Fatal(err)
	}

	for _, f := range files {
		base := f.Name();
		if base[0] == '.' { continue; }

		path := path.Join(dirpath, base);
		if f.Mode() & ModeDiscard == 0 {
			file := File { Path: path, Mtime: f.ModTime().Unix(), Size: f.Size() };
	                list.PushBack(file);
			continue;
		}
		m.ReadFiles(list, path);
	}

	return list;
}

func (m *Monitor) Scan() *list.List {
	if m.Root == "" {
		log.Fatal("No directory set.");
	}

	list := list.New();
	list = m.ReadFiles(list, m.Root);

	return list;
}

func fileExists(list *list.List, path string) bool {
	for el := list.Front(); el != nil; el = el.Next() {
		if el.Value.(File).Path == path {
			return true;
		}
	}

	return false;
}

func (m *Monitor) findModFiles(ch chan bool, first *list.List, second *list.List) {
	fmt.Println("START MOD");
	for l2 := second.Front(); l2 != nil; l2 = l2.Next() {
		filename := l2.Value.(File).Path;
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
	fmt.Println("START ADD");
	for l2 := second.Front(); l2 != nil; l2 = l2.Next() {
		filename := l2.Value.(File).Path;
		if !fileExists(first, filename) {
			if m.OnAdd != nil {
				m.OnAdd(filename)
			}
		}
	}

	ch <- true
}

func (m *Monitor) findDelFiles(ch chan bool, first *list.List, second *list.List) {
	fmt.Println("START DEL");
	for l1 := first.Front(); l1 != nil; l1 = l1.Next() {
		filename := l1.Value.(File).Path;
		if !fileExists(second, filename) {
			if m.OnDel != nil {
				m.OnDel(filename)
			}
		}
	}

	ch <- true
}

func (m *Monitor) Compare(current *list.List) {
	ch := make(chan bool, 3);

	go m.findDelFiles(ch, m.Previous, current);
	go m.findModFiles(ch, m.Previous, current);
	go m.findAddFiles(ch, m.Previous, current);

	for i := 0; i < cap(ch); i++ {
		fmt.Println("AYE");
		<-ch
	}
}

func (m *Monitor) Init() {
	var err error
	m.Cwd, err = os.Getwd(); if err != nil {
		log.Fatal(err)
	}
	m.stateDir = path.Join(m.Cwd, DefaultStateDir, DefaultStateFile)

	m.SetDirectory(m.Cwd)
}

func (m *Monitor) SaveState() {

}

func (m *Monitor) Watch() {
	m.Previous = m.Scan()

	for {
		time.Sleep(3 * time.Second);
		current := m.Scan()
		m.Compare(current);
		m.SaveState();
		m.Previous = current;
	}
}

func OnAdd(filepath string) {
	fmt.Printf("ADD: %s\n", filepath);
}

func OnDel(filepath string) {
	fmt.Printf("DEL: %s\n", filepath);
}

func OnMod(filepath string) {
	fmt.Printf("MOD: %s\n", filepath);
}

func main() {
	m := new(Monitor);
	m.Init();
	fmt.Println("BEGIN");

	m.OnAdd = OnAdd
	m.OnDel = OnDel
	m.OnMod = OnMod

	m.Watch();
}
