run:
	cd ./watcher-node && go build
	FILE_AGGREGATOR_ADDRESS=http://127.0.0.1:8000 ./watcher-node/watcher-node -p=4001 -dir=./sample-folders/folder1 &
	FILE_AGGREGATOR_ADDRESS=http://127.0.0.1:8000 ./watcher-node/watcher-node -p=4002 -dir=./sample-folders/folder2 &

stop:
	-pkill -f watcher-node

