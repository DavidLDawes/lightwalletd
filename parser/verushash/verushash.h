//-----------------------------------------------------------------------------
// Hash is a simple wrapper around the VerusCoin verus_hash algorithms.
// It is intended for use in the go lightwalletd project.
// Written by David Dawes, and is placed in the public
// domain. The author hereby disclaims copyright to this source code.

#ifndef _VERUSHASH_H_
#define _VERUSHASH_H_/* File : veruhash.h */

#include <stdio.h>
class Verushash {
public:
  bool initialized = false;
  void initialize();
  void anyverushash(const char * bytes, int length, void * hashresult);
  void anyverushash_height(const char * bytes, int length, void * hashresult, int height);
  void anyverushash_reverse(const char * bytes, int length, void * hashresult);
  void anyverushash_reverse_height(const char * bytes, int length, void * hashresult, int height);
  void verushash(const char * bytes, int length, void * hashresult);
  void verushash_reverse(const char * bytes, int length, void * hashresult);
  void verushash_v2(const char * bytes, int length, void * hashresult);
  void verushash_v2_reverse(const char * bytes, int length, void * hashresult);
  void verushash_v2b(const char * bytes, int length, void * hashresult);
  void verushash_v2b_reverse(const char * bytes, int length, void * hashresult);
  void verushash_v2b1(const char * bytes, int length, void * hashresult);
  void verushash_v2b1_reverse(const char * bytes, int length, void * hashresult);
  void reverse(char * swapme);
};
#endif