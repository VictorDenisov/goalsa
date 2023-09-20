package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/mjibson/go-dsp/fft"
	log "github.com/sirupsen/logrus"
)

/*
#cgo LDFLAGS: -lasound
#include <alsa/asoundlib.h>
#include <stdint.h>
*/
import "C"

type Window struct {
	a []int16
}

func NewWindow() *Window {
	return &Window{make([]int16, 0)}
}

func (w *Window) Add(v int16) {
	w.a = append(w.a, v)
}

func stream(ctx context.Context, device string) error {
	as, err := OpenAudioStream(device)
	if err != nil {
		return err
	}
	rawChan := as.GetChan()
	filteredChan := filterSignal(rawChan)
	spectraChan := produceSpectra(filteredChan)
	textChan := decode(spectraChan)
	for {
		select {
		case <-ctx.Done():
			return nil
		case _ = <-textChan:
		}
	}
}

type AudioStream struct {
	handle *C.snd_pcm_t
	ch     chan int16
}

func OpenAudioStream(device string) (*AudioStream, error) {
	as := &AudioStream{nil, make(chan int16)}

	var rc C.int

	log.Tracef("Openning device: %v", device)
	deviceCString := C.CString(device)
	defer C.free(unsafe.Pointer(deviceCString))

	rc = C.snd_pcm_open(&as.handle, deviceCString, C.SND_PCM_STREAM_CAPTURE, 0)
	if rc < 0 {
		return nil, fmt.Errorf("Unable to open pcm device: %v", C.snd_strerror(rc))
	}

	var params *C.snd_pcm_hw_params_t
	rc = C.snd_pcm_hw_params_malloc(&params)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't alloc hw params")
	}
	defer C.snd_pcm_hw_params_free(params)

	rc = C.snd_pcm_hw_params_any(as.handle, params)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't set default hw params")
	}
	rc = C.snd_pcm_hw_params_set_access(as.handle, params, C.SND_PCM_ACCESS_RW_INTERLEAVED)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't set access params")
	}
	rc = C.snd_pcm_hw_params_set_format(as.handle, params, C.SND_PCM_FORMAT_S16_LE)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't set endian format")
	}
	rc = C.snd_pcm_hw_params_set_channels(as.handle, params, 1)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't set channels")
	}
	var val C.uint = 44100
	var dir C.int
	rc = C.snd_pcm_hw_params_set_rate_near(as.handle, params, &val, &dir)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't set rate")
	}

	var frames C.snd_pcm_uframes_t = 8192
	rc = C.snd_pcm_hw_params_set_period_size_near(as.handle, params, &frames, &dir)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't set period size")
	}

	rc = C.snd_pcm_hw_params(as.handle, params)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't set params")
	}

	rc = C.snd_pcm_hw_params_get_period_size(params, &frames, &dir)
	if rc < 0 {
		return nil, fmt.Errorf("Couldn't get period size")
	}

	var size C.snd_pcm_uframes_t
	log.Tracef("Frames: %v\n", frames)
	size = frames * 2 // Period - 2
	buffer := make([]byte, size)
	log.Tracef("Buffer len: %v", len(buffer))

	rc = C.snd_pcm_hw_params_get_period_time(params, &val, &dir)
	if rc < 0 {
		log.Fatal("Couldn't get period time")
	}

	go func() {
		for {
			var rcl C.long
			rcl = C.snd_pcm_readi(as.handle, unsafe.Pointer(&buffer[0]), frames)
			log.Tracef("Received rcl: %v", rcl)
			if rcl == -C.EPIPE {
				fmt.Printf("Overrun occurred\n")
				C.snd_pcm_prepare(as.handle)
			} else if rcl < 0 {
				fmt.Printf("Error from read: %v\n", C.snd_strerror(rc))
			} else if rcl != C.long(frames) {
				fmt.Printf("Short read, read %v frames\n", rc)
			}
			bf := bytes.NewReader(buffer)
			var v int16
			for {
				err := binary.Read(bf, binary.LittleEndian, &v)
				log.Tracef("Pushing to channel value: %v", v)
				if err != nil {
					break
				}
				as.ch <- v
			}
		}
	}()

	return as, nil
}

func (as *AudioStream) GetChan() chan int16 {
	return as.ch
}

func (as *AudioStream) Read() int16 {
	return <-as.ch
}

func (as *AudioStream) Close() {
	C.snd_pcm_drain(as.handle)
	C.snd_pcm_close(as.handle)
}

func filterSignal(in chan int16) (out chan []float64) {
	out = make(chan []float64)
	filter := NewBpFilter(200, 7.0/441, 30.0/441, 441)
	go func() {
		for {
			buf := make([]float64, 441)
			for i := 0; i < 441; i++ {
				var v int16
				v = <-in
				buf[i] = float64(v)
			}
			buf = filter.FilterBuf(buf)
			out <- buf
		}
	}()
	return out
}

func produceSpectra(in chan []float64) (out chan []float64) {
	out = make(chan []float64)
	go func() {
		for {
			buf := <-in
			hann(buf)
			rawSpectrum := ToAbs(fft.FFTReal(buf))
			out <- rawSpectrum[0:222]
		}
	}()
	return out
}

func decode(ch chan []float64) (out chan string) {
	m := 222
	spectra := make([][]float64, 0)
	sum := make([]float64, m)
	n := 0
	for {
		sp := <-ch
		spectra := append(spectra, sp)
		n++
		sumV(sum, sp)
		fmt.Printf("%v\n", spectra)
	}
}
