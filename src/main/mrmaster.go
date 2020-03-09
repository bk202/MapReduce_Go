package main

//
// start the master process, which is implemented
// in ../mr/master.go
/*

go build -buildmode=plugin ../mrapps/wc.go
rm mr-out-*
rm mrIntermediate*
go run mrmaster.go pg*.txt
go run mrmaster.go pg-grimm.txt
go run mrworker.go wc.so &

*/

//
// Please do not change this file.
//

import "../mr"
import "time"
import "os"
import "fmt"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrmaster inputfiles...\n")
		os.Exit(1)
	}

	m := mr.MakeMaster(os.Args[1:], 10)
	for m.Done() == false {
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)
}
