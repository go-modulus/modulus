# Deployment

This section will guide you through the deployment process of your GraphQL server.
We will show you how ro work with env variables, how to build and run the server, and how to deploy it to the Digital Ocean App platform.

## Environment Variables

Configuration of the application is done via environment variables.
You can find the list of the environment variables in the `.env` file. It is composed by the installation of modules.
For example, you run `mtools module install  -m="gqlgen"`. The installer reads the manifest of the module. It gets the variables 

```json
  "envVars": [
      {
        "key": "GQL_API_URL",
        "value": "/graphql",
        "comment": ""
      },
...
```

After that the installer creates or reads the `.env` file and adds the new variables with default values to it. 
All existing variables are not changed.

You can change the values of the variables in the `.env` file. Add all the necessary variables from your local modules to the `.env` file before running the server.

We propose to add the `.env` file to the version control system. It will help you to keep the configuration of the server in one place and share it with your team.
If you want to add local variables for the local development create the `.env.local` file. It will override the default configuration. Don't add the `.env.local` file to the version control system.

For each new environment create the new environment file. For example, `.env.prod`, `.env.dev`, `.env.test`. `.env.test` - is the default file for the tests. It is used by the test runner.
Add only the variables that are different from the default configuration to the environment file.

You have the next levels of the environment configuration priority:
- `.env` - the default configuration
- `.env.<environment name>` - overrides the default configuration and runs only if the program is run with env variable `APP_ENV=<environment name>`
- `.env.local` - overrides the default configuration. It works only if the program is run without the `APP_ENV` variable.
- real environment variables - overrides all configurations


## Encrypt secrets

You can encrypt the secrets in the `.env` file. It is useful when you want to store the `.env` file in the version control system.
You can add the encrypted `.env.dev` and `.env.prod` files to the version control system and use them on the servers to run the application.

We propose to use `dotenvx` to encrypt the secrets. You can find the [installation instructions here](https://github.com/dotenvx/dotenvx?tab=readme-ov-file).

Install it, for example, for Mac running the following command:

```shell
brew install dotenvx/brew/dotenvx
```

After that, you can encrypt the secrets for the dev environment running the following command:

```shell
dotenvx set env_name env_value -f .env.dev
```

Try to encrypt the `PGX_DSN` variable for the dev environment:

```shell
dotenvx set PGX_DSN "postgres://postgres:foobar@localhost:5432/test-dev?sslmode=disable"  -f .env.dev
```

It will add the encrypted value to the `.env.dev` file. You can add the `.env.dev` file to the version control system.
To do this change the `.gitignore` file adding the `!.env.dev` line.

It adds the encrypted value to the `.env.dev` file with a public key that can be used by your teammates to encrypt the new values without opening the existent secrets.

Also, the file `.env.keys` is created. It contains the private key. You should not add it to the version control system.
This key should be used only by the CI/CD system to decrypt the secrets. You can add the key to the CI/CD system as a secret. 
Give this key only to the people who are eligible to work with secrets. After that they can read any secret from the `.env.dev` file and using a command:

```shell
dotenvx get PGX_DSN -f .env.dev
```

Or decrypt the whole file:

```shell
 dotenvx decrypt -f .env.dev
```

If you have a private key in the `.env.keys` file you can run your application with the encrypted secrets. 

```shell
dotenvx run -f .env.dev -- ./bin/console serve
```

You have to add the `-f .env.dev` parameter to make dotenvx read both the encrypted and decrypted secrets from the `.env.dev` file. And define the environment variables to run the application with them.


## Deployment to the Digital Ocean App Platform

The Digital Ocean App Platform is a platform as a service (PaaS) that allows you to deploy your applications in the cloud without managing the infrastructure.
It is a good choice for the small and medium-sized projects. It is easy to use and has a low cost tier. For example, create a new application with 512MB of RAM and 1 CPU for $5 per month.

After that, follow their guide to deploy the application using Github actions: https://docs.digitalocean.com/products/app-platform/how-to/deploy-from-github-actions.

Create DO_API_KEY variable in the Github repository secrets with the Digital Ocean API key. It is used by the Github action to deploy the application.
Create the DOTENV_PRIVATE_KEY_DEV environment variable in the App platform on the settings page of the application. It is used to decrypt the secrets in the `.env.dev` file.