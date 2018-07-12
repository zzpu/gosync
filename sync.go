package main
 
import (
    "fmt"
    "time"
 
    "os/exec"
 
    "strings"
 
    "io/ioutil"

    "github.com/andlabs/ui"

    "os"
 
    "github.com/howeyc/fsnotify"
 
    "sync"
)
 
var (
    filePath   = "/opt/code/molview/"
    hostname   = "root@103.215.190.172"
    remotePath = "/opt/code/molview/"
)
 
var watcher *fsnotify.Watcher
var mutex sync.Mutex
 
func Print(args ...interface{}) {
    fmt.Println(time.Now(), args)
}
func isDir(path string) bool {
    fileInfo, err := os.Stat(path)
    if err != nil {
        Print("error:", err.Error())
        return false
    }
    return fileInfo.IsDir()
}
func watchPath(filePath string) {
    Print("watchPath:", filePath)
    err := watcher.Watch(filePath)
    if err != nil {
        Print("Watch error: ",err.Error())
        return
    }
}
func broweDir(path string) {
    Print("broweDir:", path)
    dir, err := os.Open(path)
    if err != nil {
        Print("error:", err.Error())
        return
    }
    defer dir.Close()
    names, err := dir.Readdirnames(-1)
    if err != nil {
        Print("BrowerDir error:", err.Error())
        return
    }
    for _, name := range names {
        dirPath := path + "/" + name
        if !isDir(dirPath) {
            continue
        }
        watchPath(dirPath)
        broweDir(dirPath)
    }
}
 
func main() {

    err := ui.Main(func() {

        button := ui.NewButton("监听")
        info := ui.NewLabel("")
        vbox := ui.NewVerticalBox()
        sPath :=ui.NewHorizontalBox()
        vbox.Append(sPath,false)

        dPath :=ui.NewHorizontalBox()
        vbox.Append(dPath,false)



        sPath.Append(ui.NewLabel("输入本地路径:"), false)
        sBox := ui.NewEntry()
        sPath.Append(sBox, true)


        dPath.Append(ui.NewLabel("输入远程路径:"), false)
        dBox := ui.NewEntry()
        dPath.Append(dBox, true)



        vbox.Append(button,false)

        vbox.Append(info,false)
        //创建window窗口。并设置长宽。
        window := ui.NewWindow("Sync", 600, 200, false)
        //mac不支持居中。
        //https://github.com/andlabs/ui/issues/162
        window.SetChild(vbox)
        button.OnClicked(func(*ui.Button) {
            //可以直接打印日志。
            var err error
            watcher, err = fsnotify.NewWatcher()
            if err != nil {
                panic(err)
            }

            info.SetText("本地路径： " + sBox.Text() )
            info.SetText(info.Text()+ "\n" + "远程路劲： " + dBox.Text() )

            fmt.Println("本地路径： " + sBox.Text() )
            fmt.Println("远程路径： " + dBox.Text() )

            filePath = sBox.Text()
            remotePath = dBox.Text()

            broweDir(filePath)
            watchPath(filePath)
            go dealWatch()


        })
        window.OnClosing(func(*ui.Window) bool {
            ui.Quit()
            return true
        })
        window.Show()
    })
    if err != nil {
        panic(err)
    }



}
func copy(event *fsnotify.FileEvent) *exec.Cmd {
    return exec.Command(
        "scp",
        "-r",
        event.Name,
        hostname+":"+remotePath+strings.TrimPrefix(event.Name, filePath))
}
func remove(event *fsnotify.FileEvent) *exec.Cmd {
    return exec.Command(
        "ssh",
        hostname,
        `rm -r `+remotePath+strings.TrimPrefix(event.Name, filePath)+``)
}
func dealWatch() {
    for {
        func() {
            //mutex.Lock()
            //defer mutex.Unlock()
            select {
            case event := <-watcher.Event:
                    Print("event: ", event)
                    var cmd *exec.Cmd
                    if event.IsCreate() || event.IsModify() {
                        cmd = copy(event)
                    }
                    if event.IsDelete() || event.IsRename() {
                        cmd = remove(event)
                    }
                    Print("cmd:", cmd.Args)
                    stderr, err := cmd.StderrPipe()
                    if err != nil {
                        Print(err.Error())
                        return
                    }
                    defer stderr.Close()
                    stdout, err := cmd.StdoutPipe()
                    if err != nil {
                        Print(err.Error())
                        return
                    }
                    defer stdout.Close()
                    if err = cmd.Start(); err != nil {
                        Print(err.Error())
                        return
                    }
                    errBytes, err := ioutil.ReadAll(stderr)
                    if err != nil {
                        Print(err.Error())
                        return
                    }
                    outBytes, err := ioutil.ReadAll(stdout)
                    if err != nil {
                        Print(err.Error())
                        return
                    }
                    if len(errBytes) != 0 {
                        Print("errors:", string(errBytes))
                    }
                    if len(outBytes) != 0{
                        Print("output:", string(outBytes))
                    }
                    if err = cmd.Wait(); err != nil {
                        Print(err.Error())
                    }
            case err := <-watcher.Error:
                Print("error: ", err.Error())
            }
        }()
    }
}
