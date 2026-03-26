package main
import ("fmt"
   "log"

   "github.com/hpcloud/tail"
)
func main(){
	//a.设置配置 我们要怎么盯着文件
	config:= tail.Config{
		ReOpen:    true,                                 // 日志切分了（如从 app.log 变成 app.log.1），我也能自动切过去
		Follow:    true,                                 // 实时追踪，不出错不停止
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件末尾开始读（Whence 2 表示末尾）
		MustExist: false,                                // 文件暂时不存在也没关系，我等它创建
		Poll:      true,                                 // 轮询模式，兼容性最好
	}
	// B. 打开文件（假设我们要盯着 test.log）
	fileName := "test.log"
	tails, err := tail.TailFile(fileName, config)
	if err != nil {
		log.Fatalf("无法启动采集器: %v", err)
	}

	// C. 开始收割：这里是一个死循环
	fmt.Println("🚀 采集器启动成功，正在盯着 test.log...")
	
	// tails.Lines 是一个管道（Channel），有新行它就会吐出来
	for line := range tails.Lines {
		// 删掉所有 %d 和相关的占位符，只打印字符串
		fmt.Printf("【发现新日志】内容: %s\n", line.Text)
	}
}

