/* File : class.cxx */

#include "hash.h"

#include <stdint.h>
#include <vector>

#include "crypto/verus_hash.h"

void Hash::verushash(char * hash, const char * bytes, int length) {
    verus_hash(result, (const unsigned char*) bytes, length);
    memcpy(hash, result, 32);
}

void Hash::verushash_reverse(char * hash, const char * bytes, int length) {
    verushash(hash, bytes, length);
    for (int i=0; i<16; i++) {
            hash[i], hash[31 - i] = hash[31 - i], hash[i];
    }
}

void Hash::verushash_v2(char * hash, const char * bytes, int length) {
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    vh2.Reset();
    vh2.Write((const unsigned char*) bytes, length);
    vh2.Finalize(result);
    memcpy(hash, result, 32);
}

void Hash::verushash_v2_reverse(char * hash, const char * bytes, int length) {
    verushash_v2(hash, bytes, length);
    for (int i=0; i<16; i++) {
            hash[i], hash[31 - i] = hash[31 - i], hash[i];
    }

}

void Hash::verushash_v2b(char * hash, const char * bytes, int length) {
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    vh2.Reset();
    vh2.Write((const unsigned char*) bytes, length);
    vh2.Finalize2b(result);
    memcpy(hash, result, 32);
}

void Hash::verushash_v2b_reverse(char * hash, const char * bytes, int length) {
    verushash_v2b(hash, bytes, length);
    for (int i=0; i<16; i++) {
            hash[i], hash[31 - i] = hash[31 - i], hash[i];
    }
}

void Hash::verushash_v2b1(char * hash, const char * bytes, int length) {
    CVerusHashV2 vh2b1(SOLUTION_VERUSHHASH_V2_1);
    vh2b1.Reset();
    vh2b1.Write((const unsigned char*) bytes, length);
    vh2b1.Finalize2b(result);
    memcpy(hash, result, 32);
}

void Hash::verushash_v2b1_reverse(char * hash, const char * bytes, int length) {
    verushash_v2b1(hash, bytes, length);
    for (int i=0; i<16; i++) {
            hash[i], hash[31 - i] = hash[31 - i], hash[i];
    }
}
