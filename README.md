# MIT-6.824
## Part I: Map/Reduce input and output
相应的项目地址：https://pdos.csail.mit.edu/6.824/labs/lab-1.html

## 实验说明
这个项目地址对实验的说明可以知道，这个实验主要是实现common_map.go和common_reduce.go两个文件中的doMap和doReduce方法；
其中doMap方法主要是针对每个传入的文件nMap个文件【这个文件在测试代码中是由test_test.go生成的，每一行都是一个数字】做如下的操作：
- 读取文件中的内容，调用test_test.go中的mapFunc函数【在实际的应用中，就相当于用户自定义的Map阶段的函数】对文件的内容进行处理，在这里也就是切分文件的内容，将每一行的数字变成Key值，并设置Value值为""，将所有的KV组成一个数组并返回；
- 将返回的KV数组的所有条目划分到nReduce个文件中，如何划分呢？就是调用common_map.go文件中的ihash，对每个Key值生成hashcode并对nReduce取模，然后放到相应的数组中

经过doMap的处理之后呢，原本的1个输入文件就会产生nReduce个中间文件，这些中间呢，命名类似于mrtmp.test-i-j，其中test表示任务的名字，i表示对应的第i个输入文件，j表示该文件的第j个中间文件，其最大值是nReduce-1

doReduce方法会被调用nReduce次，每次都称为一次ReduceTask，每个ReduceTask都有一个编号，编号最大是nReduce-1，与上面的doMap生成的文件对应。doReduce做的操作如下【设定该次ReductTask的编号是reduceTask】
- 循环nMap次，每次读取mrtmp.test-i-reduceTask文件的内容。可以看出当reduceTask=1的时候，就是处理所有的输入文件产生的第2个中间文件，也就是所有的mrtmp.test-i-j中j=1的文件
- 获取文件中的KV数组，然后让其根据Key值进行排序
- 将上述的nMap个KV数组汇总成一个大的KV数组，然后再进行排序
- 针对汇总的KV数组【已排序的】，遍历出相同的Key值，将其Value值进行汇总成一个Value数组，将这个Value数组和Key值 传入到test_test.go中定义reduceFunc函数中，生成一个新的KV数组【在这个实验中，Value其实都是空】，并写入到文件中

## 代码运行
设置好GOPATH之后，在src/mapreduce目录下，运行```go test -run Sequential```
