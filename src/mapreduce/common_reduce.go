package mapreduce

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"
)

type KeyValueList []KeyValue

func (kvs KeyValueList) Len() int {
	return len(kvs)
}

func (kvs KeyValueList) Less(i, j int) bool {
	return strings.Compare(kvs[i].Key, kvs[j].Key) == -1
}

func (kvs KeyValueList) Swap(i, j int) {
	kvs[i], kvs[j] = kvs[j], kvs[i]
}

func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTask int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	//
	// doReduce manages one reduce task: it should read the intermediate
	// files for the task, sort the intermediate key/value pairs by key,
	// call the user-defined reduce function (reduceF) for each key, and
	// write reduceF's output to disk.
	//
	// You'll need to read one intermediate file from each map task;
	// reduceName(jobName, m, reduceTask) yields the file
	// name from map task m.
	//
	// Your doMap() encoded the key/value pairs in the intermediate
	// files, so you will need to decode them. If you used JSON, you can
	// read and decode by creating a decoder and repeatedly calling
	// .Decode(&kv) on it until it returns an error.
	//
	// You may find the first example in the golang sort package
	// documentation useful.
	//
	// reduceF() is the application's reduce function. You should
	// call it once per distinct key, with a slice of all the values
	// for that key. reduceF() returns the reduced value for that key.
	//
	// You should write the reduce output as JSON encoded KeyValue
	// objects to the file named outFile. We require you to use JSON
	// because that is what the merger than combines the output
	// from all the reduce tasks expects. There is nothing special about
	// JSON -- it is just the marshalling format we chose to use. Your
	// output code will look something like this:
	//
	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
	// Your code here (Part I).
	//
	//每个reduceTask读取mapTask生成的中间文件，读取的中间文件是对应的0-i文件上的i，也就是传入进来的reduceTask的值对应的那些中间文件
	//逻辑是编号为reduceTask的doReduce方法读取每个mrtmp.test-X-reduceTask.txt的keyValue内容，然后对其进行排序
	//排序完成之后，汇总到一块，再进行一次排序成整体AllKeyValue
	//针对AllKeyValue的每一项，因为是已经按key排序了的，解析出所有key及其对应的value，也就是有相同key的，并到一起
	//向reduceF传入key及其一组value，生成key的value值，最后写入文件【在这个实验中，key的值全是"",所以组的内容的""】

	tmps := make([]KeyValueList, nMap) //保存该reduceTask对应的nMap个中间文件的解析内容
	var data KeyValueList              //用于汇总全部的keyValue内容
	//循环读取X-reduceTask.txt的内容
	for i := 0; i < nMap; i++ {
		inFile := reduceName(jobName, i, reduceTask)
		f, err := os.Open(inFile)
		if err != nil {
			log.Fatal("Open intermedia file ", err)
		}
		defer f.Close()

		//--解析出内容的kv形式
		decoder := json.NewDecoder(f)
		for {
			var kv KeyValue
			err := decoder.Decode(&kv)
			if err != nil {
				break;
			}
			tmps[i] = append(tmps[i], kv)
		}
		//--对每个文件内容进行排序
		sort.Sort(tmps[i])

		//--汇总
		data = append(data, tmps[i]...)
	}

	//整体排序
	sort.Sort(data)

	//--打开文件
	f, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		log.Fatal("Open Result File: ", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)

	//对已排序的全部内容进行merge，调用用户定义的reduceF，并生成新的keyValue，最后写入文件
	for begin := 0; begin < len(data); {
		//统计出每个key的Value组
		key := data[begin].Key
		var value []string
		end := begin
		for ; end < len(data) && data[end].Key == data[begin].Key; end++ {
			value = append(value, data[end].Value)
		}
		//调用reduceF并写入
		err := encoder.Encode(KeyValue{key, reduceF(key, value)})
		if err != nil {
			log.Fatal("write result: ", err)

		}
		begin = end
	}
}
