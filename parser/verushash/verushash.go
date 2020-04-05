/* ----------------------------------------------------------------------------
 * This file was automatically generated by SWIG (http://www.swig.org).
 * Version 4.0.1
 *
 * This file is not intended to be easily readable and contains a number of
 * coding conventions designed to improve portability and efficiency. Do not make
 * changes to this file unless you know what you are doing--modify the SWIG
 * interface file instead.
 * ----------------------------------------------------------------------------- */

// source: parser/verushash/verushash.i

package verushash

/*
#define intgo swig_intgo
typedef void *swig_voidp;

#include <stdint.h>


typedef long long intgo;
typedef unsigned long long uintgo;



typedef struct { char *p; intgo n; } _gostring_;
typedef struct { void* array; intgo len; intgo cap; } _goslice_;



#cgo LDFLAGS: -L/usr/local/lib -L/usr/local/lib64 -L${SRCDIR} -L/usr/lib/x86_64-linux-gnu -l:veruslib.so -l:libboost_system.so
#cgo CPPFLAGS: -O2 -march=x86-64 -msse4 -msse2 -msse -msse4.1 -msse4.2 -msse3 -mavx -maes -fomit-frame-pointer -fPIC -Wno-builtin-declaration-mismatch -I/home/virtualsoundnw/lightwalletd/parser/hash -I/usr/include/c++/8-I/usr/include/x86_64-linux-gnu/c++/8  -pthread -w
#cgo CXXFLAGS: -O2 -march=x86-64 -msse2 -msse -msse4 -msse4.1 -msse4.2 -msse3 -mavx -maes -fomit-frame-pointer -fPIC -Wno-builtin-declaration-mismatch -I/home/virtualsoundnw/lightwalletd/parser/hash -I/usr/include/c++/8 -I/usr/include/x86_64-linux-gnu/c++/8  -pthread -w
#cgo CFLAGS: -O2 -march=x86-64 -msse2 -msse -msse4 -msse4.1 -msse4.2 -msse3 -mavx -maes -fomit-frame-pointer -fPIC -Wno-builtin-declaration-mismatch -I/home/virtualsoundnw/lightwalletd/parser/hash -I/usr/include/c++/8 -I/usr/include/x86_64-linux-gnu/c++/8  -pthread -w

typedef _gostring_ swig_type_1;
typedef _gostring_ swig_type_2;
typedef _gostring_ swig_type_3;
typedef _gostring_ swig_type_4;
typedef _gostring_ swig_type_5;
typedef _gostring_ swig_type_6;
typedef _gostring_ swig_type_7;
typedef _gostring_ swig_type_8;
typedef _gostring_ swig_type_9;
typedef _gostring_ swig_type_10;
typedef _gostring_ swig_type_11;
typedef _gostring_ swig_type_12;
typedef _gostring_ swig_type_13;
extern void _wrap_Swig_free_verushash_b5c2c4e4f55e7268(uintptr_t arg1);
extern uintptr_t _wrap_Swig_malloc_verushash_b5c2c4e4f55e7268(swig_intgo arg1);
extern void _wrap_Verushash_initialized_set_verushash_b5c2c4e4f55e7268(uintptr_t arg1, _Bool arg2);
extern _Bool _wrap_Verushash_initialized_get_verushash_b5c2c4e4f55e7268(uintptr_t arg1);
extern void _wrap_Verushash_initialize_verushash_b5c2c4e4f55e7268(uintptr_t arg1);
extern void _wrap_Verushash_anyverushash_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_1 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_anyverushash_height_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_2 arg2, swig_intgo arg3, uintptr_t arg4, swig_intgo arg5);
extern void _wrap_Verushash_anyverushash_reverse_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_3 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_anyverushash_reverse_height_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_4 arg2, swig_intgo arg3, uintptr_t arg4, swig_intgo arg5);
extern void _wrap_Verushash_verushash_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_5 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_verushash_reverse_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_6 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_verushash_v2_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_7 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_verushash_v2_reverse_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_8 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_verushash_v2b_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_9 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_verushash_v2b_reverse_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_10 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_verushash_v2b1_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_11 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_verushash_v2b1_reverse_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_12 arg2, swig_intgo arg3, uintptr_t arg4);
extern void _wrap_Verushash_reverse_verushash_b5c2c4e4f55e7268(uintptr_t arg1, swig_type_13 arg2);
extern uintptr_t _wrap_new_Verushash_verushash_b5c2c4e4f55e7268(void);
extern void _wrap_delete_Verushash_verushash_b5c2c4e4f55e7268(uintptr_t arg1);
#undef intgo
*/
import "C"

import "syscall"
import "unsafe"
import "sync"


type _ syscall.Sockaddr




type _ unsafe.Pointer



var Swig_escape_always_false bool
var Swig_escape_val interface{}


type _swig_fnptr *byte
type _swig_memberptr *byte


type _ sync.Mutex

func Swig_free(arg1 uintptr) {
	_swig_i_0 := arg1
	C._wrap_Swig_free_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0))
}

func Swig_malloc(arg1 int) (_swig_ret uintptr) {
	var swig_r uintptr
	_swig_i_0 := arg1
	swig_r = (uintptr)(C._wrap_Swig_malloc_verushash_b5c2c4e4f55e7268(C.swig_intgo(_swig_i_0)))
	return swig_r
}

type SwigcptrVerushash uintptr

func (p SwigcptrVerushash) Swigcptr() uintptr {
	return (uintptr)(p)
}

func (p SwigcptrVerushash) SwigIsVerushash() {
}

func (arg1 SwigcptrVerushash) SetInitialized(arg2 bool) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_Verushash_initialized_set_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), C._Bool(_swig_i_1))
}

func (arg1 SwigcptrVerushash) GetInitialized() (_swig_ret bool) {
	var swig_r bool
	_swig_i_0 := arg1
	swig_r = (bool)(C._wrap_Verushash_initialized_get_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0)))
	return swig_r
}

func (arg1 SwigcptrVerushash) Initialize() {
	_swig_i_0 := arg1
	C._wrap_Verushash_initialize_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0))
}

func (arg1 SwigcptrVerushash) Anyverushash(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_anyverushash_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_1)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Anyverushash_height(arg2 string, arg3 int, arg4 uintptr, arg5 int) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	_swig_i_4 := arg5
	C._wrap_Verushash_anyverushash_height_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_2)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3), C.swig_intgo(_swig_i_4))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Anyverushash_reverse(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_anyverushash_reverse_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_3)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Anyverushash_reverse_height(arg2 string, arg3 int, arg4 uintptr, arg5 int) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	_swig_i_4 := arg5
	C._wrap_Verushash_anyverushash_reverse_height_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_4)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3), C.swig_intgo(_swig_i_4))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Verushash(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_verushash_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_5)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Verushash_reverse(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_verushash_reverse_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_6)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Verushash_v2(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_verushash_v2_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_7)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Verushash_v2_reverse(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_verushash_v2_reverse_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_8)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Verushash_v2b(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_verushash_v2b_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_9)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Verushash_v2b_reverse(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_verushash_v2b_reverse_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_10)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Verushash_v2b1(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_verushash_v2b1_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_11)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Verushash_v2b1_reverse(arg2 string, arg3 int, arg4 uintptr) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	_swig_i_2 := arg3
	_swig_i_3 := arg4
	C._wrap_Verushash_verushash_v2b1_reverse_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_12)(unsafe.Pointer(&_swig_i_1)), C.swig_intgo(_swig_i_2), C.uintptr_t(_swig_i_3))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func (arg1 SwigcptrVerushash) Reverse(arg2 string) {
	_swig_i_0 := arg1
	_swig_i_1 := arg2
	C._wrap_Verushash_reverse_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0), *(*C.swig_type_13)(unsafe.Pointer(&_swig_i_1)))
	if Swig_escape_always_false {
		Swig_escape_val = arg2
	}
}

func NewVerushash() (_swig_ret Verushash) {
	var swig_r Verushash
	swig_r = (Verushash)(SwigcptrVerushash(C._wrap_new_Verushash_verushash_b5c2c4e4f55e7268()))
	return swig_r
}

func DeleteVerushash(arg1 Verushash) {
	_swig_i_0 := arg1.Swigcptr()
	C._wrap_delete_Verushash_verushash_b5c2c4e4f55e7268(C.uintptr_t(_swig_i_0))
}

type Verushash interface {
	Swigcptr() uintptr
	SwigIsVerushash()
	SetInitialized(arg2 bool)
	GetInitialized() (_swig_ret bool)
	Initialize()
	Anyverushash(arg2 string, arg3 int, arg4 uintptr)
	Anyverushash_height(arg2 string, arg3 int, arg4 uintptr, arg5 int)
	Anyverushash_reverse(arg2 string, arg3 int, arg4 uintptr)
	Anyverushash_reverse_height(arg2 string, arg3 int, arg4 uintptr, arg5 int)
	Verushash(arg2 string, arg3 int, arg4 uintptr)
	Verushash_reverse(arg2 string, arg3 int, arg4 uintptr)
	Verushash_v2(arg2 string, arg3 int, arg4 uintptr)
	Verushash_v2_reverse(arg2 string, arg3 int, arg4 uintptr)
	Verushash_v2b(arg2 string, arg3 int, arg4 uintptr)
	Verushash_v2b_reverse(arg2 string, arg3 int, arg4 uintptr)
	Verushash_v2b1(arg2 string, arg3 int, arg4 uintptr)
	Verushash_v2b1_reverse(arg2 string, arg3 int, arg4 uintptr)
	Reverse(arg2 string)
}


