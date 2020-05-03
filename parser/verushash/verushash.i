/* File : verushash.i */
%module verushash

%{
#include "verushash.h"
%}

%insert(cgo_comment_typedefs) %{
#cgo LDFLAGS: -L${SRCDIR}  -l:libverushash.a
%}


%include "verushash.h"
