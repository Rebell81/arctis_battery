SteelSeries Arctis 7 battery status tray icon



# windows build
1) install gcc

install choco https://chocolatey.org/install
```
choco install mingw -y
check: gcc -v
```

2) CGO_ENABLED
```
set CGO_ENABLED=1
go run -race .
```

3) reboot
4) If u need, change pid data at `main.go`
   ![img_1.png](img_1.png)
5) build

```go build -ldflags -H=windowsgui```

![img.png](img.png)