:: protoc --go_out=. --micro_out=. demo.proto
:: protoc --go_out=. demo.proto

protoc -odemo.txt --go_out=. demo.proto
NB_Cli gen --pb-descriptor=./demo.txt --output-path=./
del demo.txt