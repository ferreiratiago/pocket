## client Keys Import

Imports a key from a string or from a file

### Synopsis

Imports [privateKeyString] or from [--input_file] into the keybase, provided it is in the form of [--import_format]

```
client Keys Import [privateKeyString] [--input_file] [--import_format] [flags]
```

### Options

```
  -h, --help                   help for Import
      --hint string            hint for the passphrase of the private key
      --import_format string   import the private key from the specified format (default "raw")
      --input_file string      input file to read data from
      --keybase string         keybase type used by the cmd, options are: file, vault (default "file")
      --pwd string             passphrase used by the cmd, non empty usage bypass interactive prompt
      --vault-addr string      Vault address used by the cmd. Defaults to https://127.0.0.1:8200 or VAULT_ADDR env var
      --vault-mount string     Vault mount path used by the cmd. Defaults to secret
      --vault-token string     Vault token used by the cmd. Defaults to VAULT_TOKEN env var
```

### Options inherited from parent commands

```
      --config string           Path to config
      --data_dir string         Path to store pocket related data (keybase etc.) (default "/home/bigboss/.pocket")
      --non_interactive         if true skips the interactive prompts wherever possible (useful for scripting & automation)
      --remote_cli_url string   takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port) (default "http://localhost:50832")
```

### SEE ALSO

* [client Keys](client_Keys.md)	 - Key specific commands

###### Auto generated by spf13/cobra on 19-Mar-2023
