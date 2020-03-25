/* File : class.cxx */

#include "gohash.h"
#include "crypto/verus_hash.h"

std::string Gohash::verushash(const char * bytes, int length) {
    char *hash = new char[32];
    verus_hash(hash, (const unsigned char*) bytes, length);
    return std::string(hash, 32);
}

std::string Gohash::verushash_reverse(const char * bytes, int length) {
    char *hash = new char[32];
    verus_hash(hash, bytes, length);
    for (int i=0; i<16; i++) {
            hash[i], hash[31 - i] = hash[31 - i], hash[i];
    }
    return std::string(hash, 32);
}

std::string Gohash::verushash_v2(const char * bytes, int length) {
    return std::string(vh_v2(bytes, length), 32);
}

std::string Gohash::verushash_v2_reverse(const char * bytes, int length) {
    char *hash = vh_v2(bytes, length);
    for (int i=0; i<16; i++) {
            hash[i], hash[31 - i] = hash[31 - i], hash[i];
    }
    return std::string(hash, 32);
}

char * Gohash::vh_v2(const char * bytes, int length) {
    char *hash = new char[32];
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    vh2.Reset();
    vh2.Write((const unsigned char*) bytes, length);
    vh2.Finalize((unsigned char *) hash);
    return hash;
}

std::string Gohash::verushash_v2b(const char * bytes, int length) {
    return std::string(vh_v2b(bytes, length), 32);
}

std::string Gohash::verushash_v2b_reverse(const char * bytes, int length) {
    char *hash = vh_v2b(bytes, length);
    for (int i=0; i<16; i++) {
            hash[i], hash[31 - i] = hash[31 - i], hash[i];
    }
    return std::string(hash, 32);
}

char * Gohash::vh_v2b(const char * bytes, int length) {
    char *hash = new char[32];
    CVerusHashV2 vh2(SOLUTION_VERUSHHASH_V2);
    vh2.Reset();
    vh2.Write((const unsigned char*) bytes, length);
    vh2.Finalize2b((unsigned char *) hash);
    return hash;
}


std::string Gohash::verushash_v2b1(const char * bytes, int length) {
    return std::string(vh_v2b1(bytes, length), 32);
}

std::string Gohash::verushash_v2b1_reverse(const char * bytes, int length) {
    char *hash = vh_v2b1(bytes, length);
    for (int i=0; i<16; i++) {
            hash[i], hash[31 - i] = hash[31 - i], hash[i];
    }
    return std::string(hash, 32);
}

char * Gohash::vh_v2b1(const char * bytes, int length) {
    char *hash = new char[32];
    CVerusHashV2 vh2b1(SOLUTION_VERUSHHASH_V2_1);
    vh2b1.Reset();
    vh2b1.Write((const unsigned char*) bytes, length);
    vh2b1.Finalize2b((unsigned char *) hash);
    return hash;
}
