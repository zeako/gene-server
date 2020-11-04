# gene-server

## Running with `Go`
```sh
> export GO111MODULE=on
> git clone git@github.com:zeako/gene-server.git && cd gene-server
> export DNA_FILE_PATH=<dna-file-path>  # dna file location
> go run cmd/server/main.go
```

## Running with `Docker`
```sh
> git clone git@github.com:zeako/gene-server.git && cd gene-server
> docker build . -t dev/gene-server
> docker run --rm -it -p 80:8080 -e DNA_FILE_PATH=/tmp/dna_file --mount type=bind,source=<dna-file-path>,target=/tmp/dna_file dev/gene-server
```
