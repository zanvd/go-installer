Go installer is a simple installer for the Linux Go binary:

1. downloads the specified version
2. removes the current one
3. installs the new one
4. cleans up the downloaded file

```shell
go build .
sudo ./go-installer
rm go-installer
```

Please note that you do need an initial working version of Go.
