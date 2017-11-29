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
	closeQue chan interface{}=make(chan interface{})
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
    //  signal.Notify(lb.ch, syscall.SIGUSR1,syscall.SIGINT,syscall.SIGTERM)
    signal.Notify(sigQue, syscall.SIGINT,syscall.SIGQUIT)
    go func (){
		v:=<-sigQue
		switch(v){
		case syscall.SIGINT:
			fmt.Println("SIGINT received")
			closeQue<-nil
		case syscall.SIGQUIT:
			fmt.Println("SIGHUP received")
			doneQue<-nil
		}
    }()

	//time.Sleep(3*time.Second)
	wg.Wait()
	fmt.Println("after wg.Wait")

}
