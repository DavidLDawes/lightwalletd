# Disclaimer
This is an alpha build and is currently under active development. Please be advised of the following:

- This code currently is not audited by an external security auditor, use it at your own risk
- The code **has not been subjected to thorough review** by engineers at the Electric Coin Company
- We **are actively changing** the codebase and adding features where/when needed

The current version always reloads the data each time it starts. Data is loaded starting with block 1. until the most recent block is reached. After that new blocks are added as we get them, until the next restart, then it reloads etc.
🔒 Security Warnings

The Lightwalletd server is experimental software. Use it at your own risk.

---

# Overview

[lightwalletd](https://github.com/Asherda/lightwalletd) is a backend service that provides a bandwidth-efficient interface to the VerusCoin blockchain. Currently, lightwalletd supports the Sapling protocol version as its primary concern. The intended purpose of lightwalletd is to support the development of mobile-friendly shielded light wallets. The VerusCOin developers are porting this to the VerusCoin VRSC chain now. This version uses verusd rather than zcashd, but still has the old zcashd hashing support so it does not work yet. It thinks we are stuck at a reord immediately. Next PR should fix that and get lightwalletd working properly against VerusCoin's VESC chain using verusd.

lightwalletd is a backend service that provides a bandwidth-efficient interface to the Zcash blockchain for mobile and other wallets, such as [Zecwallet](https://github.com/adityapk00/zecwallet-lite-lib).

Lightwalletd has not yet undergone audits or been subject to rigorous testing. It lacks some affordances necessary for production-level reliability. We do not recommend using it to handle customer funds at this time (April 2020).

Documentation for lightwalletd clients (the gRPC interface) is in `docs/rtd/index.html`. The current version of this file corresponds to the two `.proto` files; if you change these files, please regenerate the documentation by running `make doc`, which requires docker to be installed. 
# Local/Developer docker-compose Usage

[docs/docker-compose-setup.md](./docs/docker-compose-setup.md)

# Local/Developer Usage

## Postgres support
This branch introduces storing the VRSC chain data in a PostgreSQL database. Currently it attempts to mimic the data stored in the disk cache Note that the header is not being set correctly yet.

TODO: Fix the header in the SQL DB

The schema uses tables for block, tx, spend and output. The relations and fkeys are as follows:
1. block table has height as an integer index, from 1 to the current block height, and contains block details (hashes, header, time).
2. tx table has transactions related to a block, using the tx(index)plus the block(height) (the fkey to the block) as a multipart key. This allows 0, 1 or many tx records per block(height) and makes it easy to select all txs for a given block. tx also stores the tx(hash) value which is unique to each tx.
3. spend table has nf, plus it's own serial (autoincrementing integer) index as a key, and includes the tx(hash) as a foreign reference from related tx. This allows 0 to many to exist per tx, and allows us to easily fetch all spends for a given tx(hash).
4. output table has it's own serial (autoincrementing integer) index as a key, and includes the tx(hash) as a foreign reference from related tx. This allows 0 to many to exist per tx and allows easy acces to all outputs for a given tx(hash). output also contains cmu, epk and ciphertext binary strings.

TODO: Multichain support - add chain ID to block and tx as an fkey reference to the (yet to be created) chain SQL DB table.

The initial simple implementation expects the DB on localhost:5432 and the schema can be created using SQL with the following commands:
```
CREATE DATABASE vrsc;

CREATE TABLE blocks (
   height INT PRIMARY KEY,
   hash BYTEA UNIQUE NOT NULL,
   prev_hash BYTEA UNIQUE,
   time INT NOT NULL,
   header BYTEA NOT NULL
);

CREATE TABLE tx (
    index BIGINT NOT NULL,
    height INT REFERENCES blocks (height) ON DELETE CASCADE NOT NULL,
    hash BYTEA UNIQUE NOT NULL,
    fee INT,
    PRIMARY KEY(height, index)
);

CREATE TABLE spend (
    index SERIAL PRIMARY KEY,
    tx_hash BYTEA REFERENCES tx (hash) ON DELETE CASCADE NOT NULL,
    nf BYTEA NOT NULL
);

CREATE TABLE output (
    index SERIAL PRIMARY KEY,
    tx_hash BYTEA REFERENCES tx (hash) ON DELETE CASCADE NOT NULL,
    cmu BYTEA NOT NULL,
	epk BYTEA NOT NULL,
	ciphertext BYTEA NOT NULL
);
```
TODO: add schema auto-creation

Currently it has the DB host as localhost and port 5432 hard wired, along with the password and user. Works OK for development (kinda - verusd, postgres and lwd all on the same system during rapid block ingestion is a fairly heavy load) so we can establish the code to write the DB records works.

TODO: add injection of DB details via CLI and/or environment
## Zcashd

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

We plan on extending it to include identity and token options now that those are available (identity) or becoming available (tokens in 2020).
## Lightwalletd
Install [Cmake](https://cmake.org/download/)

Install [Boost](https://www.boost.org/)

Install [Go](https://golang.org/dl/#stable) version 1.11 or later. You can see your current version by running `go version`.

Clone the [current repository](https://github.com/zcash/lightwalletd) into a local directory that is _not_ within any component of
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
from `zcashd` from that height, requiring up to an hour again (no manual
intervention is required). But this should occur rarely.

If lightwalletd detects corruption in these cache files, it will log
a message containing the string `CORRUPTION` and also indicate the
nature of the corruption.

## Darksidewalletd & Testing

Lightwalletd now supports a mode that enables integration testing of itself and
wallets that connect to it. See the [darksidewalletd
docs](docs/darksidewalletd.md) for more information.

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
