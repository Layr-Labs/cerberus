# Remote Signer Implementation of cerberus-api
This is a remote signer for BLS signatures on the BN254 curve.


### Installation
#### Quick start
```bash
$ git clone https://github.com/Layr-Labs/cerberus.git
$ cd cerberus
$ make start
```

#### Manual
```bash
git clone https://github.com/Layr-Labs/cerberus.git
cd cerberus
go build -o bin/cerberus cmd/cerberus/main.go
./bin/cerberus 
```

### Usage options
| Options      | Description                                 | Default         |
|--------------|---------------------------------------------|-----------------|
| keystore-dir | Directory to store encrypted keystore files | ./data/keystore |
| grpc-port    | gRPC port for starting signer server        | 50051           |
| log-format   | format of the logs (text, json)             | text            |
| log-level    | debug, info, warn, error                    | info            |
| metrics-port | port to expose prometheus metrics           | 9091            |
| help         | show help                                   |                 |
| version      | show version                                |                 |


### Monitoring
The signer exposes prometheus metrics on the `/metrics` endpoint. You can scrape these metrics using a prometheus server.
There is a grafana dashboard available in the `monitoring` directory. You can import this dashboard into your grafana server to monitor the signer.

### Configuring Server-side TLS (optional)

Server-side TLS support is provided to encrypt traffic between the client and server. This can be enabled by starting the service with `tls-ca-cert` and `tls-server-key` parameters set:

```
cerberus -tls-ca-cert server.crt -tls-server-key server.key
```

The server can then be queried over a secure connection using a gRPC client that supports TLS. For example, using `grpcurl`:

```
grpcurl -cacert ../cerberus/server.crt -d '{"password": "test"}' -import-path . -proto proto/keymanager.proto localhost:50051 keymanager.v1.KeyManager/GenerateKeyPair
```
#### Generating TLS certificates

For local testing purposes, the following commands can be used to generate a server certificate and key.

Create a file named `openssl.cnf` with the following content:

```
[ req ]
default_bits       = 2048
default_md         = sha256
default_keyfile    = server.key
prompt             = no
encrypt_key        = no

distinguished_name = req_distinguished_name
x509_extensions    = v3_req

[ req_distinguished_name ]
C            = US
ST           = California
L            = San Francisco
O            = My Company
OU           = My Division
CN           = localhost

[ v3_req ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = localhost
```

```bash
# Generate the private key
openssl genpkey -algorithm RSA -out server.key

# Generate the certificate signing request (CSR)
openssl req -new -key server.key -out server.csr -config openssl.cnf

# Generate the self-signed certificate with SAN
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt -extensions v3_req -extfile openssl.cnf

```

server.crt and server.key files can then be used to start the server with TLS support.

## Security Bugs
Please report security vulnerabilities to security@eigenlabs.org. Do NOT report security bugs via Github Issues.
