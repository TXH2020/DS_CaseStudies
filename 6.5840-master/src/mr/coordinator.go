package mr

import "log"
import "fmt"
import "net"
import "os"
import "sync"
import "net/rpc"
import "net/http"
import "strconv"
import "time"


type Coordinator struct {
	// Your definitions here.
  filenames []string
  mu sync.Mutex
  // each value in the map_status or reduce_status array corresponds to a filename at the same position in the array. It can be 0 or 1.
  // For map_status
  // 0-no task allocated for file
  // 1-map task underway for file
  // For reduce_status
  // 0-no task allocated for file
  // 1-reduce task underway for file.
  // if all values in map status array equal 1, map phase is over.
  // if all values in reduce status array equal 1, reduce phase is over. Job is completed
  map_status map[string]int
  reduce_status map[int]int
  nReduce int
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) linear_search(worker_status int, task_type string) int {
  if task_type == "map" {
    for i := 0; i < len(c.filenames); i++ {
    if c.map_status[c.filenames[i]] == worker_status {
      return i
    }
  }
  } else {
    for i := 0; i < c.nReduce; i++ {
    if c.reduce_status[i] == worker_status {
      return i
    }
  }
     }
  
  return -1
}

func (c *Coordinator) await_status(task_type string, task_index int) {
	// wait for 10s. During wait, check if we received a response from worker. If yes, stop. Otherwise, assign task to another worker.
  for i := 0; i < 10; i++ {
    if task_type == "map" {
      if c.map_status[c.filenames[task_index]] == 2 {
        return
      }
    } else {
      if c.reduce_status[task_index] == 2 {
        return
      }
    }
		time.Sleep(1000 * time.Millisecond)
	}
  if task_type == "map" {
    c.mu.Lock()
    c.map_status[c.filenames[task_index]] = 0
    c.mu.Unlock()
  } else {
      c.mu.Lock()
      c.reduce_status[task_index] = 0
      c.mu.Unlock()
  }
  return
}

func (c *Coordinator) MapRedHandler(args *MapRedArgs, reply *MapRedReply) error {
  if args.Filename != "" {
    if args.Task_type == "map" {
      c.map_status[args.Filename] = 2
      fmt.Printf("Received response for map file: %s from worker: %s\n", args.Filename, args.Worker_name)
    } else {
      // convert filename to integer(reduce task number)
      filename, _ := strconv.Atoi(args.Filename)
      c.reduce_status[filename] = 2
      fmt.Printf("Received response for reduce file: %d from worker: %s\n", filename, args.Worker_name)
    }
  }
  // provide one file to each worker and assign the filename with map pending status(1). 0 means unassigned.
  task_index := c.linear_search(0, "map")
  if task_index != -1 {
      reply.Task_no = strconv.Itoa(task_index)
      reply.Filename = c.filenames[task_index]
      fmt.Printf("Received request. Assigning map file: %s to worker: %s\n", reply.Filename, args.Worker_name)
      reply.Task_type = "map"
      reply.N_splits = c.nReduce
      // since status array is updated, it needs to be safe for concurrency. Hence, use Mutex.
      c.mu.Lock()
      c.map_status[c.filenames[task_index]] = 1
      c.mu.Unlock()
      go c.await_status("map", task_index)
      return nil
    }
  // if all of the files have been allocated and map is pending for all, tell worker to wait.
  task_index = c.linear_search(1, "map")
  if task_index != -1 {
    fmt.Printf("Maps in progress. Worker: %s\n", args.Worker_name)
    reply.Task_no = "x"
    reply.Task_type = "wait"
    return nil
  }
  // all map tasks have been completed. Assign one reduce task number to the worker.
  task_index = c.linear_search(0, "reduce")
  if task_index != -1 {
      reply.Task_no = strconv.Itoa(task_index)
      fmt.Printf("Received request. Assigning red file: %s to worker: %s\n", reply.Task_no, args.Worker_name)
      reply.Task_type = "reduce"
      reply.N_splits = len(c.filenames)
      // since status array is updated, it needs to be safe for concurrency. Hence, use Mutex.
      c.mu.Lock()
      c.reduce_status[task_index] = 1
      c.mu.Unlock()
      go c.await_status("reduce", task_index)
      return nil
    }
  // if all of the files have been allocated and reduce is pending for all, tell worker to wait.
  task_index = c.linear_search(1, "reduce")
  if task_index != -1 {
    fmt.Printf("Reds in progress. Worker: %s\n", args.Worker_name)
    reply.Task_no = "x"
    reply.Task_type = "wait"
    return nil
  }
  return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.
  count := 0
  for i := 0; i < c.nReduce; i++ {
    if c.reduce_status[i] == 2 {
      count += 1
    }
  }
  if count == c.nReduce {
    fmt.Println("Job Finished. Coordinator terminating.")
    ret = true
  }
	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
  // initialize the necessary variables.
  c.filenames = files
  c.map_status = make(map[string]int)
  c.reduce_status = make(map[int]int)
  c.nReduce = nReduce
  for i := 0; i < len(files); i++ {
		c.map_status[files[i]] = 0
	}
  for i := 0; i < nReduce; i++ {
		c.reduce_status[i] = 0
	}  
  
	c.server()
	return &c
}
