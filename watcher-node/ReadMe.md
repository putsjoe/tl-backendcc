# Watcher Node

Store an updated list of the changes which happen in a specified folder, and update an aggregator service with the changes

## Endpoints

`GET http://localhost:4000/files`

Response:
```
{
    "files" [
        {
            "name: "file.txt"
        },
        {
            "name": "anotherfile.txt"
        }
    ]
}
```

# To run:

system requirements: Golang

`make build` then run `./watcher-node -dir=${./yourwatched/directory}`

The default port is `4000`. You can optionally pass in a port using the `-p` flag. 

To test run `make test`
