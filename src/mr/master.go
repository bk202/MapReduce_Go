package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "fmt"


type Master struct {
	// Worker id also keeps track of the number of workers in system
	// Shut down system only when all active workers have reported shut down
	WorkerID int // An ID assigned to worker upon worker initiation
	NReduce int

	MapTaskID int // 1~nReduce
	ReduceTaskID int // 1~len(InputFiles)

	InputFiles []string // Queue holding all input file names
	IntermediateFiles [][]string // Queue holding all generated intermediate file names
	OutputFiles []string // Queue holding all generated output file names

	State string // state indicating master's current state, {map, reduce, done}
}

// Your code here -- RPC handlers for the worker to call.

// Handler for handling worker creation reports
func (m *Master) WorkerCreation(args *WorkerMessage, reply *WorkerMessage) error{
	fmt.Printf("Received creation message from worker\n")

	reply.ID = m.WorkerID
	reply.NReduce = m.NReduce

	fmt.Printf("Replying with worker ID: %v\n", reply.ID)
	m.WorkerID += 1

	return nil
}

func (m* Master) WorkerShutDown(args *WorkerMessage, reply *WorkerMessage) error{
	m.WorkerID -= 1

	return nil
}

// Handler for assigning task and file name to worker
func (m *Master) AssignTask(args *MasterMessage, reply *MasterMessage) error{
	if m.State == "map"{
		// assign map task to worker
		reply.Task = "map"

		// pop input file to worker
		fileName := m.InputFiles[0]
		m.InputFiles = m.InputFiles[1:len(m.InputFiles)]
		reply.Files = append(reply.Files, fileName)

		// assign task id
		reply.TaskID = m.MapTaskID
		m.MapTaskID += 1

		// Change state on empty input files
		if len(m.InputFiles) == 0{
			m.State = "reduce"
		}
	} else if m.State == "reduce"{
		reply.Task = "reduce"

		// Assign list of intermediate files to worker
		files := m.IntermediateFiles[0]
		m.IntermediateFiles = m.IntermediateFiles[1: len(m.IntermediateFiles)]
		reply.Files = files

		reply.TaskID = m.ReduceTaskID
		m.ReduceTaskID += 1

		// Change state on empty intermediate files
		if len(m.IntermediateFiles) == 0{
			m.State = "done"
		}
	} else{
		reply.Task = "done"
	}

	return nil
}

// Handler for receiving map done jobs
func (m *Master) MapDone(args *Files, reply *Files) error{
	// m.IntermediateFiles = append(m.IntermediateFiles, args.FileNames...)
	for i:=0; i<len(args.FileNames); i++{
		m.IntermediateFiles[i] = append(m.IntermediateFiles[i], args.FileNames[i])
	}
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	fmt.Printf("m.state: %s, m.WorkerID: %v\n", m.State, m.WorkerID)
	return m.State == "done" && m.WorkerID == 0
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	fmt.Printf("Number of files: %v\n", len(files))

	// assign file names to input files
	m.InputFiles = files

	// ID allocated to each worker
	m.WorkerID = 0

	// master should initially in map state
	m.State = "map"

	m.NReduce = nReduce
	m.MapTaskID = 0
	m.ReduceTaskID = 0

	m.IntermediateFiles = make([][]string, m.NReduce)
	for i:=0; i<m.NReduce; i++{
		m.IntermediateFiles[i] = make([]string, 0)
	}

	m.server()
	return &m
}
