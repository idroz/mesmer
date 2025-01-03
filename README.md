# mezmer
OP-XY/Z sound visualiser

## Building from Source
The build system was tested only on a mac.

Required packages:

```bash
brew install glfw
brew install pkg-config
brew install mingw-w64
```

Build for Darwin
```bash
GOOS=darwin GOARCH=arm64 go build -o bin/mezmer
```

```bash
CGO_ENABLED=1 GOOS=windows CC="x86_64-w64-mingw32-gcc" GOARCH=amd64 go build -o dist/Mezmer_Win_x64.exe
```

``` bash
CGO_ENABLED=1 GOOS=windows CC="i686-w64-mingw32-gcc" GOARCH=386 go build -o dist/Mezmer_Win_x32.exe
```