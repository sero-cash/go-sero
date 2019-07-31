#please read rungero.docker script,you can add startup params
#when you start gero with docker please add -v save the data to the chain data

build:
	docker build -t sero/client-go:v1.0.0-rc6 .

rpc:
	mkdir -p ~/geroData/log
	docker run --name gero -v ~/geroData:/root/.sero -v ~/geroData/log:/log -d -p 8545:8545 sero/client-go:v1.0.0-rc6

kill:
	docker rm -f gero

test:
	curl -X POST -H "Content-Type":application/json --data '{"jsonrpc":"2.0","method":"sero_blockNumber","params":[""],"id":1}' localhost:8545

