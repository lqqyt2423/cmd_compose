# README

Like docker compose, run multiple process all in one parent process.

- [x] Pipe sub process stdout and stderr to parent process
- [x] If one sub process exit, then all sub process exit
- [x] Prefix every sub log line
- [x] Start sub process in special order and exit in revert order
- [x] Ready status by watch log line through regexp
- [x] Parse config file

```bash
go get github.com/lqqyt2423/cmd_compose
```
