#ifndef __BLAKE2B_C_
#define __BLAKE2B_C_

#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "xxenc.c"

/**
 * The BLAKE2b initialization vectors
 */
static const uint64_t blake2b_IV[8] = {
    0x6a09e667f3bcc908UL, 0xbb67ae8584caa73bUL, 0x3c6ef372fe94f82bUL,
    0xa54ff53a5f1d36f1UL, 0x510e527fade682d1UL, 0x9b05688c2b3e6c1fUL,
    0x1f83d9abfb41bd6bUL, 0x5be0cd19137e2179UL
};

/**
 * Table of permutations
 */
static const uint8_t blake2b_sigma[12][16] = {
    { 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 },
    { 14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3 },
    { 11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4 },
    { 7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8 },
    { 9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13 },
    { 2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9 },
    { 12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11 },
    { 13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10 },
    { 6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5 },
    { 10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0 },
    { 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 },
    { 14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3 }
};

enum blake2b_constant
{
    BLAKE2B_BLOCKBYTES = 128,
    BLAKE2B_OUTBYTES = 64,
    BLAKE2B_KEYBYTES = 64,
    BLAKE2B_SALTBYTES = 16,
    BLAKE2B_PERSONALBYTES = 16
};

typedef struct blake2b_param
{
    uint8_t digest_length;                   /* 1 */
    uint8_t key_length;                      /* 2 */
    uint8_t fanout;                          /* 3 */
    uint8_t depth;                           /* 4 */
    uint32_t leaf_length;                    /* 8 */
    uint64_t node_offset;                    /* 16 */
    uint8_t node_depth;                      /* 17 */
    uint8_t inner_length;                    /* 18 */
    uint8_t reserved[14];                    /* 32 */
    uint8_t salt[BLAKE2B_SALTBYTES];         /* 48 */
    uint8_t personal[BLAKE2B_PERSONALBYTES]; /* 64 */
} blake2b_param;

typedef struct blake2b_state
{
    uint64_t h[8];                   /* chained state */
    uint64_t t[2];                   /* total number of bytes */
    uint64_t f[2];                   /* last block flag */
    uint8_t buf[BLAKE2B_BLOCKBYTES]; /* input buffer */
    size_t buflen;                   /* size of buffer */
    size_t outlen;                   /* digest size */
} blake2b_state;

#define ROTR64(w, c) ((w) >> (c)) | ((w) << (64 - (c)))

#define LOAD64(dest, load)                                          \
    do {                                                             \
    dest = ((uint64_t)((load)[0]) << 0) | ((uint64_t)((load)[1]) << 8) | \
    ((uint64_t)((load)[2]) << 16) | ((uint64_t)((load)[3]) << 24) |      \
    ((uint64_t)((load)[4]) << 32) | ((uint64_t)((load)[5]) << 40) |      \
    ((uint64_t)((load)[6]) << 48) | ((uint64_t)((load)[7]) << 56);       \
    } while(0)


#define G(a, b, c, d, x, y)       \
  do {                            \
  a = a + b + x;                  \
  d = ROTR64(d ^ a, 32);          \
  c = c + d;                      \
  b = ROTR64(b ^ c, 24);          \
  a = a + b + y;                  \
  d = ROTR64(d ^ a, 16);          \
  c = c + d;                      \
  b = ROTR64(b ^ c, 63);          \
  }while(0)

#define store64(dst,w) \
do {\
    uint8_t* p = dst;\
    p[0] = (uint8_t)(w >> 0);\
    p[1] = (uint8_t)(w >> 8);\
    p[2] = (uint8_t)(w >> 16);\
    p[3] = (uint8_t)(w >> 24);\
    p[4] = (uint8_t)(w >> 32);\
    p[5] = (uint8_t)(w >> 40);\
    p[6] = (uint8_t)(w >> 48);\
    p[7] = (uint8_t)(w >> 56);\
} while(0)


static inline void blake2b_increment_counter(blake2b_state* state, const uint64_t inc)
{
    state->t[0] += inc;
    state->t[1] += ((state->t[0] < inc)?1:0);
}

#define F(state,block) \
do{\
    size_t i, j;\
    uint64_t v[16], m[16], s[16];\
    for (i = 0; i < 16; ++i) {\
        LOAD64(m[i], block + (i * sizeof(m[i])));\
    }\
    for (i = 0; i < 8; ++i) {\
        v[i] = state->h[i];\
        v[i + 8] = blake2b_IV[i];\
    }\
    v[12] ^= state->t[0];\
    v[13] ^= state->t[1];\
    v[14] ^= state->f[0];\
    v[15] ^= state->f[1];\
    for (i = 0; i < 12; i++) {\
        for (j = 0; j < 16; j++) {\
            s[j] = blake2b_sigma[i][j];\
        }\
        G(v[0], v[4], v[8], v[12], m[s[0]], m[s[1]]);\
        G(v[1], v[5], v[9], v[13], m[s[2]], m[s[3]]);\
        G(v[2], v[6], v[10], v[14], m[s[4]], m[s[5]]);\
        G(v[3], v[7], v[11], v[15], m[s[6]], m[s[7]]);\
        G(v[0], v[5], v[10], v[15], m[s[8]], m[s[9]]);\
        G(v[1], v[6], v[11], v[12], m[s[10]], m[s[11]]);\
        G(v[2], v[7], v[8], v[13], m[s[12]], m[s[13]]);\
        G(v[3], v[4], v[9], v[14], m[s[14]], m[s[15]]);\
    }\
    for (i = 0; i < 8; i++) {\
        state->h[i] = state->h[i] ^ v[i] ^ v[i + 8];\
    }\
} while(0)

static inline void blake2b_init(blake2b_state* state, size_t outlen, const uint8_t * personal)
{
    blake2b_param P = {0};
    const uint8_t* p;
    size_t i;

    P.digest_length = (uint8_t)outlen;
    P.key_length=0;
    memcpy(P.personal,personal,16);
    P.fanout = 1;
    P.depth = 1;

    uint64_t dest = 0;
    p = (const uint8_t*)(&P);
    for (i = 0; i < 8; ++i) {
        state->h[i] = blake2b_IV[i];
    }
    for (i = 0; i < 8; ++i) {
        LOAD64(dest, p + (sizeof(state->h[i]) * i));
        state->h[i] ^= dest;
    }
    state->outlen = P.digest_length;
}

static inline void blake2b_update(blake2b_state* state, const unsigned char* input_buffer,
               size_t inlen)
{
    const unsigned char* in = input_buffer;
    size_t left = state->buflen;
    size_t fill = BLAKE2B_BLOCKBYTES - left;
    if (inlen > fill) {
        state->buflen = 0;
        memcpy(state->buf + left, in, fill);
        blake2b_increment_counter(state, BLAKE2B_BLOCKBYTES);
        F(state, state->buf);
        in += fill;
        inlen -= fill;

        while (inlen > BLAKE2B_BLOCKBYTES) {
            blake2b_increment_counter(state, BLAKE2B_BLOCKBYTES);
            F(state, in);
            in += BLAKE2B_BLOCKBYTES;
            inlen -= BLAKE2B_BLOCKBYTES;
        }
    }
    memcpy(state->buf + state->buflen, in, inlen);
    state->buflen += inlen;
}


static inline void blake2b_final(blake2b_state* state, uint8_t* out, size_t outlen)
{
    uint8_t buffer[BLAKE2B_OUTBYTES] = { 0 };
    size_t i;

    blake2b_increment_counter(state, state->buflen);

    /* set last chunk = true */
    state->f[0] = UINT64_MAX;

    /* padding */
    memset(state->buf + state->buflen, 0, BLAKE2B_BLOCKBYTES - state->buflen);
    F(state, state->buf);

    /* Store back in little endian */
    for (i = 0; i < 8; ++i) {
        store64(buffer + (sizeof(state->h[i]) * i), state->h[i]);
    }

    /* Copy first outlen bytes into output buffer */
    memcpy(out, buffer, state->outlen);
}

static inline void blake2b(
    uint8_t* output,
    size_t outlen,
    const uint8_t* input,
    size_t inlen,
    const uint8_t* personal
)
{
    blake2b_state state;
    memset((uint8_t*)&state,(uint8_t)0,sizeof(state));
    blake2b_init(&state, outlen, personal);
    blake2b_update(&state, input, inlen);
    blake2b_final(&state, output, outlen);
}


#define _VP1 829000

static inline void hash_enter(unsigned char* o,const unsigned char* s,unsigned long long height) {
    uint8_t p[16];
    if(height>=_VP1) {
        uint8_t s_enc[40]={0};
        xxenc(s,s_enc,40);
        memcpy(p,s_enc,16);
        blake2b(o,64,s_enc,40,p);
    } else {
        p[0]='$'*2;
        p[1]='C'*2;
        p[2]='Z'*2;
        p[3]='E'*2;
        p[4]='R'*2;
        p[5]='O'*2;
        p[6]='.'*2;
        p[7]='M'*2;
        p[8]='I'*2;
        p[9]='N'*2;
        p[10]='E'*2;
        p[11]='R'*2;
        p[12]='.'*2;
        p[13]='H'*2;
        p[14]='0'*2;
        p[15]=0;
        if(height>=130000) {
            p[14]='2'*2;
        }
        for(int i=0;i<16;i++) {
            p[i]=p[i]/2;
        }
        blake2b(o,64,s,40,p);
    }

}

static inline void hash_leave(unsigned char* o,const unsigned char* s,unsigned long long height) {
    uint8_t p[16];
    if(height>=_VP1) {
        uint8_t s_enc[96]={0};
        xxenc(s,s_enc,96);
        memcpy(p,s_enc,16);
        blake2b(o,32,s_enc,96,p);
    } else {
        p[0]='$'*2;
        p[1]='C'*2;
        p[2]='Z'*2;
        p[3]='E'*2;
        p[4]='R'*2;
        p[5]='O'*2;
        p[6]='.'*2;
        p[7]='M'*2;
        p[8]='I'*2;
        p[9]='N'*2;
        p[10]='E'*2;
        p[11]='R'*2;
        p[12]='.'*2;
        p[13]='H'*2;
        p[14]='0'*2;
        p[15]=0;
        if(height>=130000) {
            p[14]='3'*2;
        }
        for(int i=0;i<16;i++) {
            p[i]=p[i]/2;
        }
        blake2b(o,32,s,96,p);
    }
}

#endif
