/* File : verushash.cxx */

#include "verushash.h"

#include <stdint.h>
#include <vector>

#include "crypto/verus_hash.h"

void Verushash::initialize() {
    if (!initialized)
    {
        CVerusHash::init();
        CVerusHashV2::init();
    }
    initialized = true;
}

void Verushash::verushash(const char * bytes, int length, void * hashresult) {
    initialize();
    verus_hash(hashresult, (const unsigned char*) bytes, length);
}

void Verushash::verushash_reverse(const char * bytes, int length, void * hashresult) {
    verushash(bytes, length, hashresult);
    char * chash = (char *) hashresult;
    reverse((char *) hashresult);
}

void Verushash::verushash_v2(const char * bytes, int length, void * hashresult) {
    initialize();
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    vh2.Reset();
    vh2.Write((const unsigned char*) bytes, length);
    vh2.Finalize((unsigned char*) hashresult);
}

void Verushash::verushash_v2_reverse(const char * bytes, int length, void * hashresult) {
    verushash_v2(bytes, length, hashresult);
    reverse((char *) hashresult);
}

void Verushash::verushash_v2b(const char * bytes, int length, void * hashresult) {
    initialize();
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    vh2.Reset();
    vh2.Write((const unsigned char*) bytes, length);
    vh2.Finalize2b((unsigned char*) hashresult);
}

void Verushash::verushash_v2b_reverse(const char * bytes, int length, void * hashresult) {
    verushash_v2b(bytes, length, hashresult);
    reverse((char *) hashresult);
}

void Verushash::verushash_v2b1(const char * bytes, int length, void * hashresult) {
    initialize();
    CVerusHashV2 vh2b1(SOLUTION_VERUSHHASH_V2_1);
    vh2b1.Reset();
    vh2b1.Write((const unsigned char*) bytes, length);
    vh2b1.Finalize2b((unsigned char*) hashresult);
}

void Verushash::verushash_v2b1_reverse(const char * bytes, int length, void * hashresult) {
    verushash_v2b1(bytes, length, hashresult);
    reverse((char *) hashresult);
}

void Verushash::reverse(char * swapme) {
    for (int i=0; i<16; i++) {
            swapme[i], swapme[31 - i] = swapme[31 - i], swapme[i];
    }
}
