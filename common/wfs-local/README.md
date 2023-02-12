Web File System - Local files driver
=========

File system abstraction with access management
This is the Local File System adapter for the [core interface](https://github.com/xbsoftware/wfs)

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

### License 

MIT