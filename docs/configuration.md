# Configuration

This section guides you through configuring an application built on the Modulus framework.
It covers working with environment variables and running the application using environment variable files.

## Environment Variables

Configuration of the application is done via environment variables.
The `.env` file contains the list of all environment variables and is populated during module installation.
For example, when you run `mtools module install -m="gqlgen"`, the installer reads the module manifest and extracts its declared variables:

```json
  "envVars": [
      {
        "key": "GQL_API_URL",
        "value": "/graphql",
        "comment": ""
      },
...
```

After that, the installer creates or updates the `.env` file, adding the new variables with their default values.
Existing variables are left unchanged.

You can change the values of the variables in the `.env` file. Add all the necessary variables from your local modules to the `.env` file before running the server.

We recommend committing the `.env` file to version control. This keeps the server configuration in one place and makes it easy to share with your team.
For local overrides, create a `.env.local` file — it takes precedence over `.env` but should not be committed to version control.

For each deployment environment, create a dedicated environment file: `.env.prod`, `.env.dev`, `.env.test`, and so on. `.env.test` is the default file used by the test runner.
Only include variables that differ from the defaults.

Environment configuration is applied in the following priority order (highest last):
- `.env` — base defaults
- `.env.<environment>` — overrides defaults; applied when the program is started with `APP_ENV=<environment>`
- `.env.local` — overrides defaults; applied only when `APP_ENV` is not set
- real environment variables — override everything


## Encrypt secrets

You can encrypt the secrets in the `.env` file. It is useful when you want to store the `.env` file in the version control system.
You can add the encrypted `.env.dev` and `.env.prod` files to the version control system and use them on the servers to run the application.

We recommend using `dotenvx` to encrypt secrets. You can find the [installation instructions here](https://github.com/dotenvx/dotenvx?tab=readme-ov-file).

For example, on macOS, install it with:

```shell
brew install dotenvx/brew/dotenvx
```

To encrypt a secret for the dev environment, run:

```shell
dotenvx set env_name env_value -f .env.dev
```

Try to encrypt the `PGX_DSN` variable for the dev environment:

```shell
dotenvx set PGX_DSN "postgres://postgres:foobar@localhost:5432/test-dev?sslmode=disable"  -f .env.dev
```

It will add the encrypted value to the `.env.dev` file. You can add the `.env.dev` file to the version control system.
To do this change the `.gitignore` file adding the `!.env.dev` line.

The encrypted value is written to `.env.dev` alongside a public key. Your teammates can use this public key to encrypt new values without being able to read existing secrets.

A `.env.keys` file is also created containing the private key. DO NOT commit this file to version control.
The private key should only be given to people authorized to access secrets and to the CI/CD system for decryption. Those with the key can read any secret from `.env.dev` using:

```shell
dotenvx get PGX_DSN -f .env.dev
```

Or decrypt the whole file:

```shell
 dotenvx decrypt -f .env.dev
```


## Run the application
On the local environment, if you have the private key in the `.env.keys` file, you can run your application with the encrypted secrets:

```shell
dotenvx run -f .env.dev -- ./bin/console serve
```

The `-f .env.dev` flag tells `dotenvx` to read both encrypted and plain values from `.env.dev`, decrypt them, and expose them as environment variables for the process.

For server deployments, do not copy the `.env.keys` file to the server. Instead, set the `DOTENV_PRIVATE_KEY_<ENVIRONMENT>` environment variable with the private key value — for example, `DOTENV_PRIVATE_KEY_DEV=abcde...` for the dev environment.
