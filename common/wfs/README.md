Web File System - core interface
=========

File system abstraction with access management.

API provides common file operations for some folder on local drive. Any operations outside of the folder will be blocked. Also, it possible to configure a custom policy for read/write operations.


Can be used as backend for Webix File Manager https://webix.com/filemanager

## API

### Initialization

```go
import (
	"github.com/xbsoftware/wfs-local"
)

fs, err := wfs.NewLocalDrive("./sandbox", nil)
```

### Get data

```go
//get files in a folder
files, err := fs.List("/subfolder");

//get files in a folder and subfolders as plain list
files, err = fs.List("/subfolder", &wfs.ListConfig{ SubFolders: true });

//get files in a folder and subfolders as nested structure
files, err = fs.List("/subfolder", &wfs.ListConfig{ SubFolders: true, Nested:true });

//get folder only
files, err = fs.List("/subfolder", &wfs.ListConfig{ SkipFiles: true });

//get files that match a mask
files, err = fs.List("/subfolder", &wfs.ListConfig{
    Include: func(file string) bool { return strings.HasSufix(file, ".txt") },
});

//ignore some files
files, err = fs.List("/subfolder", &wfs.ListConfig{
    Exclude: func(file string) bool { return file == ".git" },
});

//get info about a single file
info, err = fs.Info("some.txt");

//check if file exists
check := fs.Exists("some.txt");
```

### Modify files

```go
//make folder
fs.Make("/", "sub2", true);

//make file
fs.Make("/", "my.txt", false);

//remove
fs.Remove("some.txt");

//copy
fs.Copy("some.txt", "/sub/", "");

//copy as
fs.Copy("some.txt", "/sub/", "other.txt");

//move
fs.Move("some.txt", "/data/", "");

//rename
fs.Move("some.txt", "", "some-data.txt");

//read
reader, err := fs.Read("some.txt");

//write
fs.Write("some.txt", writer)
```

### Configuration

```go
// Access policies
// ForceRoot policy is added automatically
fs, err := wfs.NewLocalDrive("./sandbox", &wfs.DriveConfig{
    Policy: &ReadOnlyPolicy{},
})

// Logging
fs, err := wfs.NewLocalDrive("./sandbox", &wfs.DriveConfig{
    Verbose:true
})

// Exec operation with different config
path, err := fs.WithOperationConfig(&OperationConfig{
    PreventNameCollision:true,
}).Make("/", "some", true)
```

### License 

MIT