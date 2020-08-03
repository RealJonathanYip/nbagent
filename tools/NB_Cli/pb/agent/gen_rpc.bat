:: protoc --go_out=. --micro_out=. agent.proto
:: protoc --go_out=. agent.proto

protoc -oagent.txt --go_out=. agent.proto
NB_Cli gen --pb-descriptor=./agent.txt --output-path=./
del agent.txt