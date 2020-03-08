package mr

import "fmt"
import "log"
import "net/rpc"
import "hash/fnv"
import "io/ioutil"
import "os"
import "strconv"
import "bufio"
import "strings"

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func min(a int, b int) int{
	if a < b{
		return a
	}
	return b
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// uncomment to send the Example RPC to the master.
	// CallExample()

	// 1. notify master of worker creation
	workerID, nReduce := WorkerCreation()

	for true{
		task, files, taskID := RequestWork()

		if task == "done"{
			fmt.Printf("Worker %v received done signal", workerID)

			// Notify master of shut down completion
			WorkerShutDown(workerID)
			return
		}

		if task == "map"{
			fmt.Printf("Worker %v received map task\n", workerID)

			fileName := files[0]
			// read file contents
			file, _:= os.Open(fileName)
			contents, _ := ioutil.ReadAll(file)
			file.Close()

			kva := mapf(fileName, string(contents))

			// Generate 10 intermediate files
			offset := len(kva) / nReduce
			start := 0
			end := start + offset

			intermediateFiles := make([]string, 0)

			for i:=0; i<nReduce; i++{
				end = min(end, len(kva))

				segment := kva[start:end]
				start += offset
				end += offset

				// Write to intermediate file
				fileName := "mrIntermediate-" + strconv.Itoa(taskID) + "-" + strconv.Itoa(i)
				intermediateFiles = append(intermediateFiles, fileName)

				ofile, _ := os.Create(fileName)
				for j:=0; j<len(segment); j++{
					pair := segment[j]

					fmt.Fprintf(ofile, "%v %v\n", pair.Key, pair.Value)
				}
			}

			MapDone(intermediateFiles)

		} else if task == "reduce"{
			// Create <word, list(pair(word, 1))> hash map
			kv_map := make(map[string]([]string))

			fmt.Printf("Worker %v reduce task received\n", workerID)

			// Hash all rows in each intermediate file
			for i:=0; i<len(files); i++{
				file := files[i]

				// read file contents
				f, _ := os.Open(file)

				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					line := scanner.Text()

					words := strings.Fields(line)
					key := words[0]

					kv_map[key] = append(kv_map[key], line)
				}

				f.Close()
			}

			// Sort keys in ascending order
			sortedKeys := make([]string, 0)

			for k, _ := range kv_map{
				sortedKeys = append(sortedKeys, k)
			}

			// Create output file
			fileName := "mr-out-" + strconv.Itoa(taskID)
			ofile, _ := os.Create(fileName)

			// Perform reduce on each sorted key
			for i:=0; i<len(sortedKeys); i++{
				count := reducef(sortedKeys[i], kv_map[sortedKeys[i]])

				fmt.Fprintf(ofile, "%v %v\n", sortedKeys[i], count)
			}
		}
	}

}

// Helper function for notifying master of worker creation and
// receive an ID from master
func WorkerCreation() (int, int){ // rval: Worker id, nReduce
	workerMsg := WorkerMessage{}

	call("Master.WorkerCreation", &workerMsg, &workerMsg)

	fmt.Printf("Received ID (%v) from master\n", workerMsg.ID)

	return workerMsg.ID, workerMsg.NReduce
}

// Helper function for requesting work and inputs from master
func RequestWork() (string, []string, int){ // rtype: task type, file name, task id
	masterMsg := MasterMessage{}
	masterMsg.Files = make([]string, 0)

	call("Master.AssignTask", &masterMsg, &masterMsg)

	return masterMsg.Task, masterMsg.Files, masterMsg.TaskID
}

// Helper function for notifying master map job finished and
// sends master intermdiate file names
func MapDone(fileNames []string){
	msg := Files{}
	msg.FileNames = make([]string, len(fileNames))

	for i:=0; i<len(fileNames); i++{
		msg.FileNames[i] = fileNames[i]
	}

	call("Master.MapDone", &msg, &msg)
}

// Helper function for notifying master of worker shut down
func WorkerShutDown(workerID int){
	workerMsg := WorkerMessage{}
	workerMsg.ID = workerID

	call("Master.WorkerShutDown", &workerMsg, &workerMsg)
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
