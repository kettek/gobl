`init` is a command for initializing an existing or empty directory as a gobl project.

## Usage

```shell
init <directory> <main cmd> [watchdir1 watchdir2 ...]
```

## Example
```shell
init myproject cmd/mycmd assets/ templates/
```

```shell
go run github.com/kettek/gobl/cmd/init myproject cmd/mycmd assets/ templates/
```
