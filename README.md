# mysql-diff
Little utility to help produce a diff between two mysql schemas

# Usage

```shell
go build .
./mysql-diff summary -conn "addr=127.0.0.1:3306 user=root pass=great_password db=db1 db=db2" > summary.json
```