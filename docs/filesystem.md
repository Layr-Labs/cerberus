## Using Filesystem as a backend for cerberus
You can use Filesystem as a backend for cerberus. To use Filesystem as a backend, you need to set the `STORAGE_TYPE` environment variable to `filesystem`.

You will need to setup the storage directory where the private keys will be stored. By default, the private keys are stored in the `./data/keystore` directory. You can change this by setting the `KEYSTORE_DIR` environment variable.

Example
```bash
cerberus \
  --storage-type filesystem \
  --keystore-dir /path/to/keystore
```