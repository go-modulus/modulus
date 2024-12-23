## Getting Started

This is a guide to help you get started with the project. It will walk you through the steps to get the project up and running on your local machine.

### Installation
First, you need to install the Modulus CLI tool. You can do this by running the following command:

```bash
go install github.com/go-modulus/modulus/cmd/mtools@latest
```

Next, you need to initialize a new project. You can do this by running the following command:

```bash
mtools init --path=./testproj --name=testproj
```

If the `init` command runs without parameters it will prompt you to enter a name for your project. You can enter any name you like, but for this guide, we will use `testproj`.

### Adding Modules
Once you have initialized your project, you can add modules to it. Modules are reusable components that provide functionality to your project. To add a module, run the following command:

```bash
mtools module install --proj-path=./testproj -m "pgx"
```

or
    
```bash 
mtools module install --proj-path=./testproj
```

if you want to select the modules from the list.

Or even
    
```bash
cd testproj
mtools module install
```

if you want to install the modules in the current directory.