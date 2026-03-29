package main
import ("fmt"
   
   "time"

   "github.com/hpcloud/tail"
)
func main(){
	//1.【漏斗核心】：造一根传送带（管道）
	//chan string 表示这根管子里只能跑字符串
	logChan := make(chan string)

	//2.修改核心函数：增加一个参数把管子接进来
	runTail := func(fileName string, pipe chan string) {
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
		// 【漏斗核心】：不再直接打印，而是把日志塞进管子
			// 加上文件名作为标识，方便下游识别
			pipe <- fmt.Sprintf("文件: %s | 内容: %s", fileName, line.Text)
		}
	}
	// 3. 启动工人时，把管子交给他们
    go runTail("test1.log", logChan)
	go runTail("test2.log", logChan)

	fmt.Println(" 漏斗模式已开启，统一接收中...")

	// 4. 【漏斗核心】：主工位（唯一出口）
	// 这是一个死循环，谁往管子里塞了东西，这里就立刻能拿到
	for msg := range logChan {
		now := time.Now().Format("15:04:05")
		fmt.Printf("[%s] 【中心工位统一输出】-> %s\n", now, msg)
	}
}