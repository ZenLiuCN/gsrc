# pkgen 
written by zen.Liu  
go helper for generate *.pc file and for fast switch pkg with different architecture
## script wapper
```shell
#!/bin/sh
root=`dirname $0`
cd $root
./pkg/pkgen $*
```
## usage
`pkgen i686` generate *.pc file in current binary path for i686 architecture  
`pkgen x86_64` generate *.pc file in current binary path for x86_64 architecture
## directory structure example
before generate
```
-/pkgen.exe
-/libiconv-1.15-3/
-/libiconv-1.15-3/i686/lib/libcharset.a
-/libiconv-1.15-3/i686/lib/libcharset.dll.a
-/libiconv-1.15-3/i686/lib/libiconv.a
-/libiconv-1.15-3/i686/lib/libiconv.dll.a
-/libiconv-1.15-3/i686/lib/pkgconfig/iconv.pc
-/libiconv-1.15-3/i686/include/...
-/libiconv-1.15-3/i686/...
-/libiconv-1.15-3/x86_64/lib/libcharset.a
-/libiconv-1.15-3/x86_64/lib/libcharset.dll.a
-/libiconv-1.15-3/x86_64/lib/libiconv.a
-/libiconv-1.15-3/x86_64/lib/libiconv.dll.a
-/libiconv-1.15-3/x86_64/lib/pkgconfig/iconv.pc
-/libiconv-1.15-3/x86_64/include/...
-/libiconv-1.15-3/x86_64/...
```
after generate
```
-/pkgen.exe
-/iconv.pc
-/libiconv-1.15-3/
-/libiconv-1.15-3/i686/lib/libcharset.a
-/libiconv-1.15-3/i686/lib/libcharset.dll.a
-/libiconv-1.15-3/i686/lib/libiconv.a
-/libiconv-1.15-3/i686/lib/libiconv.dll.a
-/libiconv-1.15-3/i686/lib/pkgconfig/iconv.pc
-/libiconv-1.15-3/i686/include/...
-/libiconv-1.15-3/i686/...
-/libiconv-1.15-3/x86_64/lib/libcharset.a
-/libiconv-1.15-3/x86_64/lib/libcharset.dll.a
-/libiconv-1.15-3/x86_64/lib/libiconv.a
-/libiconv-1.15-3/x86_64/lib/libiconv.dll.a
-/libiconv-1.15-3/x86_64/lib/pkgconfig/iconv.pc
-/libiconv-1.15-3/x86_64/include/...
-/libiconv-1.15-3/x86_64/...
```
source of `-/libiconv-1.15-3/i686/lib/pkgconfig/iconv.pc`
```
prefix=/mingw32
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include

Name: iconv
Description: libiconv
URL: https://www.gnu.org/software/libiconv/
Version: 1.15
Libs: -L${libdir} -liconv
Cflags: -I${includedir}
```
source of `-/libiconv-1.15-3/x86_64/lib/pkgconfig/iconv.pc`
```
prefix=/mingw64
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include

Name: iconv
Description: libiconv
URL: https://www.gnu.org/software/libiconv/
Version: 1.15
Libs: -L${libdir} -liconv
Cflags: -I${includedir}
```
when use `i686` result `-/iconv.pc` will be
```
prefix=/libiconv-1.15-3/i686
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include

Name: iconv
Description: libiconv
URL: https://www.gnu.org/software/libiconv/
Version: 1.15
Libs: -L${libdir} -liconv
Cflags: -I${includedir}
```
when use `x86_64` result `-/iconv.pc` will be
```
prefix=/libiconv-1.15-3/x86_64
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include

Name: iconv
Description: libiconv
URL: https://www.gnu.org/software/libiconv/
Version: 1.15
Libs: -L${libdir} -liconv
Cflags: -I${includedir}
```
