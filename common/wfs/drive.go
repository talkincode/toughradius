package wfs

import (
	"errors"
	"io"
	"log"
	"path"
	"sort"
	"strings"
)

// Drive represents an isolated file system
type DriveFacade struct {
	adapter   Adapter
	list      *ListConfig
	operation *OperationConfig
	verbose   bool

	policy Policy
}

// NewLocalDrive returns new LocalDrive object
// which represents the file folder on local drive
// due to ForceRootPolicy all operations outside of the root folder will be blocked
func NewDrive(adapter Adapter, config *DriveConfig) Drive {
	drive := DriveFacade{adapter: adapter, policy: adapter}

	if config != nil {
		drive.verbose = config.Verbose
		drive.list = config.List
		drive.operation = config.Operation

		if config.Policy != nil {
			drive.policy = CombinedPolicy{
				[]Policy{
					*config.Policy,
					drive.policy,
				},
			}
		}
	}

	if drive.list == nil {
		drive.list = &ListConfig{}
	}
	if drive.operation == nil {
		drive.operation = &OperationConfig{}
	}

	return &drive
}

// allow method checks is operation on object allowed or not
func (d *DriveFacade) allow(id FileID, operation int) bool {
	return d.policy.Comply(id, operation)
}

func (d *DriveFacade) Search(id, search string, config ...*ListConfig) ([]File, error) {
	path := d.adapter.ToFileID(id)
	if d.verbose {
		log.Printf("Search %s at %s", search, id)
	}

	if !d.allow(path, ReadOperation) {
		return nil, errors.New("Access Denied")
	}

	data, err := d.adapter.Search(path, search)
	if err != nil {
		return nil, err
	}

	out := make([]File, 0)
	for _, file := range data {
		out = append(out, File{file.Name(), id, file.Size(), file.ModTime().Unix(), GetType(file.Name(), file.IsDir()), nil})
	}

	return out, nil

}

// List method returns array of files from the target folder
func (d *DriveFacade) List(id string, config ...*ListConfig) ([]File, error) {
	path := d.adapter.ToFileID(id)

	if d.verbose {
		log.Printf("List %s", id)
	}

	if !d.allow(path, ReadOperation) {
		return nil, errors.New("Access Denied")
	}

	var list *ListConfig
	if len(config) > 0 {
		list = config[0]
	} else {
		list = d.list
	}

	if d.verbose {
		log.Printf("with config %+v", config)
	}

	return d.listFolder(path, list, nil)
}

// Remove deletes a file or a folder
func (d *DriveFacade) Remove(id string) error {
	path := d.adapter.ToFileID(id)

	if d.verbose {
		log.Printf("Remove %s", id)
	}

	if !d.allow(path, WriteOperation) {
		return errors.New("Access Denied")
	}

	return d.adapter.Remove(path)
}

// Read returns content of a file
func (d *DriveFacade) Read(id string) (io.ReadSeeker, error) {
	path := d.adapter.ToFileID(id)

	if d.verbose {
		log.Printf("Read %s", id)
	}

	if !d.allow(path, ReadOperation) {
		return nil, errors.New("Access Denied")
	}

	return d.adapter.Read(path)
}

// Write saves content to a file
func (d *DriveFacade) Write(id string, data io.Reader) error {
	if d.verbose {
		log.Printf("Write %s", id)
	}

	path := d.adapter.ToFileID(id)
	if !d.allow(path, WriteOperation) {
		return errors.New("Access Denied")
	}

	err := d.adapter.Write(path, data)
	if err != nil {
		return err
	}

	return nil
}

// Exists checks is file / folder with defined path does exist
func (d *DriveFacade) Exists(id string) bool {
	path := d.adapter.ToFileID(id)
	if !d.allow(path, ReadOperation) {
		return false
	}

	return d.adapter.Exists(path, "")
}

// Info returns info about a single file / folder
func (d *DriveFacade) Info(id string) (File, error) {
	path := d.adapter.ToFileID(id)
	if !d.allow(path, ReadOperation) {
		return File{}, errors.New("Access Denied")
	}

	info, err := d.adapter.Info(path)
	if err != nil {
		return File{}, errors.New("Access denied")
	}

	return File{ID: info.File().ClientID(), Name: info.Name(), Size: info.Size(), Date: info.ModTime().Unix(), Type: GetType(info.Name(), info.IsDir()), Files: nil}, nil
}

// Mkdir creates a new folder
func (d *DriveFacade) Make(id, name string, isFolder bool) (string, error) {
	if d.verbose {
		log.Printf("Make folder %s", id)
	}

	path := d.adapter.ToFileID(id)
	if !d.allow(path, WriteOperation) {
		return "", errors.New("Access Denied")
	}

	if d.operation.PreventNameCollision {
		var err error
		name, err = d.checkName(path, name, isFolder)
		if err != nil {
			return "", err
		}
	}

	path, err := d.adapter.Make(path, name, isFolder)
	if err != nil {
		return "", err
	}

	return path.ClientID(), nil
}

// Copy makes a copy of file or a folder
func (d *DriveFacade) Copy(source, target, name string) (string, error) {
	if d.verbose {
		log.Printf("Copy %s to %s", source, target)
	}

	if name == "" {
		name = path.Base(source)
	}
	from := d.adapter.ToFileID(source)
	to := d.adapter.ToFileID(target)

	if !d.allow(from, ReadOperation) || !d.allow(to, WriteOperation) {
		return "", errors.New("Access Denied")
	}

	st := from.IsFolder()
	if st && to.Contains(from) {
		return "", errors.New("Can't copy folder into self")
	}

	if d.operation.PreventNameCollision {
		var err error

		name, err = d.checkName(to, name, st)
		if err != nil {
			return "", err
		}
	}

	copied, err := d.adapter.Copy(from, to, name, st)
	if err != nil {
		return "", err
	}

	return copied.ClientID(), nil
}

// Move renames(moves) a file or a folder
func (d *DriveFacade) Move(source, target, name string) (string, error) {
	if d.verbose {
		log.Printf("Move %s to %s", source, target)
	}

	if name == "" {
		name = path.Base(source)
	}
	from := d.adapter.ToFileID(source)
	st := from.IsFolder()

	var to FileID
	if target == "" {
		to = d.adapter.GetParent(from)
	} else {
		to = d.adapter.ToFileID(target)
		if st && to.Contains(from) {
			return "", errors.New("Can't copy folder into self")
		}
	}

	if !d.allow(from, WriteOperation) || !d.allow(to, WriteOperation) {
		return "", errors.New("Access Denied")
	}

	if d.operation.PreventNameCollision {
		var err error

		name, err = d.checkName(to, name, st)
		if err != nil {
			return "", err
		}
	}

	moved, err := d.adapter.Move(from, to, name, st)
	if err != nil {
		return "", err
	}

	return moved.ClientID(), nil
}

func (d *DriveFacade) listFolder(path FileID, config *ListConfig, res []File) ([]File, error) {
	list, er := d.adapter.List(path)
	if er != nil {
		return nil, er
	}

	needSortData := false
	if config.Nested || res == nil {
		res = make([]File, 0, len(list))
		needSortData = true
	}

	for _, file := range list {
		skipFile := false
		if config.Exclude != nil && config.Exclude(file.Name()) {
			continue
		}
		if config.Include != nil && !config.Include(file.Name()) {
			skipFile = true
		}

		isDir := file.IsDir()
		if !isDir && (config.SkipFiles || skipFile) {
			continue
		}

		id := file.File().ClientID()
		fs := File{file.Name(), id, file.Size(), file.ModTime().Unix(), GetType(file.Name(), file.IsDir()), nil}

		if isDir && config.SubFolders {
			sub, err := d.listFolder(file.File(),
				config, res)

			fs.Type = "folder"
			if err != nil {
				return nil, err
			}

			if !config.Nested {
				res = sub
			} else if len(sub) > 0 {
				fs.Files = sub
			}
		}

		if !skipFile {
			res = append(res, fs)
		}
	}

	// sort files and folders by name, folders first
	if needSortData {
		sort.Slice(res, func(i, j int) bool {
			aFolder := res[i].Type == "folder"
			bFolder := res[j].Type == "folder"
			if (aFolder || bFolder) && res[i].Type != res[j].Type {
				return aFolder
			}

			return strings.ToUpper(res[i].Name) < strings.ToUpper(res[j].Name)
		})
	}

	return res, nil
}
func (d *DriveFacade) checkName(p FileID, name string, isFolder bool) (string, error) {
	counter := 0
	for d.adapter.Exists(p, name) {
		ext := path.Ext(name)

		if isFolder || ext == "" {
			name = name + ".new"
		} else {
			index := len(name) - len(ext)
			name = name[:index] + ".new" + name[index:]
		}

		counter++
		if counter > 20 {
			return name, errors.New("Can't create a new name for the file")
		}
	}

	return name, nil
}

// WithOperationConfig makes a copy of drive with new operation config
func (d *DriveFacade) WithOperationConfig(config *OperationConfig) Drive {
	copy := *d
	copy.operation = config

	return &copy
}

func (d *DriveFacade) Stats() (uint64, uint64, error) {
	return d.adapter.Stats()
}
