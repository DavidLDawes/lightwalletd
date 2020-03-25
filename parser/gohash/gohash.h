//-----------------------------------------------------------------------------
// Hash is a simple wrapper around the VerusCoin verus_hash algorithms.
// It is intended for use in the go lightwalletd project.
// Written by David Dawes, and is placed in the public
// domain. The author hereby disclaims copyright to this source code.

#ifndef _GOHASH_H_
#define _GOHASH_H_/* File : hash.h */

#include <stdio.h>
#include <stdint.h>
#include <string>

class Gohash {
public:
  unsigned char *result;
  std::string verushash(const char * bytes, int length);
  std::string verushash_reverse(const char * bytes, int length);
  std::string verushash_v2(const char * bytes, int length);
  std::string verushash_v2_reverse(const char * bytes, int length);
  char * vh_v2(const char * bytes, int length);
  std::string verushash_v2b(const char * bytes, int length);
  std::string verushash_v2b_reverse(const char * bytes, int length);
  char * vh_v2b(const char * bytes, int length);
  std::string verushash_v2b1(const char * bytes, int length);
  std::string verushash_v2b1_reverse(const char * bytes, int length);
  char * vh_v2b1(const char * bytes, int length);
};
#endif