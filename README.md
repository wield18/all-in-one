### Hello?

#### State

Local Video Multi Player —— 目前仅实现了本地随机多视频播放功能。

#### Before You Get Started / 准备工作

你需要准备一个包含视频文件的文件夹。  
视频可以直接放在该文件夹中，也可以放在它的任意子文件夹下，程序会自动扫描。
You need a folder that contains video files.
The videos can be placed directly in this folder or in any of its subfolders — the program will scan them automatically.

#### How to Start

```bash
git clone https://github.com/wield18/all-in-one.git
cd ./all-in-one
```

依照 `./config/config.tempalte.yaml` 创建 `./config/config.yaml` 并将 `videoRoot` 属性改为你的视频文件夹路径。确保你有足够多的竖视频(宽高比小于3/4)才能获得最好效果,至少一个,而且必须下载ffmpeg

```bash
go mod tidy
go run ./main.go
```

然后双击或打开 `html.html` 文件即可使用。

