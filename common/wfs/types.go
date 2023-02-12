package wfs

import (
	"io"
	"os"
)

// File stores info about single file
type File struct {
	Name  string `json:"value"`
	ID    string `json:"id"`
	Size  int64  `json:"size"`
	Date  int64  `json:"date"`
	Type  string `json:"type"`
	Files []File `json:"data,omitempty"`
}

// File stores info about single file
type Drive interface {
	List(id string, config ...*ListConfig) ([]File, error)
	Search(id, search string, config ...*ListConfig) ([]File, error)
	Remove(id string) error
	Read(id string) (io.ReadSeeker, error)
	Write(id string, data io.Reader) error
	Exists(id string) bool
	Info(id string) (File, error)
	Make(id, name string, isFolder bool) (string, error)
	Copy(source, target, name string) (string, error)
	Move(source, target, name string) (string, error)
	Stats() (uint64, uint64, error)
}

type FileInfo interface {
	os.FileInfo
	File() FileID
}

type Adapter interface {
	// implements Policy
	Comply(FileID, int) bool

	// converts client id <-> server id
	ToFileID(id string) FileID
	GetParent(f FileID) FileID

	// file operations
	List(id FileID) ([]FileInfo, error)
	Search(id FileID, search string) ([]FileInfo, error)
	Remove(id FileID) error
	Read(id FileID) (io.ReadSeeker, error)
	Write(id FileID, data io.Reader) error
	Make(id FileID, name string, isFolder bool) (FileID, error)
	Copy(source, target FileID, name string, isFolder bool) (FileID, error)
	Move(source, target FileID, name string, isFolder bool) (FileID, error)
	Info(id FileID) (FileInfo, error)
	Exists(id FileID, name string) bool
	Stats() (uint64, uint64, error)
}

type FileID interface {
	GetPath() string
	ClientID() string
	IsFolder() bool
	Contains(FileID) bool
}

// ListConfig contains file listing options
type ListConfig struct {
	SkipFiles  bool
	SubFolders bool
	Nested     bool
	Exclude    MatcherFunc
	Include    MatcherFunc
}

// DriveConfig contains drive configuration
type DriveConfig struct {
	Verbose   bool
	List      *ListConfig
	Operation *OperationConfig
	Policy    *Policy
}

// OperationConfig contains file operation options
type OperationConfig struct {
	PreventNameCollision bool
}

// MatcherFunc receives path and returns true if path matches the rule
type MatcherFunc func(string) bool

// Supported operation modes
const (
	ReadOperation int = iota
	WriteOperation
)

// Policy is a rule which allows or denies operation
type Policy interface {
	// Comply method returns true is operation for the path is allowed
	Comply(FileID, int) bool
}
