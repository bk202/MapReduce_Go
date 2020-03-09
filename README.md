# MapReduce_Go

This is an implementation of MIT distributed systems course 6.824 lab1.
See instructions at https://pdos.csail.mit.edu/6.824/labs/lab-mr.html

## Instructions to run
Execute `bash src/main/test.sh` to spawn a master process and 3 worker processes to carry out the distributed work count task.

This lab implements a distributed word count function using the following structure:

![alt text](https://upload-images.jianshu.io/upload_images/1863961-e3049b9f4645a329.PNG?imageMogr2/auto-orient/strip%7CimageView2/2/w/840)

## Master-Worker scheme
1. The master process will be responsible for assigning tasks and input files to workers.
2. All communications between master and workers will be done via remote-procedure-calls (RPC).
3. Master process is a finite state machine, i.e. Master assigns tasks to workers based on its state. (Map, Reduce, Done)
4. Master will keep track of input files and intermediate files in a queue and assign to workers along with task info.

## Work assignment
1. All workers will report to master upon it's initialization and receive a worker ID from master, in practice, master will use this worker ID to keep track of tasks assigned to workers, if master discovers workers death (via heartbeat or wait time), master will put task back to task queue.
2. All tasks (map/reduce) will be assigned with a task ID for output file naming conventions.
3. All workers will report to master upon task completion.

## Intermediate and output files
1. Master will be assigned with a `nReduce` variable, each input file will be executed with map function, hence master will have a `nMap` variable.
2. Each input file will be splitted into `nReduce` intermediate files, hence we have `nMap * nReduce` intermediate files:

` buckets[0] = ['mrIntermediate-0-0', 'mrIntermediate-1-0', 'mrIntermediate-2-0' ...]`

`buckets[1] = ['mrIntermediate-0-1', 'mrIntermediate-1-1', 'mrIntermediate-2-1' ...]`
 
` ...`

 `buckets[9] = ['mrIntermediate-0-9', 'mrIntermediate-1-9', 'mrIntermediate-2-9' ...]`
 
 3. Each bucket will ultimately produce 1 output file `mr-out-X`
 
 ## Workers
 1. Workers will notify master upon map/reduce completion and sends intermdiate files names and output file names to master.
 2. Workers will use hash map in reduce tasks to count word frequencies.
 
 ## Master
 1. Master will be updating its variables concurrently, hence it is essential to lock its critical code.
 2. There may exist a gap between state transitions and input files for next state is not read
 i.e. master is in reduce state but have yet received all intermediate files, some may still be processing, if master have insufficient intermediate files to assign to workers, master will pass a -1 taskID to worker and worker will sleep through this task
 
 ## State Change
 1. It is critical how we determine when the master changes state, hence I defined two conditions for master to update its state:
 
 `Map` -> `Reduce`: Master changes its state to `Reduce` upon receiving `nMap * nReduce` intermediate files from workers.
 `Reduce` -> `Done`: Master changes its state to `Done` upon receiving `nReduce` output files from workers.
 
 2. Master terminates itself on receiving `nReduce` output files, however, in a more careful design, master should terminate itself when it sends `Done` message to all workers and an acknowledgement message from all workers.
