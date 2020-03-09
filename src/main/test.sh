
go build -buildmode=plugin ../mrapps/wc.go
rm mr-out-*
rm mrIntermediate*
go run mrmaster.go pg*.txt &
# go run mrmaster.go pg-grimm.txt
go run mrworker.go wc.so &
go run mrworker.go wc.so &
go run mrworker.go wc.so &