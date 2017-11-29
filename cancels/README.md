Test for golang chan cancel and drain
问题：如何优雅的shutdown一个数据处理机
何为数据处理机，想象一个物联网的数据处理中心，多个终端为其提供实时数据，数据处理机负责将数据处理后存储在数据库中，但数据可能是像潮汐一样，有时涌来一大批，有时却非常少，我们需要为其提供缓存机制，以及时响应终端的需求。这种组件一般称为数据处理机（DPM, Data Processing Mechine）
对于数据处理机的shutdown，我们一定要确保DP中的数据安全，就是说，在关闭DP之前一定要将缓存的数据处理完成，释放所有资源
长话短说，我们用一段代码来说明，完整代码可以在此下载

```
package main

import (
	"container/list"
	"fmt"
	"strconv"
	"sync"
	"time"
	"os"
	"os/signal"
	"syscall"
)
/*
This routine is for forever data mechine cancel and drain
*/

//	pump generate data then push it into chan in
func pump(in chan string, wg *sync.WaitGroup) {
	fmt.Println("pump start")
	defer wg.Done()
	i := 0
	pumploop:
	for {
		select{
		case _,ok:=<-closeQue:
			if ok {
				close(closeQue)
			}
			fmt.Println("break pumploop")
			break pumploop
		case _,ok:=<-doneQue:
			if ok {
				in<-""
			}
			fmt.Println("quit message received")
			fmt.Println("break pumploop")
			break pumploop


		case in <- "msg" + strconv.Itoa(i):

			fmt.Printf("pump msg%d \n ",i)
			time.Sleep(2 * time.Millisecond)
		}
		i++
		
	}
	fmt.Println("pump wg.Done")
}

//process get data from out then process
func process(out chan string, wg *sync.WaitGroup) {
	processLoop:
	for {
		select{
		case _, ok :=<-closeQue:
			if ok {
				close(closeQue)
			}
			fmt.Println("closeQue closed")
			break processLoop
		case str := <-out:
			if len(str)==0{
				break processLoop
			}
			fmt.Printf("proces,%s\n",str)
		}
			time.Sleep(20 * time.Millisecond)
	}
	fmt.Println("process wg.Done")
	wg.Done()
}

//cache for cache data
func cache(in, out chan string,wg *sync.WaitGroup) {
	c := list.New()

	var str string
	cacheLoop:
	for {

		fmt.Printf("c.len=%d\n",c.Len())
		if c.Len() == 0 {
			select{
			case _, ok:=<-closeQue:
				if ok {
					close(closeQue)
				}

				break cacheLoop
			case v := <-in:
					c.PushBack(v)
					str = v
			}
		}

		select {
		case _, ok:=<-closeQue:
			if ok {
				close(closeQue)
			}
			break cacheLoop

		case strx := <-in:
			c.PushBack(strx)
			//str = strx

		case out <- str:
			v := c.Remove(c.Front())
			if strx, ok := v.(string); ok == true {
				str = strx
				if len(str)==0{
					close(closeQue)
					break cacheLoop
				}
			}
		}
	}
	fmt.Println("cache wg.Done")
	wg.Done()
}

var (
	//closeQue is the chan for close the data mechine immediately
	closeQue chan interface{}=make(chan interface{})
	//doneQue is the chan for close the data mechine before data in list drain 
	doneQue chan interface{}=make(chan interface{})
)

func main() {
	fmt.Println("main start")
	in := make(chan string)
	out := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go pump(in, &wg)

	for i:=0;i<4;i++{
		wg.Add(1)
		go process(out, &wg)
	}

	wg.Add(1)
	go cache(in,out, &wg)

	sigQue := make(chan os.Signal)

	//Notify me if user press Ctrl-C(SIGINT), Ctrl-\(SIGQUIT)
    signal.Notify(sigQue, syscall.SIGINT,syscall.SIGQUIT)
    go func (){
		v:=<-sigQue
		switch(v){
		case syscall.SIGINT:
			fmt.Println("SIGINT received")
			closeQue<-nil
		case syscall.SIGQUIT:
			fmt.Println("SIGQUIT received")
			doneQue<-nil
		}
    }()

	//time.Sleep(3*time.Second)
	wg.Wait()
	fmt.Println("after wg.Wait")

}
```
程序说明：
pump 在一个for循环中产生数据，并送入chan in
 
```
in <- "msg" + strconv.Itoa(i)
```
process 在一个for循环中不断从chan out接受数据，然后处理

```
```
cache 在一个for循环中不断从chan in取出数据，然后将数据推入list后端
同时，从不断从list前端取出数据，送入chan out
```
 func cache(in, out chan string,wg *sync.WaitGroup) {
```

