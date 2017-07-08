# Deviceio CLI

Deviceio CLI provides command line tools to interact with connected devices. 

# Quickstart

Download one of the CLI binaries for your respective platform from the release page https://github.com/deviceio/cli/releases/latest

You may wish to rename the binary to `deviceio-cli` for clarity.

If you are on linux ensure we add execute permissions to the binary

```
chmod +x deviceio-cli
```

be sure to review the programs help

```
./deviceio-cli --help
```

Next we will configure the default hub profile. Issue the `configure` command and 
fill in the prompts with your hub user credentials

```
./deviceio-cli configure
Hub API Address or Hostname [127.0.0.1]: hub.mydomain.com
Hub API Port [4431]: 443
User ID [admin]: someuser
User Private Key:
User TOTP Secret:
```

Using the hostname of a previous connected agent test the cli by issuing an exec command

```
./deviceio-cli device exec somedevice.mydomain.com whoami
```

Next: