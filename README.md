## Disclaimer
This is an experimental build and is currently under active development. Please be advised of the following:

- This code currently is not audited by an external security auditor, use it at your own risk
- The code **has not been subjected to thorough review** by engineers at the Electric Coin Company or anywhere else
- We **are actively changing** the codebase and adding features where/when needed on multiple forks

🔒 Security Warnings

The Lightwalletd Server is experimental and a work in progress. Use it at your own risk. Developers should familiarize themselves with the [wallet app threat model](https://zcash.readthedocs.io/en/latest/rtd_pages/wallet_threat_model.html), since it contains important information about the security and privacy limitations of light wallets that use Lightwalletd.

---

# Overview

[lightwalletd](https://github.com/Asherda/lightwalletd) is a backend service that provides a bandwidth-efficient interface to the VerusCoin blockchain. Currently, lightwalletd supports the Sapling protocol version as its primary concern. The intended purpose of lightwalletd is to support the development of mobile-friendly shielded light wallets.

The VerusCoin developers ported lightwalletd to the VerusCoin VRSC chain. This version uses verusd rather than zcashd and implements the new VerusCoin hashing algorithms, up to and including V2b2. 

Lightwalletd has not yet undergone audits or been subject to rigorous testing. It lacks some affordances necessary for production-level reliability. We do not recommend using it to handle customer funds at this time (April 2020).
## Documentation
Documentation for lightwalletd clients (the gRPC interface) is in `docs/rtd/index.html`. The current version of this file corresponds to the two `.proto` files; if you change these files, please regenerate the documentation by running `make doc`, which requires docker to be installed. 
## Startup
On initial startup lightwalletd starts loading all the blocks starting from block 1. As they are loaded they are added to a levelDB name/value DB store (in a db subdirectory below the --data-dir value you input on the commmand line) for fast scalable access. Records are compactBlocks and they are stored by block height.

Once all the blocks are added - well over 2 million - lightwalletd continues waiting for more blocks, adding them as they occur. If lightwalletd is stopped then restarted, it picks up where it left off, catching up rapidly and then continues ingesting new blocks as they occur.

Once started lightwsalletd serves data via the GRPC endpoint. While the chain is being scanned in the first time performance is many times slower, but once the blockchain ingestor has caught up the overhead drops to a small amount minute and the GRPC endpoint can easily serve 25K requests per second on my dev box running both lightwalletd and verusd (plus browsers and tools etc.)

Latency is around 4ms mostly at up to 100 or so parrallel request pn my machine. With more than 100 requests the throughput slowly increasess but the latency gets large/poor pretty quickly, 500 is slow and 1,000 is very slow.
# Local/Developer docker-compose Usage
Note: when using docker, map the data directory (the one you put your DB on when you specified the --data-dir inside the container) via CLI to a location on your file system, so that the data persists even if the containers are destroyed. This avoids reloading everything every time.

Reloading isn't too horrible, I get about 5K per 4 sedonds, 75K per minute, a bit less than 15 miuntes to load.
[docs/docker-compose-setup.md](./docs/docker-compose-setup.md)
# Local/Developer Usage
Added leveldb support for storing local chain and tx data, replacing the flat file with simple indexing approach. It's included automatically and uses the normal command line options so no change should be needed, existing configurations will continue working.

## Testing
Fixed the unit tests so they all pass. Removed a couple but mostly got them repaired.

Simply run make test to run all the tests:
```
~/levelDB/lightwalletd$ make test
go test -v ./...
# github.com/Asherda/Go-VerusHash
verushash.cxx: In member function ‘void Verushash::initialize()’:
verushash.cxx:21:20: warning: ignoring return value of ‘int sodium_init()’, declared with attribute warn_unused_result [-Wunused-result]
         sodium_init();
         ~~~~~~~~~~~^~
=== RUN   TestHashV2b2
Got the correct v2b2 hash for block 1053660
Got the correct v2b2 hash for block 1053661!
--- PASS: TestHashV2b2 (0.00s)

<deleted lots of passing test results>

=== RUN   TestString_read
--- PASS: TestString_read (0.00s)
PASS
ok  	github.com/Asherda/lightwalletd/walletrpc	(cached)

```
## Code Coverage
If you want to measure unit test coverage of the code run this go test command from the project's root diretory:
```
go test $(go list ./...) -coverprofile coverage.out
ok  	github.com/Asherda/lightwalletd	0.009s	coverage: 0.0% of statements
ok  	github.com/Asherda/lightwalletd/cmd	0.008s	coverage: 37.3% of statements
ok  	github.com/Asherda/lightwalletd/common	0.152s	coverage: 17.4% of statements
ok  	github.com/Asherda/lightwalletd/common/logging	0.006s	coverage: 91.7% of statements
ok  	github.com/Asherda/lightwalletd/frontend	0.200s	coverage: 49.5% of statements
ok  	github.com/Asherda/lightwalletd/parser	0.508s	coverage: 94.6% of statements
ok  	github.com/Asherda/lightwalletd/parser/internal/bytestring	0.003s	coverage: 100.0% of statements
?   	github.com/Asherda/lightwalletd/testtools/zap	[no test files]
ok  	github.com/Asherda/lightwalletd/walletrpc	0.006s	coverage: 3.1% of statements
```
Once that runs you can take a look at coverage while viewing the source code by running:
```
go tool cover -html=coverage.out
```
## Multichain
We can put chains into separate DB files (effectively a DB per chain) or we can use a single DB and add a chain indication to the key.

We'll work that out, for now this gets us live on levelDB which, even with two writes per record (block by height and max height) is about three times as fast as the prior flat file method.
## LevelDB
Switching from the simple two file index & serialzed compact data approach to using levelDB via [goleveldb](https://github.com/syndtr/goleveldb.git).

This gives us better performance at large scale. We support looking blocks up by block height. With 2 writes per block on my dev machine I'm getting about 5.6K blocks every 4 seconds. The old schem was more like 1.7K.
### Progress - Max Block Height
To simplify housekeeping, we record the highest block cached in leveldb. On restart this allows us to resume where we left off and avoid rescanning. We check that the new block's prior_hash matches our recorded hash for the last cached block, so if we get a reorg we will notice and rewind and re-cache the data.

### Corruption Check
Every block record is prepended by an 8 byte checksum for the block that is calculated when we store it. Each time we get a value we redo the checksum to ensure nothing has been corrupted.
### Validation, Reorg and --redownload
Each time we load we scan through the DB to make sure all the block records are present and the compactBlock checksums are correct. If something is not correct we will fix the corruption, or at least try to. The current test suite does not clean u p after itself completely, and creates records at height 2.3m or something like that. After running the tests successfully, when I ran a normal lightwalletd pass it complained abou corruption and worked backward from 2.3M or so all the way down ti 1.15M where the real current records are, then continued from there. I'll take a look at fixing the tests, but for now they have that side effect. They also gave me a good solid test of the "recover from corruption" code and it works fine.

Once the levelDB records have been validated, as each new block shows up we compare it's prevHash field to the has we calculated for the prior block. If they do not match we assume we hit a chain reorg and rewind, getting the prior blocks, checking the hashes and rewinding up to 100 blocks before giving up. Any typical fork will be resolved within a much smaller number of blocks so this is pretty reasonable.

The key value store is idempotent, so as soon as we write a new value for a given height the old one is gone. There's the usual small risk of data loss due to failures since we do not sync on writes, but the system notices corrupted blocks and hash mismatches and automatically corrects for them, so it's pretty resilient.

Reorgs work on the most recent blocks, no more than 100 of them, presumably due to forks. If you want to simply flush the levelDB data, use the --redownload flag.

The --redownload flag on the command line makes lightwalletd flush the levelDB and reload from scratch. Note that we need to delete all the records previously stored for the block before adding a new one. Since we are single threaded and single process, and there is a single record per key type, this works fine. It takes about 8 seconds to delete them all on the current VerusCoin chain, wkich has a bit over 1M records in August 2020.

We have a utility function in the code to flush ranges of blocks in cache.go called flushBlocks(first int, last int)
### Schema
We ingest the blockchain data and store the results. A siplified view of the result:
An array of blocks
- Each block is serialized into a single []byte array called a compactBlock
- The block contains block details and an array of TXs, each of which is a compactTX. Each TX contains arrays of spends and outputs; all are serialzied into a single array.
- When storing the block, we save it under Bnnnnnnnn where nnnnnnnn is the block height.
- When we store a new "latestBlock" we store the height under the key Icccccccc where cccccccc is the chainID.
## Verusd

You must start a local instance of `verusd`, and its `VRSC.conf` file must include the following entries
(set the user and password strings accordingly):
```
txindex=1
insightexplorer=1
experimentalfeatures=1
rpcuser=xxxxx
rpcpassword=xxxxx
```

verusd can be configured to run `mainnet` or `testnet` (or `regtest`). If you stop `verusd` and restart it on a different network (switch from `testnet` to `mainnet`, for example), you must also stop and restart lightwalletd.

It's necessary to run `verusd --reindex` one time for these options to take effect. This typically takes several hours, and requires more space in the data directory.

Lightwalletd uses the following `verusd` RPCs:
- `getblockchaininfo`
- `getblock`
- `getrawtransaction`
- `getaddresstxids`
- `sendrawtransaction`

We plan on extending it to include identity and token options now that those are available (identity) or becoming available (tokens in may 2020).
## Lightwalletd
Install [Cmake](https://cmake.org/download/)

Install [Boost](https://www.boost.org/)

Install [Go](https://golang.org/dl/#stable) version 1.11 or later. You can see your current version by running `go version`.

Clone the [current repository](https://github.com/Asherda/lightwalletd) into a local directory that is _not_ within any component of
your `$GOPATH` (`$HOME/go` by default), then build the lightwalletd server binary by running `make`.

## To run SERVER

Assuming you used `make` to build the server, here's a typical developer invocation:

```
./lightwalletd --log-file /logs/server.log --grpc-bind-addr 127.0.0.1:18232 --verusd-conf-path VRSC.conf --data-dir .
```
Type `./lightwalletd help` to see the full list of options and arguments.

Note that the --zcash-conf-path option is still listed but it doesn't do anything at the moment.
# Production Usage

Run a local instance of `zcashd` (see above), except do _not_ specify `--no-tls-very-insecure`.
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
./lightwalletd --tls-cert cert.pem --tls-key key.pem --verus-conf-file VRSC.conf --log-file /logs/server.log --grpc-bind-addr 127.0.0.1:18232
```

## Block cache

Lightwalletd caches all blocks from Sapling activation up to the
most recent block, which takes about an hour the first time you run
lightwalletd. During this syncing, lightwalletd is fully available,
but block fetches are slower until the download completes.

After syncing, lightwalletd will start almost immediately,
because the blocks are cached in local files (by default, within
`/var/lib/lightwalletd/db`; you can specify a different location using
the `--data-dir` command-line option).

Lightwalletd checks the consistency of these files at startup and during
operation as these files may be damaged by, for example, an unclean shutdown.
If the server detects corruption, it will automatically re-downloading blocks
from `verusd` from that height, requiring up to an hour again (no manual
intervention is required). But this should occur rarely.

If lightwalletd detects corruption in these cache files, it will log
a message containing the string `CORRUPTION` and also indicate the
nature of the corruption.

## Darksidewalletd & Testing

Lightwalletd now supports a mode that enables integration testing of itself and
wallets that connect to it. See the [darksidewalletd
docs](docs/darksidewalletd.md) for more information.

# Visual Studio Code
Using Visual Studio Code to run and debug lightwalletd requires setting up the path and command line options in launch.json. Using the "Open Configurations" selection from the Run menu, put the following code in:
```
{
    //    "--verusd-url", "127.0.0.1:27486",
    "version": "0.2.0",
    "configurations": [


        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "/home/virtualsoundnw/levelDB/lightwalletd/main.go",
            "env": {},
            "args": ["--log-file", "/logs/server.log", "--grpc-bind-addr", "localhost:18232", "--verusd-conf-path", "/home/virtualsoundnw/.komodo/VRSC/VRSC.conf", "--data-dir", ".", "--rpc-host", "localhost", "--rpc-port", "27486", "--rpc-user", "verus", "--rpc-password", "nOWBdmihcwPS5xNjkd78HkjnOp0-pQ3h06hjlv0inO-g"]
        }
    ]
}
```
You'll need to correct the paths and dig the user and password out of ~/.komodo/VRSC/VRSC.conf.
# Testing the GRPC server
You can use [grcpurl, a command line utility for hitting GRPC endpoints](https://github.com/fullstorydev/grpcurl/releases) (like curl but specialized) to hit the GRPC endpoint. GRPC allows you to discover the detials interactively:
```
grpcurl --cacert cert.pem  localhost:18232  list
cash.z.wallet.sdk.rpc.CompactTxStreamer
grpc.reflection.v1alpha.ServerReflection
```
If you don't want to bother with certs then use the -insecure option:
```
grpcurl insecure  localhost:18232  list
cash.z.wallet.sdk.rpc.CompactTxStreamer
grpc.reflection.v1alpha.ServerReflection
```
We have a CompactTxServer, so let's look at that usingg the list command again:
```
grpcurl -insecure localhost:18232  list cash.z.wallet.sdk.rpc.CompactTxStreamer
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlock
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlockRange
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLatestBlock
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLightdInfo
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetTaddressBalance
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetTaddressBalanceStream
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetTaddressTxids
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetTransaction
cash.z.wallet.sdk.rpc.CompactTxStreamer.Ping
cash.z.wallet.sdk.rpc.CompactTxStreamer.SendTransaction
```
GetLightdInfo is the simplest as it has no parameters:
```
grpcurl -insecure localhost:18232 cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLightdInfo
{
  "version": "v0.0.0.0-dev",
  "vendor": "ECC LightWalletD",
  "taddrSupport": true,
  "chainName": "main",
  "saplingActivationHeight": "227520",
  "consensusBranchId": "76b809bb",
  "blockHeight": "1153505"
}
```
You can also get the endpoint described like so:
```
grpcurl -insecure localhost:18232 describe cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLightdInfo
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetLightdInfo is a method:
rpc GetLightdInfo ( .cash.z.wallet.sdk.rpc.Empty ) returns ( .cash.z.wallet.sdk.rpc.LightdInfo );
```
Mostly you provide JSOn data with a -d flag, depending on the verb. Check the describe first for assistance, and use describe on parameters and return values too. For GetBlock:
```
grpcurl --cacert cert.pem localhost:18232  describe cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlock
cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlock is a method:
rpc GetBlock ( .cash.z.wallet.sdk.rpc.BlockID ) returns ( .cash.z.wallet.sdk.rpc.CompactBlock );
```
So drilling into what a BlockID (your input) is, we see:
```
grpcurl --cacert cert.pem localhost:18232  describe cash.z.wallet.sdk.rpc.BlockID
cash.z.wallet.sdk.rpc.BlockID is a message:
message BlockID {
  uint64 height = 1;
  bytes hash = 2;
}
```
We can use parameter 1 to request a block like so:
```
grpcurl --cacert cert.pem -d '{"height": 100}' localhost:18232  cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlock
{
  "height": "100",
  "hash": "3qqfTq1mJ5daCLLmAs4X51iwSiGDnyE2xCgJDs8AAAA=",
  "prevHash": "kTvo85dvyYMDcw3oyH2QODjq5vZawQoBro2C8SUAAAA=",
  "time": 1526887503
}
```
If you're copying and pasting the above examples, be careful of the single quotes ' as they sometimes get converted into "more attractive left and right leaning single quotes" to surround things, which breaks GRPC.
# Load/Latency Testing
Using [ghz, a "Simple gRPC benchmarking and load testing tool" ](https://github.com/bojand/ghz/releases) we can check latency and throughout under load, for example here we hammer on Geblock:
```
 ghz  --cacert cert.pem -d '{"height": 10}' -i walletrpc/service.proto,walletrpc/compact_formats -c 100 -n 100000 --call cash.z.wallet.sdk.rpc.CompactTxStreamer.GetBlock localhost:18232

Summary:
  Count:	100000
  Total:	4.08 s
  Slowest:	21.67 ms
  Fastest:	0.23 ms
  Average:	3.96 ms
  Requests/sec:	24519.14

Response time histogram:
  0.227 [1]	|
  2.371 [28853]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  4.516 [29173]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  6.660 [30424]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  8.805 [9288]	|∎∎∎∎∎∎∎∎∎∎∎∎
  10.949 [1898]	|∎∎
  13.094 [302]	|
  15.238 [52]	|
  17.383 [8]	|
  19.527 [0]	|
  21.672 [1]	|

Latency distribution:
  10 % in 0.69 ms 
  25 % in 1.75 ms 
  50 % in 4.17 ms 
  75 % in 5.41 ms 
  90 % in 6.87 ms 
  95 % in 7.83 ms 
  99 % in 9.81 ms 

Status code distribution:
  [OK]   100000 responses   
```


# Pull Requests

We welcome pull requests! We like to keep our Go code neatly formatted in a standard way,
which the standard tool [gofmt](https://golang.org/cmd/gofmt/) can do. Also, run golint
prior to checkin and keep things clean.

Our current PR template asks for a design document link from the PR description and a
test plan added as a comment. If no design is needed (i.e. a README update or depenency
version update) then don't check off the box, explain the exception. Ditto the test plan.

 Please consider adding the following to the
file `.git/hooks/pre-commit` in your clone:

```
#!/bin/sh

modified_go_files=$(git diff --cached --name-only -- '*.go')
if test "$modified_go_files"
then
    need_formatting=$(gofmt -l $modified_go_files)
    if test "$need_formatting"
    then
        echo files need formatting (then don't forget to git add):
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
