/* File : class.cxx */

#include "hash.h"

#include <stdint.h>
#include <vector>

#include "crypto/verus_hash.h"

unsigned char * Hash::verushash(const void * bytes, int length) {
    verus_hash(result, (const unsigned char*) bytes, length);
    return result;
}

unsigned char * Hash::verushash_v2(const void * bytes, int length) {
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    vh2.Reset();
    vh2.Write((const unsigned char*) bytes, length);
    vh2.Finalize(result);
    return result;
}

unsigned char * Hash::verushash_v2b(const void * bytes, int length) {
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    vh2.Reset();
    vh2.Write((const unsigned char*) bytes, length);
    vh2.Finalize2b(result);
    return result;
}

unsigned char * Hash::verushash_v2b1(const void * bytes, int length) {
    CVerusHashV2 vh2b1(SOLUTION_VERUSHHASH_V2_1);
    vh2b1.Reset();
    vh2b1.Write((const unsigned char*) bytes, length);
    vh2b1.Finalize2b(result);
    return result;
}