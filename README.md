# ioengine
IO engine library supports BufferedIO/DirectIO/AsyncIO and provides a unified common unix-like system file operation interface.

# usage
Import package:

```
import (
	"github.com/silentsharer/ioengine"
)
```

```
go get github.com/silentsharer/ioengine
```

# example
```
// create AIO ioengine option
opt := ioengine.DefaultOptions
opt.IOEngine = AIO

fd, err := ioengine.Open("/tmp/test", opt)
if err != nil {
	handler(err)
}
defer fd.Close()

data1, err := ioengine.MemAlign(4*ioengine.BlockSize)
if err != nil {
	handler(err)
}
copy(data1, []byte("hello"))

data2, err := ioengine.MemAlign(4*ioengine.BlockSize)
if err != nil {
	handler(err)
}
copy(data1, []byte("world"))

b := NewBuffers()
b.Write(data1).Write(data2)

fd.WriteAtv(*b, 0)
fd.Append(*b)
```

# license
Apache (see [LICENSE](https://github.com/silentsharer/ioengine/blob/master/LICENSE) file)