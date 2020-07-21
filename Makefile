AGGREGATOR_ADDRESS?=http://127.0.0.1:8000

run:
	cd ./watcher-node && go build
	./watcher-node/watcher-node -p=4001 -dir=./sample-folders/folder1 -aggregator=${AGGREGATOR_ADDRESS} &
	./watcher-node/watcher-node -p=4002 -dir=./sample-folders/folder2 -aggregator=${AGGREGATOR_ADDRESS} &

stop:
	-pkill -f watcher-node

