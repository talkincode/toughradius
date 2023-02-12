package local

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/talkincode/toughradius/common/wfs"
)

type LocalDrive struct {
	root string
}

type localFile struct {
	path string
	id   string
}

func (l localFile) GetPath() string {
	return l.path
}
func (l localFile) ClientID() string {
	return l.id
}
func (l localFile) IsFolder() bool {
	fi, err := os.Stat(l.GetPath())
	if err == nil && fi.IsDir() {
		return true
	}
	return false
}
func (l localFile) Contains(target wfs.FileID) bool {
	if strings.HasPrefix(target.GetPath(), l.GetPath()+string(filepath.Separator)) {
		return true
	}

	return false
}

type localFileInfo struct {
	os.FileInfo
	f wfs.FileID
}

func (i localFileInfo) File() wfs.FileID {
	return i.f
}

func NewLocalDrive(path string, config *wfs.DriveConfig) (wfs.Drive, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.New("Invalid path: " + root)
	}

	d := LocalDrive{root: root}

	return wfs.NewDrive(&d, config), nil
}

func (l *LocalDrive) Comply(f wfs.FileID, operation int) bool {
	return strings.Contains(filepath.Clean(f.GetPath()), l.root)
}

func (l *LocalDrive) ToFileID(id string) wfs.FileID {
	return localFile{filepath.Clean(filepath.Join(l.root, filepath.FromSlash(id))), id}
}

func (l *LocalDrive) GetParent(f wfs.FileID) wfs.FileID {
	return localFile{filepath.Dir(f.GetPath()), filepath.Dir(f.ClientID())}
}

func (l *LocalDrive) newLocalFile(path string) localFile {
	return localFile{path, filepath.ToSlash(strings.Replace(path, l.root, "", 1))}
}

func (l *LocalDrive) Remove(f wfs.FileID) error {
	return os.RemoveAll(f.GetPath())
}
func (l *LocalDrive) Read(f wfs.FileID) (io.ReadSeeker, error) {
	file, err := os.Open(f.GetPath())
	if err != nil {
		return nil, errors.New("Can't open file for reading")
	}
	return file, nil
}
func (l *LocalDrive) Write(f wfs.FileID, data io.Reader) error {
	file, err := os.Create(f.GetPath())
	if err != nil {
		return errors.New("Can't open file for writing")
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		return errors.New("Can't write data")
	}

	return err
}
func (l *LocalDrive) Make(f wfs.FileID, name string, isFolder bool) (wfs.FileID, error) {
	full := filepath.Join(f.GetPath(), name)
	out := l.newLocalFile(full)
	if isFolder {
		return out, os.MkdirAll(full, os.FileMode(int(0700)))
	} else {
		file, err := os.Create(full)
		if err != nil {
			return out, err
		}
		return out, file.Close()
	}
}

func (l *LocalDrive) Copy(source, target wfs.FileID, name string, isFolder bool) (wfs.FileID, error) {
	full := filepath.Join(target.GetPath(), name)
	if isFolder {
		return nil, copyDir(source.GetPath(), full)
	}

	// copy file
	return l.newLocalFile(full), copyFile(source.GetPath(), full)
}

// Move renames(moves) a file or a folder
func (l *LocalDrive) Move(source, target wfs.FileID, name string, isFolder bool) (wfs.FileID, error) {
	full := filepath.Join(target.GetPath(), name)
	return l.newLocalFile(full), os.Rename(source.GetPath(), full)
}

// Info returns info about a single file
func (l *LocalDrive) Info(f wfs.FileID) (wfs.FileInfo, error) {
	file, err := os.Stat(f.GetPath())
	if err != nil {
		return nil, err
	}

	return localFileInfo{file, f}, nil
}

func (l *LocalDrive) List(f wfs.FileID) ([]wfs.FileInfo, error) {
	full := f.GetPath()
	files, err := ioutil.ReadDir(full)
	if err != nil {
		return nil, err
	}

	info := make([]wfs.FileInfo, 0, len(files))
	for i := range files {
		info = append(info, localFileInfo{
			files[i],
			l.newLocalFile(filepath.Join(full, files[i].Name())),
		})
	}

	return info, nil
}

func (l *LocalDrive) Search(f wfs.FileID, search string) ([]wfs.FileInfo, error) {
	matches, err := glob(f.GetPath(), search)
	if err != nil {
		return nil, err
	}

	out := make([]wfs.FileInfo, len(matches))
	for i := range matches {
		info, _ := os.Stat(matches[i])
		out[i] = localFileInfo{info, l.newLocalFile(matches[i])}
	}

	return out, nil
}

func (l *LocalDrive) Exists(f wfs.FileID, name string) bool {
	full := f.GetPath()
	if name != "" {
		full = filepath.Join(full, name)
	}

	_, err := os.Stat(full)
	if err != nil {
		return false
	}

	return true
}

func (l *LocalDrive) Stats() (uint64, uint64, error) {
	return getFSSize(l.root)
}
