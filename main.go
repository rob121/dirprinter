package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/rob121/dirprinter/icon"
	"github.com/rob121/vhelp"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"runtime"
	"path"
	"strings"
	"time"
	"flag"
)

var conf *viper.Viper
var status chan string
var dstat chan string
var debug bool

func main(){

	flag.BoolVar(&debug,"debug",false,"Debug Mode?")
	flag.Parse()

	status = make(chan string)
	dstat = make(chan string)

	vhelp.Load("config")
	conf,_ = vhelp.Get("config")

	go loadDirWatch()

	go systray.Run(onReady, onExit)

	select{}

}

func onReady() {

	dirwatch:= conf.GetString(fmt.Sprintf("%s.dirwatch",runtime.GOOS))

	systray.SetIcon(icon.Data)
//	systray.SetTitle("Print Watcher")
	systray.AddMenuItem(fmt.Sprintf("Watching: %s",dirwatch), "")

	mStatus := systray.AddMenuItem("Idle...", "")

	var mDebug = &systray.MenuItem{}

	if(debug){

	mDebug = systray.AddMenuItem("Debug Window...", "")

	}

	mQuit := systray.AddMenuItem("Quit", "Exit")

	// Sets the icon of a menu item. Only available on Mac and Windows.
	//mQuit.SetIcon(icon.Data)

	for {
		select {
		case <-mQuit.ClickedCh:
			systray.Quit()
			os.Exit(1)
			return
		case s:= <-status:
			mStatus.SetTitle(s)
		case ds:= <-dstat:
			if(debug) {
				mDebug.SetTitle(ds)
			}
		}

	}

}

func onExit() {
	// clean up here

	os.Exit(1)
}

func loadDirWatch(){


    dirwatch:= conf.GetString(fmt.Sprintf("%s.dirwatch",runtime.GOOS))

    log.Println("Watching for PDF in: ",dirwatch)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return
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

	status <- fmt.Sprintf("Printing: %s",path.Base(event.Name))

	log.Println("Running:",strings.Join(printcmd," "))

	if len(printer)>0 {
		printcmd = append(printcmd,event.Name)
		printcmd = append(printcmd,printer)

		cmd = exec.CommandContext(ctx,printcmd[0],printcmd[1:]...)

	}else {
		printcmd = append(printcmd,event.Name)
		cmd = exec.CommandContext(ctx,printcmd[0],printcmd[1:]...)

	}

	dstat <- fmt.Sprint("Print CMD:",strings.Join(printcmd," "))

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

	time.Sleep(5 * time.Second)

	status <- "Idle..."
	dstat <- "Debug Window..."
}

