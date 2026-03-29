package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hpcloud/tail"
)

// 1. 【账本结构】
type LogEntry struct {
	Source, Time, Content string
}

var (
	entryPool = sync.Pool{New: func() interface{} { return new(LogEntry) }}
	registry  = make(map[string]int64) // 内存账本：文件名 -> 偏移量
	regMu     sync.Mutex               // 账本锁
)

// 2. 【存盘逻辑】：将内存里的 registry 存入 registry.json
func saveRegistry() {
	regMu.Lock()
	defer regMu.Unlock()
	data, _ := json.MarshalIndent(registry, "", "  ")
	_ = os.WriteFile("registry.json", data, 0644)
}

// 3. 【读取逻辑】：启动时加载旧账本
func loadRegistry() {
	data, err := os.ReadFile("registry.json")
	if err == nil {
		_ = json.Unmarshal(data, &registry)
	}
}

func main() {
	loadRegistry() // 启动先读账本
	logChan := make(chan *LogEntry, 100)

	runTail := func(fileName string) {
		regMu.Lock()
		offset := registry[fileName] // 获取上次的位置
		regMu.Unlock()

		config := tail.Config{
			ReOpen: true, Follow: true, Poll: true,
			Location: &tail.SeekInfo{Offset: offset, Whence: 0},
		}

		tails, err := tail.TailFile(fileName, config)
		if err != nil {
			fmt.Println("开启失败:", fileName)
			return
		}

		for line := range tails.Lines {
			entry := entryPool.Get().(*LogEntry)
			entry.Source, entry.Time, entry.Content = fileName, time.Now().Format("15:04:05"), line.Text
			logChan <- entry

			// ✅ 修正报错：使用 tails.Tell() 获取当前偏移量
			pos, _ := tails.Tell()
			regMu.Lock()
			registry[fileName] = pos
			regMu.Unlock()
		}
	}

	// 启动采集工人
	go runTail("test1.log")
	go runTail("test2.log")

	// 4. 【记账秘书】：每 5 秒存一次盘
	go func() {
		for {
			time.Sleep(5 * time.Second)
			saveRegistry()
			fmt.Println("💾 [系统] 进度已自动存档至 registry.json")
		}
	}()

	fmt.Println("🏆 终极版已就绪！支持断点续传 + 自动存盘")
	for msg := range logChan {
		fmt.Printf("[%s] | 源: %-10s | 内容: %s\n", msg.Time, msg.Source, msg.Content)
		entryPool.Put(msg)
	}
}