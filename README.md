# step-badger ![Static](https://img.shields.io/badge/bulaj-biznes-darkorchid?style=for-the-badge&labelColor=darkslategray)

This tool has 3 features:

- display issued [x509 certificates](#step-badger-x509certs) from step-ca badger database.
- display issued [ssh certificates](#step-badger-sshcerts) from step-ca badger database.
- display [content of a given data prefix](#step-badger-dbtable) from step-ca badger database.

## step-badger x509Certs

Export data of x509 certificates.

```bash
step-badger x509Certs PATH
```

### Example

![alt text](samples/out-x509.png)

## step-badger sshCerts

Export data of ssh certificates.

```bash
step-badger sshCerts PATH
```

### Example

![alt text](samples/out-ssh.png)

## step-badger dbTable

Export data of a given table.

```bash
step-badger dbTable PATH TABLE
```

### Example

![alt text](samples/out-dbtable.png)

## Build

See [BUILD.md](BUILD.md) file.

## License

`step-badger` was created by Lukasz Lobocki. It is licensed under the terms of the CC0 v1.0 Universal license.

All components used retain their original licenses.

## Credits

`step-badger` was created with [cookiecutter](https://cookiecutter.readthedocs.io/en/latest/) and [template](https://github.com/lukasz-lobocki/go-cookiecutter).
