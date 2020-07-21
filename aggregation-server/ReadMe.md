# Aggregation Server

Communicates with Watcher Nodes to retrieve the complete list of files across multiple nodes. Accepts requests from these nodes to update the current list.

## Endpoints

`GET http://localhost:8000/files`

Retrieves sorted list of filenames for all files across connected watcher nodes.

Response:
```
{
    "files" [
        {
            "filename": "anotherfile.txt"
        },
        {
            "filename: "file.txt"
        }
    ]
}
```

`POST http://localhost:8000/hello`

Received periodically from watcher nodes to confirm the active state of the node. JSON body contains instance ID and listen port. Aggregation server should retrieve the watcher node's listen address from the http connection.

Expected form of request body:

```
{
    "instance": "56d1a8de-14a8-403b-b3e7-d49307c63553",
    "port": 4001,
}
```

`POST http://localhost:8000/bye`

Received from the watcher node following a clean shutdown of the node. Indicates that no more updates will come from this node and files from this node should be removed from the aggregated list.

Expected form of request body:

```
{
    "instance": "56d1a8de-14a8-403b-b3e7-d49307c63553"m
}
```

`PATCH http://localhost:8000/files`

Received from watcher nodes to update the aggregated list of files. JSON body may contain multiple patch operations.

Each patch operation specifies the instance id of the watcher node, the operation type (either `add` or `remove`), the sequence number of the operation, and the file details. The sequence number will be monotonic, incrementing for each operation per watcher node.

Expected form of request body:

```
[
    {
        "instance": "56d1a8de-14a8-403b-b3e7-d49307c63553",
        "op": "add",
        "seqno": 3,
        "value": {
            "filename: "badger.png"
        }
    },
    {
        "instance": "56d1a8de-14a8-403b-b3e7-d49307c63553",
        "op": "remove",
        "seqno": 4,
        "value": {
            "filename: "fish.jpg"
        }
    }
]
```