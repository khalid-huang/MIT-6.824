# MIT-6.824
+ Part III: Distributing MapReduce tasks
+ Part IV: Handling worker failures

### 代码逻辑
这个实验的目的是实现一个调度方法，就是scheduler.go里面的scheduler方法，这个方法的逻辑就是通过调用nTask次Call方法来分配任务，通过waitGroup来确保所有协程的结束，每次Worker完成任务后都要将Worker的状态设置为就绪态，也就是重新放回registerChan中；
针对partIV，就是将task的处理放在一个无限循环中，只要任务没有成功就重新分配新的Worker

### 代码运行
```go test -run TestParallel```
