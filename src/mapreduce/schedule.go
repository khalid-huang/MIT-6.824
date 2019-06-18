package mapreduce

import (
	"fmt"
	"sync"
)

//
// schedule() starts and waits for all tasks in the given phase (mapPhase
// or reducePhase). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int
	var n_other int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		n_other = nReduce
	case reducePhase:
		ntasks = nReduce
		n_other = len(mapFiles)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, n_other)

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part III, Part IV).
	//
	//循环调用call ntasks次,并等待所有的协程都结束
	var wg sync.WaitGroup
	for i := 0; i < ntasks; i++ {
		wg.Add(1)
		//封装当前任务需要的参数
		var args DoTaskArgs
		args.JobName = jobName
		args.Phase = phase
		args.TaskNumber = i
		args.NumOtherPhase = n_other
		if phase == mapPhase {
			args.File = mapFiles[i]
		}
		//每个task调用一次call
		go func() {
			defer wg.Done()
			for { //放入到无限循环中，只要task没有完成，就一直执行
				worker := <-registerChan //获取一个就绪态的worker
				ok := call(worker, "Worker.DoTask", &args, nil)
				if ok { //如果没有成功的话重新开始循环; 如果成功就结束
					go func() {
						registerChan <- worker //将worker重新放回，让其可以接受新的task
					}()
					break
				}
			}
		}()
	}

	fmt.Printf("Schedule: %v done\n", phase)
}
