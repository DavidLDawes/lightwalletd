# Disclaimer
This is an alpha build and is currently under active development. Please be advised of the following:

- This code currently is not audited by an external security auditor, use it at your own risk
- The code **has not been subjected to thorough review** by engineers at the Electric Coin Company
- The ZCash version was forked to ceate this VerusHash version
- We **are actively changing** the codebase and adding features where/when needed

🔒 Security Warnings

The Lightwalletd Server is experimental and a work in progress. Use it at your own risk.

---

# Overview

[lightwalletd](https://github.com/asherda/lightwalletd) is a backend service that provides a bandwidth-efficient interface to the Zcash blockchain. Currently, lightwalletd supports the Sapling protocol version as its primary concern. The intended purpose of lightwalletd is to support the development of mobile-friendly shielded light wallets.

lightwalletd is a backend service that provides a bandwidth-efficient interface to the Zcash blockchain for mobile and other wallets, such as [Zecwallet](https://github.com/adityapk00/zecwallet-lite-lib).

Lightwalletd has not yet undergone audits or been subject to rigorous testing. It lacks some affordances necessary for production-level reliability. We do not recommend using it to handle customer funds at this time (October 2019).

To view status of [CI pipeline](https://gitlab.com/mdr0id/lightwalletd/pipelines)

To view detailed [Codecov](https://codecov.io/gh/zcash/lightwalletd) report

Documentation for lightwalletd clients (the gRPC interface) is in `docs/rtd/index.html`. The current version of this file corresponds to the two `.proto` files; if you change these files, please regenerate the documentation by running `make doc`, which requires docker to be installed. 
# VerusCoin support
Using swig to get to the C++ VerusCoin hash implementations.

You can generate the verushash.go and verushash_wrap.cxx files from the verushash/verushash.i and verushash/verushash.cxx via this swig command: 
```
swig -go  -intgosize 64 -c++ -cgo -gccgo -Wall -v parser/verushash/verushash.i
```
Once that has completed a simple make comand should assemble everything.
```
make
```

lightwalletd requires the veruslib.so and libboost_system.so libraries. See parser/verushash/verushash.i for the cgo defintiions used to set include directories at compile time and lib directories and libs at link time. You mileage may well very so check that if you have lots of unresolved externals.
## verusd
lightwalletd uses the rpc interface of verusd, the VerusCoin daemon, to get block information for the ingestor and clients and to take actions based on the frontend API requests.

Load verusd - either using the VerusCli or VerusDesktop depending on your preferences - before starting the lightwalletd service.

Once you've got verusd running, check that it has loaded the Verus chain. The verus program in the VerusCoin cli (or the same program in the VerusCoin desktop) is used to request data and take actions using the VerusCoin RPC. A simple request for the current block count makes a good check on the health and status of verusd:
```
./verus getblockcount
``` 
If verusd is not ready yet then you will need to wait until it finishes loading the block chain. If it is not running then get it  running, lightwalletd can only run on old cached information if verusd is not available.
## lightwallet service
verusd is runnig properly and responding correctly to verus RPC requests, you generated fresh swig code and make worked, time to run the service.
```
./server --conf-file ~/.komodo/VRSC/VRSC.conf --log-file /logs/server.log --bind-addr 127.0.0.1:18232
```
Production services will need to deal with certs for SSL and DNS and so on.
## grpcurl
You can test the compact TX streamer service using gpcurl. gcpurl has nice features like listing available methods. Install it using go:
```
go get github.com/fullstorydev/grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl
```

With the service running you can get a list of methods using grpcurl, which shows we provide the compact TX streamer and server reflection:
```
grpcurl  -insecure 127.0.0.1:18232  list
cash.z.wallet.sdk.rpc.CompactTxStreamer
grpc.reflection.v1alpha.ServerReflection
```
Focussing on the TX streamer 

grpcurl  -insecure 127.0.0.1:18232  list cash.z.wallet.sdk.rpc.CompactTxStreamer
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetAddressTxids
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlock
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlockRange
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetIdentity
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLatestBlock
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLightdInfo
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetTransaction
cash.z.wallet.sdk.rpc.CompactTxStreamer.RecoverIdentity
cash.z.wallet.sdk.rpc.CompactTxStreamer.RegisterIdentity
cash.z.wallet.sdk.rpc.CompactTxStreamer.RegisterNameCommitment
cash.z.wallet.sdk.rpc.CompactTxStreamer.RevokeIdentity
cash.z.wallet.sdk.rpc.CompactTxStreamer.SendTransaction
cash.z.wallet.sdk.rpc.CompactTxStreamer.UpdateIdentity
cash.z.wallet.sdk.rpc.CompactTxStreamer.VerifyMessage

```
grpcurl -d '{"height":800199}' -insecure 127.0.0.1:18232  cash.z.wallet.sdk.rpc.CompactTxStreamer/GetBlock
```

## Validating VerusCoin hashes


# Local/Developer docker-compose Usage

[docs/docker-compose-setup.md](./docs/docker-compose-setup.md)

# Local/Developer Usage

First, ensure [Go >= 1.11](https://golang.org/dl/#stable) is installed. Once your go environment is setup correctly, you can build/run the below components.

To build the server, run `make`.

This will build the server binary, where you can use the below commands to configure how it runs.

## To run SERVER

Assuming you used `make` to build SERVER:

```
./server --no-tls-very-insecure=true --conf-file /home/.komodo/VRSC/VRSC.conf --zconf-file /home/zcash/.zcash/zcash.conf --log-file /logs/server.log --bind-addr 127.0.0.1:18232
```

# Production Usage

Ensure [Go >= 1.11](https://golang.org/dl/#stable) is installed.

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
./server --tls-cert cert.pem --tls-key key.pem --conf-file /home/.komodo/VRSC/VRSC.conf --zconf-file /home/zcash/.zcash/zcash.conf --log-file /logs/server.log --bind-addr 127.0.0.1:18232
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
`gofmt` command as indicated and rerun the `git commit` command.
