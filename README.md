# mysql-diff
Little utility to help summarize one or more mysql schemas

# Usage

## Print JSON Summary

```shell
go build .
./mysql-diff -conn "label=srv1 addr=127.0.0.1:3306 user=root pass=great_password db=db1,label=srv2 addr=127.0.0.1:3307 user=root pass=great_password2 db=db2" summary > summary.json
```

You may also add the `-pretty` flag to the end to produced formatted JSON.

## Print Table Diff to Console

```shell
go build .
./mysql-diff -conn "label=srv1 addr=127.0.0.1:3306 user=root pass=great_password db=db1,label=srv2 addr=127.0.0.1:3307 user=root pass=great_password2 db=db2" diff
```
