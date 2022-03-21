package main

import(
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/rob121/vhelp"
	"log"
	"os"
	"runtime"
	"strings"
	"github.com/spf13/viper"
	"os/exec"
	"time"
	"context"
)

var conf *viper.Viper

func main(){

	vhelp.Load("config")
	conf,_ = vhelp.Get("config")
    dirwatch:= conf.GetString(fmt.Sprintf("%s.dirwatch",runtime.GOOS))

    log.Println("Watching for PDF in: ",dirwatch)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {

					if(strings.Contains(event.Name,".pdf")) {
						log.Printf("%#v", event)
						go handleFile(event)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dirwatch)
	if err != nil {
		log.Println(err)
	}
	<-done

}

func handleFile(event fsnotify.Event){

	printcmd:= conf.GetStringSlice(fmt.Sprintf("%s.printcmd",runtime.GOOS))
	printer:= conf.GetString(fmt.Sprintf("%s.printer",runtime.GOOS))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	cmd := &exec.Cmd{}

	log.Println("Running:",strings.Join(printcmd," "))

	if len(printer)>0 {
		printcmd = append(printcmd,event.Name)
		printcmd = append(printcmd,printer)

		cmd = exec.CommandContext(ctx,printcmd[0],printcmd[1:]...)

	}else {
		printcmd = append(printcmd,event.Name)
		cmd = exec.CommandContext(ctx,printcmd[0],printcmd[1:]...)

	}


	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Println(err)
	}

	if len(out.String())>0  {
		log.Printf("Output: %q\n", out.String())
	}

	log.Println("Removing file",event.Name)

	rerr := os.Remove(event.Name)

	if rerr != nil {
		log.Println(rerr)
		return
	}
}

