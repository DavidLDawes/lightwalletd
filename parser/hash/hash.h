//-----------------------------------------------------------------------------
// Hash is a simple wrapper around the VerusCoin verus_hash algorithms.
// It is intended for use in the go lightwalletd project.
// Written by David Dawes, and is placed in the public
// domain. The author hereby disclaims copyright to this source code.

#ifndef _HASH_H_
#define _HASH_H_/* File : hash.h */

#include <stdio.h>
class Hash {
public:
  bool initialized = false;
  void initialize();
  void verushash(char * hash_result, const char * bytes, int length);
  void verushash_reverse(char * hash_result, const char * bytes, int length);
  void verushash_v2(char * hash_result, const char * bytes, int length);
  void verushash_v2_reverse(char * hash_result, const char * bytes, int length);
  void verushash_v2b(char * hash_result, const char * bytes, int length);
  void verushash_v2b_reverse(char * hash_result, const char * bytes, int length);
  void verushash_v2b1(char * hash_result, const char * bytes, int length);
  void verushash_v2b1_reverse(char * hash_result, const char * bytes, int length);
};
#endif