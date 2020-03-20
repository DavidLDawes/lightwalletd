// Go callable version of verushash-node
//
//

#include <stdint.h>
#include <vector>

#include "crypto/verus_hash.h"

bool initialized = false;

void initialize()
{
    if (!initialized)
    {
        CVerusHash::init();
        CVerusHashV2::init();
    }
    initialized = true;
}

unsigned char * verushash(const std::string bytes) {
    unsigned char *result = new unsigned char[32];

    if (initialized == false) {
        initialize();
    }
    verus_hash(result, bytes.data(), bytes.size());
    return result;
}

unsigned char * verushash_v2(const std::string bytes) {
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    unsigned char *result = new unsigned char[32];

    if (initialized == false) {
        initialize();
    }

    vh2.Reset();
    vh2.Write((const unsigned char *)bytes.data(), bytes.size());
    vh2.Finalize(result);
    return result;
}

unsigned char * verushash_v2b(const std::string bytes) {
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    unsigned char *result = new unsigned char[32];

    if (initialized == false) {
        initialize();
    }

    vh2.Reset();
    vh2.Write((const unsigned char *)bytes.data(), bytes.size());
    vh2.Finalize2b(result);
    return result;
}

unsigned char * verushash_v2b1(const std::string bytes) {
    CVerusHashV2 vh2b1(SOLUTION_VERUSHHASH_V2_1);
    unsigned char *result = new unsigned char[32];

    if (initialized == false) {
        initialize();
    }

    vh2b1.Reset();
    vh2b1.Write((const unsigned char *)bytes.data(), bytes.size());
    vh2b1.Finalize2b(result);
    return result;
}
