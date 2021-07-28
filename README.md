# av_demo

---

主要是golang播放音视频的demo，golang这方面资料还是比较少，大多也都属于demo阶段


---

GUI这边主要是有[fyne](https://github.com/fyne-io/fyne)这个库可以跑通
然后还有就是[pixel](https://github.com/faiface/pixel/tree/masterpixel)这个库用的比较多

音视频解码就是ffmpeg了，fork了[goav](https://github.com/giorgisio/goav) 加了点方法，这个库也很久不更新了最新版的ffmpeg都跑不通。
用的ffmpeg4.0.3

音频播放研究了俩库[oto](github.com/hajimehoshi/oto)和[portaudio](https://github.com/gordonklaus/portaudio)portaudio没有跑通数据格式转换的有点问题

---
video.go

视频播放代码是跑通了这个[pixel-video](https://github.com/zergon321/pixel-video)
这个用了go-av和pixel，仅解析了视频的画面

audio.go

视频里的音频播放，仅解析音频然后调用oto播放，可以播放就是有杂音，杂音还挺大

---
##TODO
+ 解决音频播放杂音的问题
+ 音视频同步播放
