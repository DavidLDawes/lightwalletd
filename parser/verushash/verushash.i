/* File : verushash.i */
%module verushash

%{
#include "verushash.h"
%}

%insert(cgo_comment_typedefs) %{
#cgo LDFLAGS: -L/usr/local/lib -L/usr/local/lib64 -L${SRCDIR} -L/usr/lib/x86_64-linux-gnu -l:veruslib.so -l:libboost_system.so
#cgo CPPFLAGS: -O2 -march=x86-64 -msse4 -msse2 -msse -msse4.1 -msse4.2 -msse3 -mavx -maes -fomit-frame-pointer -fPIC -Wno-builtin-declaration-mismatch -I/home/virtualsoundnw/lightwalletd/parser/hash -I/usr/include/c++/8-I/usr/include/x86_64-linux-gnu/c++/8  -pthread -w
#cgo CXXFLAGS: -O2 -march=x86-64 -msse2 -msse -msse4 -msse4.1 -msse4.2 -msse3 -mavx -maes -fomit-frame-pointer -fPIC -Wno-builtin-declaration-mismatch -I/home/virtualsoundnw/lightwalletd/parser/hash -I/usr/include/c++/8 -I/usr/include/x86_64-linux-gnu/c++/8  -pthread -w
#cgo CFLAGS: -O2 -march=x86-64 -msse2 -msse -msse4 -msse4.1 -msse4.2 -msse3 -mavx -maes -fomit-frame-pointer -fPIC -Wno-builtin-declaration-mismatch -I/home/virtualsoundnw/lightwalletd/parser/hash -I/usr/include/c++/8 -I/usr/include/x86_64-linux-gnu/c++/8  -pthread -w
%}


%include "verushash.h"
