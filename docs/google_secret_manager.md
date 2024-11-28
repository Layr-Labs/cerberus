## Using Google Secret Manager as a backend for cerberus
You can use Google Secret Manager as a backend for cerberus. To use Google Secret Manager as a backend, you need to set the `STORAGE_TYPE` environment variable to `google-secrets-manager`. 
All the public keys are stored in `cerberus<pub-key-hex>` format. They will also have a label with key as `project` and value as `cerberus`.

### Environment variables
You will need to set the `GCP_PROJECT_ID` environment variable to `environment`. Make sure you have the necessary permissions to access the secrets.

Example
```bash
cerberus \
  --storage-type google-secrets-manager \
  --gcp-project-id my-project
```