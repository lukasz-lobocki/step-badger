# step-badger

## Build

```bash
goreleaser build --clean
```

## Typical release workflow

```bash
git add --update
```

```bash
git commit -m "fix: change"
```

```bash
git tag "$(svu next)"
git push --tags
goreleaser release --clean
```

## Cookiecutter initiation

```bash
cookiecutter \
  ssh://git@github.com/lukasz-lobocki/go-cookiecutter.git \
  package_name="step-badger"
```

### was run with following variables

- package_name: **`step-badger`**;
package_short_description: `Exporting certificate data out of the badger database of step-ca.`

- package_version: `0.1.0`

- author_name: `Lukasz Lobocki`;
open_source_license: `CC0 v1.0 Universal`

- __package_slug: `step-badger`

### on

`2024-07-10 09:10:09 +0200`
