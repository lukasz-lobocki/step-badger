# step-badger ![Static](https://img.shields.io/badge/bulaj-biznes-darkorchid?style=for-the-badge&labelColor=darkslategray)

This tool has 3 features:

- display issued [x509 certificates](#step-badger-x509certs) from step-ca badger database.
- display issued [ssh certificates](#step-badger-sshcerts) from step-ca badger database.
- display [content of a given data prefix](#step-badger-dbtable) from step-ca badger database.

## step-badger x509Certs

Export data of x509 certificates.

```bash
step-badger x509Certs PATH [flags]
```

```text
Flags:
  -e, --emit {t|j|m|o}   emit format: table|json|markdown|openssl (default t)
  -t, --time {i|s}       time format: iso|short (default i)
  -s, --sort {s|f}       sort order: start|finish (default f)
  -d, --dnsnames         DNSNames column shown
  -m, --emailaddresses   EmailAddresses column shown
  -i, --ipaddresses      IPAddresses column shown
  -u, --uris             URIs column shown
  -c, --crl              crl column shown
  -p, --provisioner      provisioner column shown
  -v, --valid            valid certificates shown (default true)
  -r, --revoked          revoked certificates shown (default true)
  -x, --expired          expired certificates shown
```

### Example

![alt text](samples/out-x509.png)

## step-badger sshCerts

Export data of ssh certificates.

```bash
step-badger sshCerts PATH [flags]
```

```text
Flags:
  -e, --emit {t|j|m}   emit format: table|json|markdown (default t)
  -t, --time {i|s}     time format: iso|short (default i)
  -s, --sort {s|f}     sort order: start|finish (default f)
  -k, --kid            Key ID column shown
  -v, --valid          valid certificates shown (default true)
  -r, --revoked        revoked certificates shown (default true)
  -x, --expired        expired certificates shown
```

### Example

![alt text](samples/out-ssh.png)

## step-badger dbTable

Export data of a given table.

```bash
step-badger dbTable PATH TABLE
```

> See [this](https://github.com/smallstep/certificates/blob/077f688e2d781fa12fd3d702cfab5b6f989a4391/db/db.go#L18) for table names.

### Example

![alt text](samples/out-dbtable.png)

## Info

See [this](https://smallstep.com/docs/step-ca/certificate-authority-server-production/#enable-active-revocation-on-your-intermediate-ca).

## Build

See [BUILD.md](BUILD.md) file.

## License

`step-badger` was created by Lukasz Lobocki. It is licensed under the terms of the CC0 v1.0 Universal license.

All components used retain their original licenses.

## Credits

Inspired by [github.com/maraino](https://gist.github.com/maraino/4dcb64cb051b17ef6d421892cb4e55a8#file-listcerts-go).

`step-badger` was created with [cookiecutter](https://cookiecutter.readthedocs.io/en/latest/) and [template](https://github.com/lukasz-lobocki/go-cookiecutter).
