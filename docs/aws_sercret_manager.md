## Using AWS Secret Manager as a backend for cerberus
You can use AWS Secret Manager as a backend for cerberus. To use AWS Secret Manager as a backend, you need to set the `STORAGE_TYPE` environment variable to `aws-secrets-manager`. 
All the public keys are stored in `cerberus/<pub-key-hex>` format.

You have two options for authenticating with AWS Secret Manager:
### Environment variables
You will need to set the `AWS_AUTHENTICATION_MODE` environment variable to `environment`. This is the default mode. You will also need to set the `AWS_REGION`. If you are using a profile, you can set the `AWS_PROFILE` environment variable. If you are using the default profile, you do not need to set the `AWS_PROFILE` environment variable.

Example
```bash
cerberus \
  --storage-type aws-secrets-manager \
  --aws-region us-east-2 \
  --aws-profile SomeProfile
```
### Specified
You will need to set the `AWS_REGION`, `AWS_ACCESS_KEY_ID`, and `AWS_SECRET_ACCESS_KEY` environment variables.

Example
```bash
cerberus \
  --storage-type aws-secrets-manager \
  --aws-region us-east-2 \
  --aws-access-key-id SomeAccessKey \
  --aws-secret-access-key SomeSecretKey
```