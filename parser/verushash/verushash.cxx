/* File : verushash.cxx */

#include "verushash.h"
#include "solutiondata.h"
#include <sstream>
#include <stdint.h>
#include <vector>
#include <iostream>

#include "include/sodium/core.h"
#include "include/verus_hash.h"

void Verushash::initialize() {
    if (!initialized)
    {
        CVerusHash::init();
        CVerusHashV2::init();

        int si = sodium_init();
    }
    initialized = true;
}

void Verushash::anyverushash(const char * bytes, int length, void * hashresult) {
    std::cout << "anyverushash";
    if (bytes[0] == 4 and bytes[2] >= 1) {
            if (bytes[2] < 3) {
                verushash_v2b(bytes, length, hashresult);
            } else {
                verushash_v2b1(bytes, length, hashresult);
            }
    } else {
                verushash(bytes, length, hashresult);
    }
}

void Verushash::anyverushash_height(const char * bytes, int length, void * hashresult, int height) {
    std::cout << "anyverushash_height:" << height;
    if (height > 105359) {
        verushash_v2b2(bytes, length, hashresult);
    }
    if (bytes[0] == 4 and bytes[2] >= 1) {
            if (bytes[2] < 3) {
                if (height > 800199) {
                    verushash_v2b1(bytes, length, hashresult);
                } else {
                    verushash_v2b(bytes, length, hashresult);
                }
            } else {
                verushash_v2b1(bytes, length, hashresult);
            }
    } else {
                verushash(bytes, length, hashresult);
    }
}

void Verushash::anyverushash_reverse(const char * bytes, int length, void * hashresult) {
    std::cout << "anyverushash_reverse";
    if (bytes[0] == 4 and bytes[2] >= 1) {
            if (bytes[2] < 3) {
                verushash_v2b_reverse(bytes, length, hashresult);
            } else {
                verushash_v2b1_reverse(bytes, length, hashresult);
            }
    } else {
            verushash_reverse(bytes, length, hashresult);
    }
}

void Verushash::anyverushash_reverse_height(const char * bytes, int length, void * hashresult, int height) {
    std::cout << "anyverushash_reverse_height:" << height;
    if (height > 105359) {
        verushash_v2b2_reverse(bytes, length, hashresult);
    }
    if (bytes[0] == 4 and bytes[2] >= 1) {
            if (bytes[2] < 3) {
                if (height > 800199) {
                    verushash_v2b1_reverse(bytes, length, hashresult);
                } else {
                    verushash_v2b_reverse(bytes, length, hashresult);
                }
            } else {
                verushash_v2b1_reverse(bytes, length, hashresult);
            }
    } else {
            verushash_reverse(bytes, length, hashresult);
    }
}

void Verushash::verushash(const char * bytes, int length, void * hashresult) {
    initialize();
    verus_hash(hashresult, bytes, length);
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

void Verushash::verushash_v2b2(const char * bytes, int length, void * hashresult) {
    uint256 result;

    std::cout << ":verushash_v2b2\n";
    initialize();
    CVerusHashV2 vh2b1(SOLUTION_VERUSHHASH_V2_2);
    vh2b1.Reset();

/*
    vh2b1.Write((const unsigned char*) bytes, length);
    vh2b1.Finalize2b((unsigned char*) hashresult);
*/

    CBlockHeader bh;
    CDataStream s(bytes, bytes + length, SER_GETHASH, 0);

    try
    {
        s >> bh;
        result = bh.GetVerusV2Hash();
        memcpy(hashresult, (void *) &result, 32);
    }
    catch(const std::exception& e)
    {
        std::cout << ":verushash_v2b2 got an exception\n";
    }

}

void Verushash::verushash_v2b2_reverse(const char * bytes, int length, void * hashresult) {
    std::cout << ":verushash_v2b2_reverse\n";
    verushash_v2b2(bytes, length, hashresult);
    reverse((char *) hashresult);
}

void Verushash::reverse(char * swapme) {
    for (int i=0; i<16; i++) {
            swapme[i], swapme[31 - i] = swapme[31 - i], swapme[i];
    }
}
