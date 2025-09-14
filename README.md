# go Shred

 Shred(path) is a simple Go function that overwrites your file with random bytes before deletion.

## Quick Start

```bash
go get github.com/Acu43/go-shred
```

Then in your code:

```go
package main

import (
    "log"
    "github.com/Acu43/go-shred"
)

func main() {
    if err := main.Shred("highly-top-secret-stuff.pdf"); err != nil {
        log.Fatal("Oops, didn't work. Would you consider creating a bug?", err)
    }
}
```

##  What it does

When you call `Shred()`, here's what happens behind the scenes:

1. **Triple overwrite** - Fills your file with random garbage data (3 times!)
2. **Truncate** - Shrinks the file down to nothing
3. **Sync to disk** - Makes sure changes are actually written
4. **Delete** - Removes the file entry

##  Use it standalone

Want a command-line tool instead?

```bash
go build -o go-shred
./go-shred my-sensitive-file.txt
```

## Reality check

### Pros

- This is really nice for not leaving any sensitive files laying around in the filesystem and stopping intro level file rescue programs.

### Cons

- The tool doesn't actually shred anything. The most reliable way to destroy the data without a doubt is phscially destoring the hardware. Writing with random data and removing the file is a low hanging fruit. We get some safety with relatively low cost. There are levels to safety.
- From an abstract point of view, the data we write may not be actually commited to the actual physical hardware. This could depend on compiler, libraries, OS, hardware. This would be my assumption unless it's explicitly specified. Here none of this is guaranteed. (to have an abstract view see C memset, memset_s functions.)

- There is a reason why deleting a big file doesnt take as much time as creating a big one. This stuff is optimised for everday use. Use this with caution, beware of performance penalties.
