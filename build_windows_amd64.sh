CGO_ENABLED=1 GOOS=windows CC="x86_64-w64-mingw32-gcc" GOARCH=amd64 go build -o dist/Mezmer_Win_x64.exe

CGO_ENABLED=1 GOOS=windows CC="i686-w64-mingw32-gcc" GOARCH=386 go build -o dist/Mezmer_Win_x32.exe