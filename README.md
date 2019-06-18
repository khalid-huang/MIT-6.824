# MIT-6.824
Part II: Single-worker word count

### 代码逻辑
这个实验的目的是去实验main/wc.go下面折mapFunc和reduceFunc两个函数，这两个是就是用户自定义的用于处理数据的函数，mapFunc函数的作用是针对每个传入的文件内容，获取内容中的所有单词，作为Key值 ，其Value值设置为1；reduceFunc函数的作用是根据传入的Key值和Value列表值，计算该Key值所有的Value值的和，其实就是统计这个Key值对应的单词有几个

### 代码运行
```bash ./test-wc.sh```
