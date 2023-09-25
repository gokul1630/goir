# goir
zero configs live reload program for go projects

 <img width="891" alt="Screenshot 2023-09-25 at 11 42 34 PM" src="https://github.com/gokul1630/goir/assets/43679827/3d7d2be2-64e3-4342-b505-42bbb53ddbe3">


## install
```
go install github.com/gokul1630/goir@latest
```

## Setup your `.bashrc` or `.zshrc` on your Linux/macOS
```
export PATH="$(go env GOPATH)/bin:$PATH"
```

### If you're running on Windows add the `go/bin` path to your `PATH` variable in Environmental variables


## Config for customization
> Note: Config is optional only. You can run goir without any config. If you want to customize, below are the available options.

```
{
	"output": "main",
	"buildArgs": [""],
	"runArgs": [""],
	"excludedPaths": ["tmp"],
	"tmp_dir": "tmp"
}

```

## TODO
- kill child processes
- refactor code
- more customization

## Thanks to
- [cosmtrek](https://github.com/cosmtrek/air)
- [radovskyb](https://github.com/radovskyb/watcher)