package mr

import "fmt"
import "log"
import "net/rpc"
import "hash/fnv"
import "os"
import "io/ioutil"
import "encoding/json"
import "strconv"
import "sort"
import "math/rand/v2"
import "time"

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

// unique identifier for the worker
var worker_name string

func rpc_call(status int, task_type string, filename string) *MapRedReply {
  args := MapRedArgs{}
  reply := MapRedReply{}
  args.Worker_name = worker_name
  if status == 1 || status ==2 {
    args.Filename = filename
    args.Task_type = task_type
  }
  ok := call("Coordinator.MapRedHandler", &args, &reply)
	if ok {
    return &reply
  }
	fmt.Printf("call failed!\n")
  return nil
}
//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
 
  // create a unique identifier for the worker
  worker_name = "W" + strconv.Itoa(rand.IntN(100)) + strconv.Itoa(rand.IntN(100))
  
	// Your worker implementation here.
  status := 0
  reply := rpc_call(status, "", "x")
  for {
    if status == 1 {
      reply = rpc_call(status, "map", reply.Filename)
    } else if status == 2 {
      reply = rpc_call(status, "reduce", reply.Task_no)
    } else if status == 3 {
      reply = rpc_call(status, "", "x")
    }
    if reply.Task_no == "" {
      fmt.Printf("Job has been finished. Worker: %s exiting\n", worker_name) 
      break
    }
    if reply.Task_type == "map" {
      // read file and apply map function (mapf) 
  		file, err := os.Open(reply.Filename)
  		if err != nil {
  			log.Fatalf("cannot open %v", reply.Filename)
  		}
  		content, err := ioutil.ReadAll(file)
  		if err != nil {
  			log.Fatalf("cannot read %v", reply.Filename)
  		}
  		file.Close()
      // perform map
  		kva := mapf(reply.Filename, string(content))
      
  	  // create reply.N_splits(nReduce) no of files
      files := make([]*os.File, reply.N_splits)
      for i := 0; i < reply.N_splits; i++ {
     	  files[i], err = os.CreateTemp(".", "someprefix")
        if err != nil {
            log.Fatal(err)
        } 
      }
     
      // write kv pairs to files idenfied by the keys of reduce_task_map
      for i := 0; i < len(kva); i++ {
        reduce_task_no := ihash(kva[i].Key) % reply.N_splits
        //file, err := os.OpenFile("MR-" + reply.Task_no + "-" + strconv.Itoa(reduce_task_no), os.O_APPEND | os.O_WRONLY, os.ModeAppend)
        if err != nil {
            log.Fatal(err)
        }
        // defer file.Close()
        enc := json.NewEncoder(files[reduce_task_no])
        enc_err := enc.Encode(&kva[i])
        if enc_err != nil {
			    log.Fatalf("cannot write json to %v", "MR-" + reply.Task_no + "-" + strconv.Itoa(reduce_task_no))
		    }
      }
      
      // atomically rename the temporary files
      for i := 0; i < reply.N_splits; i++ {
     	  err := os.Rename(files[i].Name(), "MR-" + reply.Task_no + "-" + strconv.Itoa(i))
        if err != nil {
            log.Fatal(err)
        }
        files[i].Close() 
      }
      
     status = 1
     // reduce operation
     } else if reply.Task_type == "reduce" {
          var kva []KeyValue
          for i := 0; i < reply.N_splits; i++ {
            file, err := os.Open("MR-" + strconv.Itoa(i) + "-" + reply.Task_no)
            if err != nil {
                log.Fatal(err)
            }
            defer file.Close()
            dec := json.NewDecoder(file)
            for {
              var kv KeyValue
              if err := dec.Decode(&kv); err != nil {
                break
              }
              kva = append(kva, kv)
            }
          }
          intermediate := []KeyValue{}
          intermediate = append(intermediate, kva...)
          sort.Sort(ByKey(intermediate))
          oname := "mr-out-" + reply.Task_no
    	    ofile, _ := os.Create(oname)
          p := 0
        	for p < len(intermediate) {
        		q := p + 1
        		for q < len(intermediate) && intermediate[q].Key == intermediate[p].Key {
        			q++
        		}
        		values := []string{}
        		for k := p; k < q; k++ {
        			values = append(values, intermediate[k].Value)
        		}
        		output := reducef(intermediate[p].Key, values)
        
        		// this is the correct format for each line of Reduce output.
        		fmt.Fprintf(ofile, "%v %v\n", intermediate[p].Key, output)
        
        		p = q
        	}
          ofile.Close()
          status = 2
    // wait for 2s before asking coordinator again
    } else if reply.Task_type == "wait" {
          for i := 0; i < 2; i++ {
        		time.Sleep(1000 * time.Millisecond)
      		}
          reply.Filename = ""
          status = 3
    }
    
  }
	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

  
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
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
