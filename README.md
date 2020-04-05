
[![pipeline status](https://gitlab.com/zcash/lightwalletd/badges/master/pipeline.svg)](https://gitlab.com/asherda/lightwalletd/commits/master)
[![coverage report](https://gitlab.com/zcash/lightwalletd/badges/master/coverage.svg)](https://gitlab.com/asherda/lightwalletd/commits/master)

# Disclaimer
This is an alpha build and is currently under active development. Please be advised of the following:

- This code currently is not audited by an external security auditor, use it at your own risk
- The code **has not been subjected to thorough review** by engineers at the Electric Coin Company
- The code **has not been subjected to thorough review** by the VerusCoin developers
- We **are actively changing** the codebase and adding features where/when needed

🔒 Security Warnings

The Lightwalletd Server is experimental and a work in progress. Use it at your own risk.

---

# Overview

[lightwalletd](https://github.com/asherda/lightwalletd) is a backend service that provides a bandwidth-efficient interface to the VerusCoin blockchain. Currently, lightwalletd supports the Sapling protocol version and the VerusCoin ifdentity methds wil be added shortly. The intended purpose of lightwalletd is to support the development of mobile-friendly shielded light wallets.

Zcash has a wallet reference here [Zecwallet](https://github.com/adityapk00/zecwallet-lite-lib).

Lightwalletd has not yet undergone audits or been subject to rigorous testing. It lacks some affordances necessary for production-level reliability. We do not recommend using it to handle customer funds at this time (February 2020).

To view status of [CI pipeline](https://gitlab.com/asherda/lightwalletd/pipelines)

Code coverage reporting is not available

Documentation for lightwalletd clients (the gRPC interface) is in `docs/rtd/index.html`. The current version of this file corresponds to the two `.proto` files; if you change these files, please regenerate the documentation by running `make doc`, which requires docker to be installed. A generic make will check the dependencies and recreate any documents that are older than their source files.
# Local/Developer docker-compose Usage

[docs/docker-compose-setup.md](./docs/docker-compose-setup.md)

# Local/Developer Usage

## verusd

You must start a local instance of `verusd`, and its `.komodo/VRSC/VRSC.conf` file must include the following entries:
```
txindex=1
```

It may be necessary to run `verusd --reindex` one time if you modify the txindex option for it to take effect. This typically takes several hours.

Lightwalletd uses the following `verusd` RPCs:
- `getblockchaininfo`
- `getblock`
- `getrawtransaction`
- `getaddresstxids`
- `sendrawtransaction`

## Lightwalletd

First, install [Go](https://golang.org/dl/#stable) version 1.11 or later. You can see your current version by running `go version`.

### Generate C++ Wrappers
To create the C++ wrapper code needed to access the VerusHash code use swig:
```
swig -go  -intgosize 64 -c++ -cgo -gccgo -Wall -v parser/verushash/verushash.i
```
This generates the wrappers in the parser/verushash directory

# Build lightwalletd
To build the lightwalletd server, run `make`.

This will build the lightwalletd binary, where you can use the below commands to configure how it runs.

## Run lightwalletd

Assuming you used `make` to build lightwalletd, and you are testing so we will disable TLS. Disabling TLS should only be done when testing locally, and a lightwalletd endpoint should never be publicly exposed with the --no-tls-very-insecure option. The directories below may need adjustment.
```
./lightwalletd --verus-conf-path ~/.komodo/VRSC/VRSC.conf --log-file /logs/server.log --bind-addr 127.0.0.1:18232 --no-tls-very-insecure 
```
## Verify lightwalletd Is Ingesting verusd Blocks
Use tail -f on the log file you set in the lightwalletd command (so it would be `tail -f /logs/server.log` for the example above) and you should see some startup information then a list of ingested blocks; it reports every 100 blocks while initializng and should look like this:

```
{"app":"frontend-grpc","level":"info","msg":"Lightwalletd starting version v0.3.0","time":"2020-04-04T17:31:48-07:00"}
{"app":"frontend-grpc","level":"warning","msg":"Certificate and key not provided, generating self signed values","time":"2020-04-04T17:31:48-07:00"}
{"app":"frontend-grpc","level":"info","msg":"Got sapling height 227520 block height 955243 chain main branchID 76b809bb","time":"2020-04-04T17:31:48-07:00"}
{"app":"frontend-grpc","level":"info","msg":"Starting gRPC server on 127.0.0.1:18232","time":"2020-04-04T17:31:48-07:00"}
{"app":"frontend-grpc","level":"info","msg":"Ingestor adding block to cache: 227600","time":"2020-04-04T17:31:48-07:00"}
{"app":"frontend-grpc","level":"info","msg":"Ingestor adding block to cache: 227700","time":"2020-04-04T17:31:48-07:00"}
```
Time and date vary on usage of course. It should grind on through up close to a million or even more once enough time passes. Once it no longer outputs updates and the last one was within 100 fo the current block height it's caught up.

Check the cache files in the directory you ran ./lightwalletd from, they should look something like this:
```
ls db-main-*
db-main-blocks  db-main-lengths
```
As of early April 2020 these files are a bit over 130MB and a bit under 3MB respectively. The sizes will coninue to grow as blocks and transactions get added to the chain and ingested, of course.
## Verify lightwalletd CompactTxStreamer Is Working
Install grpcurl

Try grpcurl list. Note the required -plaintext flag to switch of SSL since we ran lightwalletd with the --no-tls-very-insecure flag:
```
grpcurl -plaintext 127.0.0.1:18232 list 
cash.z.wallet.sdk.rpc.CompactTxStreamer
grpc.reflection.v1alpha.ServerReflection
```
In any production situation you'd remove the -plaintext here and provide proper certs to lightwalletd, using the command line to set that up properly.

The lightwalletd provides a compact TX streamer and a aserver reflection protoset. Focus here is on the compact TX streamer, so digging into that:
```
grpcurl -plaintext 127.0.0.1:18232 list cash.z.wallet.sdk.rpc.CompactTxStreamer
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetAddressTxids
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlock
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlockRange
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLatestBlock
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLightdInfo
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetTransaction
cash.z.wallet.sdk.rpc.CompactTxStreamer.Ping
cash.z.wallet.sdk.rpc.CompactTxStreamer.SendTransaction
```

How to use GetBlock:
```
grpcurl -d '{"height":260000}' -plaintext 127.0.0.1:18232 cash.z.wallet.sdk.rpc.CompactTxStreamer/GetBlock
{
  "height": "260000",
  "hash": "Y54zD2QQby5k78a9Lh8h76yl9E0mYoLwxS0AAAAAAAA=",
  "prevHash": "N7/Fe9CN3/JY6eK2KSZNt+oOKzvMtaOilh0AAAAAAAA=",
  "time": 1542965524,
  "vtx": [
    {
      "index": "2",
      "hash": "otp9InYO+HjnjR+t9+crk5lzf0i7QfIMmIk48eEFY+E=",
      "outputs": [
        {
          "cmu": "SE9ZYVIMZpIcv0EkP8708YCIZ50a0c2fc2r7K1+Ygmc=",
          "epk": "0gRIffwb2TmSEp91GRwiE8qvyF7uzXByP8VURvQxc7Y=",
          "ciphertext": "ks/5xaxJwdUZErtmIFSfm1v1uE+aDxQ1WWhBA7plX+Szle4Qu7ccsfXUKLOLdlqkZkUhJw=="
        }
      ]
    },
    {
      "index": "3",
      "hash": "ma4SQ+8DTYjsS5XBDEzJGj31JA/kGU60eKy4A88od+E=",
      "outputs": [
        {
          "cmu": "mD+gtp+Cfq/Ai7ZRH2+nUejDRHkdJOEMmauokIGivC4=",
          "epk": "7S9V/lxvKQrL5KDZl7arRm3MxhVdzNMDiELUiO61AIo=",
          "ciphertext": "qYHqhisJWFAOey6UBuDKYkdXkH6DOAlx+VcQBL6cVFV5DtvZXCHThctOPchYd0Ob8ajyVQ=="
        }
      ]
    }
  ]
}
```
# Production Usage

Run a local instance of `verud` (see above).
Ensure [Go](https://golang.org/dl/#stable) version 1.11 or later is installed.

**x509 Certificates**
You will need to supply an x509 certificate that connecting clients will have good reason to trust (hint: do not use a self-signed one, our SDK will reject those unless you distribute them to the client out-of-band). We suggest that you be sure to buy a reputable one from a supplier that uses a modern hashing algorithm (NOT md5 or sha1) and that uses Certificate Transparency (OID 1.3.6.1.4.1.11129.2.4.2 will be present in the certificate).

To check a given certificate's (cert.pem) hashing algorithm:
```
openssl x509 -text -in certificate.crt | grep "Signature Algorithm"
```

To check if a given certificate (cert.pem) contains a Certificate Transparency OID:
```
echo "1.3.6.1.4.1.11129.2.4.2 certTransparency Certificate Transparency" > oid.txt
openssl asn1parse -in cert.pem -oid ./oid.txt | grep 'Certificate Transparency'
```

To use Let's Encrypt to generate a free certificate for your frontend, one method is to:
1) Install certbot
2) Open port 80 to your host
3) Point some forward dns to that host (some.forward.dns.com)
4) Run
```
certbot certonly --standalone --preferred-challenges http -d some.forward.dns.com
```
5) Pass the resulting certificate and key to frontend using the -tls-cert and -tls-key options.

## To run production SERVER

Example using server binary built from Makefile:

```
./lightwalletd --tls-cert cert.pem --tls-key key.pem --verus-conf-path ~/.komodo/VRSC/VRSC.conf --log-file /logs/server.log --bind-addr 127.0.0.1:18232
```

# Pull Requests

We welcome pull requests! We like to keep our Go code neatly formatted in a standard way,
which the standard tool [gofmt](https://golang.org/cmd/gofmt/) can do. Please consider
adding the following to the file `.git/hooks/pre-commit` in your clone:

```
#!/bin/sh

modified_go_files=$(git diff --cached --name-only -- '*.go')
if test "$modified_go_files"
then
    need_formatting=$(gofmt -l $modified_go_files)
    if test "$need_formatting"
    then
        echo files need formatting:
        echo gofmt -w $need_formatting
        exit 1
    fi
fi
```

You'll also need to make this file executable:

```
$ chmod +x .git/hooks/pre-commit
```

Doing this will prevent commits that break the standard formatting. Simply run the
`gofmt` command as indicated and rerun the `git add` and `git commit` commands.
