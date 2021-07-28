package main

import (
	"bytes"
	"fmt"
	"github.com/fdsnsf/goav/avcodec"
	"github.com/fdsnsf/goav/avformat"
	"github.com/fdsnsf/goav/avutil"
	"github.com/fdsnsf/goav/swresample"
	"github.com/hajimehoshi/oto"
	"io"
	"os"
	"unsafe"
)

func init() {
	avformat.AvRegisterAll()
}

func audio() (<-chan []byte, error) {
	buffer := make(chan []byte, 19200*2)
	go func() {
		in := "./sample.mp4"
		inCtx := avformat.AvformatAllocContext()
		if avformat.AvformatOpenInput(&inCtx, in, nil, nil) < 0 {
			fmt.Println("open file error")
			os.Exit(1)
		}
		defer inCtx.AvformatCloseInput()
		if inCtx.AvformatFindStreamInfo(nil) < 0 {
			fmt.Println("couldn't find stream information")
			os.Exit(1)
		}
		var audioStreamIndex = -1
		for i := 0; i < len(inCtx.Streams()); i++ {
			if inCtx.Streams()[i].CodecParameters().AvCodecGetType() == avformat.AVMEDIA_TYPE_AUDIO {
				fmt.Println("find audio index", i)
				audioStreamIndex = i
			}
		}
		pCodecCtxOrig := inCtx.Streams()[audioStreamIndex].Codec()
		// Find the decoder for the video stream
		pCodec := avcodec.AvcodecFindDecoder(avcodec.
			CodecId(pCodecCtxOrig.GetCodecId()))

		if pCodec == nil {
			fmt.Println("Unsupported codec!")
			os.Exit(1)
		}

		// Copy context
		pCodecCtx := pCodec.AvcodecAllocContext3()
		defer pCodecCtx.AvcodecClose()

		if pCodecCtx.AvcodecCopyContext((*avcodec.
			Context)(unsafe.Pointer(pCodecCtxOrig))) != 0 {
			fmt.Println("Couldn't copy codec context")
			os.Exit(1)
		}

		// Open codec
		if pCodecCtx.AvcodecOpen2(pCodec, nil) < 0 {
			fmt.Println("Could not open codec")
			os.Exit(1)
		}

		pkt := avcodec.AvPacketAlloc()
		defer pkt.AvPacketUnref()
		if pkt == nil {
			fmt.Println("packet all" +
				"oc erbror")
			os.Exit(1)
		}

		utilFrame := avutil.AvFrameAlloc()
		if utilFrame == nil {
			fmt.Println("frame alloc error")
			os.Exit(1)
		}
		defer avutil.AvFrameFree(utilFrame)

		audio_out_buff := avutil.AvMalloc(19200 * 2)

		fmt.Println("SampleRate,ChannelLayout:", pCodecCtx.SampleRate(), pCodecCtx.ChannelLayout())
		swrCtx := swresample.SwrAlloc()
		swrCtx.SwrAllocSetOpts(avutil.AV_CH_LAYOUT_STEREO,
			swresample.AV_SAMPLE_FMT_S16,
			pCodecCtx.SampleRate(),
			pCodecCtx.ChannelLayout(),
			(swresample.AvSampleFormat)(pCodecCtx.SampleFmt()),
			pCodecCtx.SampleRate(),
			0, 0)
		swrCtx.SwrInit()

		//var gotName int
		index := 0
		for inCtx.AvReadFrame(pkt) >= 0 {
			index += 1
			//fmt.Println("%%:", index)
			if pkt.StreamIndex() == audioStreamIndex {
				response := pCodecCtx.AvcodecSendPacket(pkt)
				if response < 0 {
					fmt.Printf("Error while sending a packet to the decoder: %s\n", avutil.ErrorFromCode(response))
				}
				if response >= 0 {
					response = pCodecCtx.AvcodecReceiveFrame((*avcodec.Frame)(unsafe.Pointer(utilFrame)))

				}

				if response == avutil.AvErrorEAGAIN || response == avutil.AvErrorEOF {
					fmt.Println("err:response:", response)
					break
				} else if response < 0 {
					fmt.Printf("Error while receiving a frame from the decoder: %s\n", avutil.ErrorFromCode(response))
					break
				}
				data := avutil.Data(utilFrame)
				p := (*uint8)(audio_out_buff)
				nb := avutil.NbSamples(utilFrame)
				ll := swrCtx.SwrConvert(&p, 19200, (**uint8)(unsafe.Pointer(&data[0])), int(nb))
				//l := pCodecCtx.AvcodecDecodeAudio4((*avcodec.Frame)(unsafe.Pointer(utilFrame)), &gotName, pkt)
				//fmt.Println("AvcodecDecodeAudio4:", l)
				//if l < 0 {
				//	fmt.Println("codec decode audio4 error")
				//	os.Exit(1)
				//}
				//if gotName > 0 {
				//	fram := getFramBytes(utilFrame)
				//	fmt.Println("buf add:", index)
				//	buffer <- fram
				//
				//}
				//fmt.Println(ll)
				pix := []byte{}
				startPos := uintptr(unsafe.Pointer(p))
				for i := 0; i < ll; i++ {
					e := *(*uint8)(unsafe.Pointer(startPos + uintptr(i)))
					pix = append(pix, e)
				}
				//fmt.Println(*(*string)(unsafe.Pointer(&pix)))
				buffer <- pix

			}
			pkt.AvFreePacket()
		}
		go func() {
			for {
				if len(buffer) <= 0 {
					fmt.Println("close buf")
					close(buffer)
					break
				}
			}
		}()

		(*avcodec.Context)(unsafe.Pointer(pCodecCtxOrig)).AvcodecClose()
	}()
	return buffer, nil
}

func getFramBytes(f *avutil.Frame) []byte {
	data := avutil.Data(f)
	//var bs = make([]byte, 0, len(data))
	var bf = make([]byte, len(data))
	for i := 0; i < len(data); i++ {

		if data[i] != nil {
			bf = append(bf, *data[i])
		}
	}
	return bf
}
func main() {

	//portaudio.Initialize()
	//defer portaudio.Terminate()
	//out := make([]uint8, 8192)
	//stream, err := portaudio.OpenDefaultStream(0, 1, 48000, len(out), &out)
	//defer stream.Close()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//err = stream.Start()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//defer stream.Stop()
	buf, err := audio()
	if err != nil {
		fmt.Println(err)
		return
	}
	index := 0
	c, err := oto.NewContext(44100, 2, 2, 8192)
	if err != nil {
		return
	}
	defer c.Close()

	p := c.NewPlayer()
	defer p.Close()
	for {
		select {
		case frame, ok := <-buf:
			if !ok {
				os.Exit(0)
			}
			index += 1
			//fmt.Println("$$:", index)
			if _, err := io.Copy(p, bytes.NewReader(frame)); err != nil {
				fmt.Println(err)
				return
			}

			//err := binary.Read(bytes.NewReader(frame), binary.BigEndian, out)
			//if err != nil {
			//	fmt.Println("binary.Read:", err)
			//	os.Exit(0)
			//}
			//err = stream.Write()
			//if err != nil {
			//	fmt.Println("stream.Write:", err)
			//	os.Exit(0)
			//}
		}
	}

}
