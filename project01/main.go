package main
import ("fmt"
   "log"
   "time"

   "github.com/hpcloud/tail"
)
func main(){
	//1.抽离出一个核心函数 专门负责监听一个文件
	runTail := func(fileName string) {
	config:= tail.Config{
		ReOpen:    true,                                 // 日志切分了（如从 app.log 变成 app.log.1），我也能自动切过去
		Follow:    true,                                 // 实时追踪，不出错不停止
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件末尾开始读（Whence 2 表示末尾）
		MustExist: false,                                // 文件暂时不存在也没关系，我等它创建
		Poll:      true,                                 // 轮询模式，兼容性最好
	}
	tails, err := tail.TailFile(fileName, config)
	if err != nil {
		log.Printf("启动文件%s采集失败:%v\n",fileName,err)
		return
	}
    //持续读取该文件内容
	for line := range tails.Lines {
		// 获取现在的时间
			now := time.Now().Format("15:04:05")
			// 把结果印在屏幕上
			fmt.Printf("[%s] 发现新日志 -> 文件: %s 内容: %s\n", now, fileName, line.Text)
		}
	}

	// 【分身启动】同时盯着两个文件
	go runTail("test1.log")
	go runTail("test2.log")

	// 【门闩】别让程序关掉
	fmt.Println("🚀 采集器已启动，请往 test1.log 或 test2.log 写入内容...")
	select {}
}

