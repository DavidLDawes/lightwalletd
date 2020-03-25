//-----------------------------------------------------------------------------
// Hash is a simple wrapper around the VerusCoin verus_hash algorithms.
// It is intended for use in the go lightwalletd project.
// Written by David Dawes, and is placed in the public
// domain. The author hereby disclaims copyright to this source code.

#ifndef _HASH_H_
#define _HASH_H_/* File : hash.h */

class Hash {
public:
  Hash() {
    result = new unsigned char[32];
  }
  virtual ~Hash() {
    delete(result);
  }

  unsigned char *result;
  unsigned char * verushash(const void * bytes, int length);
  unsigned char * verushash_v2(const void * bytes, int length);
  unsigned char * verushash_v2b(const void * bytes, int length);
  unsigned char * verushash_v2b1(const void * bytes, int length);
};
#endif