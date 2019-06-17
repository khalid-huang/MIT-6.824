# MIT-6.824
## Part I: Map/Reduce input and output
相应的项目地址：https://pdos.csail.mit.edu/6.824/labs/lab-1.html

## 实验说明
这个项目地址对实验的说明可以知道，这个实验主要是实现common_map.go和common_reduce.go两个文件中的doMap和doReduce方法；
其中doMap方法主要是针对每个传入的文件【这个文件在测试代码中是由test_test.go生成的，每一行都是一个数字】做如下的操作：
++ 读取文件中的内容，调用test_test.go中的mapFunc函数【在实际的应用中，就相当于用户自定义的Map阶段的函数】对文件的内容进行处理，在这里也就是切分文件的内容，将每一行的数字变成Key值，并设置Value值为""；

## 代码运行
设置好GOPATH之后，在src/mapreduce目录下，运行```go test -run Sequential```
