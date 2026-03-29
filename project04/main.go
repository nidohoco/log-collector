package main
import ("fmt"
   "sync" // ✅ 必须导入这个包，它是 sync.Pool 的亲爹
   "time"

   "github.com/hpcloud/tail"
)
// 【定义集装箱（结构体）
// 以后每条日志都会包含这三个信息：来源、时间、内容
type LogEntry struct {
	Source  string
	Time    string
	Content string
}
// 2. 【核心优化】：定义对象池
// 作用：像共享单车一样，用完还回来，下次接着骑，不浪费内存
var entryPool = sync.Pool{
	New: func() interface{} {
		return new(LogEntry)
	},
}


func main(){
	// 修改管道类型
	// 现在管子里跑的是 *LogEntry（集装箱的指针），容量设为 100
	logChan := make(chan *LogEntry, 100)
	// 修改核心函数：让它学会“打包”
	runTail := func(fileName string, pipe chan *LogEntry) {
	config:= tail.Config{
		ReOpen:    true,                                 // 日志切分了（如从 app.log 变成 app.log.1），我也能自动切过去
		Follow:    true,                                 // 实时追踪，不出错不停止
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件末尾开始读（Whence 2 表示末尾）
		MustExist: false,                                // 文件暂时不存在也没关系，我等它创建
		Poll:      true,                                 // 轮询模式，兼容性最好
	}
	tails, err := tail.TailFile(fileName, config)
	if err != nil {
		fmt.Printf("启动文件%s采集失败:%v\n",fileName,err)
		return
	}
    //持续读取该文件内容
	for line := range tails.Lines {
		// 【// ✅ 从池子里“借”一个结构体
			entry := entryPool.Get().(*LogEntry)
				entry.Source = fileName
			entry.Time = time.Now().Format("15:04:05")
			entry.Content = line.Text
			// 扔进管道
			pipe <- entry
		}
	}
    go runTail("test1.log", logChan)
	go runTail("test2.log", logChan)

	fmt.Println("🔥 高性能对象池模式已开启！正在监控日志...")

	// 4. 【漏斗核心】：主工位（唯一出口）
	// 【新变化 4】：中心工位统一“拆箱”打印
	for msg := range logChan {
		// 现在我们可以精准地拿到每一个字段
		fmt.Printf("[%s] | 源: %-10s | 内容: %s\n", msg.Time, msg.Source, msg.Content)
		// ✅ 【非常关键】：用完后，把它丢回池子复用
		// 归还前不需要手动清空，因为下次 Get 出来后会被新数据覆盖
		entryPool.Put(msg)
	}
}