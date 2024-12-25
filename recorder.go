package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"unsafe"

	log "github.com/sirupsen/logrus"
)

/*
#cgo LDFLAGS: -lasound
#include <alsa/asoundlib.h>
#include <stdint.h>
*/
import "C"

func record(fileName string, device string) {
	done := setupSignalHandling()

	var handle *C.snd_pcm_t
	var rc C.int

	deviceCString := C.CString(device)
	defer C.free(unsafe.Pointer(deviceCString))

	rc = C.snd_pcm_open(&handle, deviceCString, C.SND_PCM_STREAM_CAPTURE, 0)
	if rc < 0 {
		log.Fatalf("Unable to open pcm device: %v", C.snd_strerror(rc))
	}
	defer C.snd_pcm_close(handle)
	defer C.snd_pcm_drain(handle)

	// Setup params
	var params *C.snd_pcm_hw_params_t
	rc = C.snd_pcm_hw_params_malloc(&params)
	if rc < 0 {
		log.Fatal("Couldn't alloc hw params")
	}
	defer C.snd_pcm_hw_params_free(params)
	rc = C.snd_pcm_hw_params_any(handle, params)
	if rc < 0 {
		log.Fatal("Couldn't set default hw params")
	}
	rc = C.snd_pcm_hw_params_set_access(handle, params, C.SND_PCM_ACCESS_RW_INTERLEAVED)
	if rc < 0 {
		log.Fatal("Couldn't set access params")
	}
	rc = C.snd_pcm_hw_params_set_format(handle, params, C.SND_PCM_FORMAT_S16_LE)
	if rc < 0 {
		log.Fatal("Couldn't set endian format")
	}
	rc = C.snd_pcm_hw_params_set_channels(handle, params, 1)
	if rc < 0 {
		log.Fatal("Couldn't set channels")
	}
	var val C.uint = 44100
	var dir C.int
	rc = C.snd_pcm_hw_params_set_rate_near(handle, params, &val, &dir)
	if rc < 0 {
		log.Fatal("Couldn't set rate")
	}

	var frames C.snd_pcm_uframes_t = 32
	rc = C.snd_pcm_hw_params_set_period_size_near(handle, params, &frames, &dir)
	if rc < 0 {
		log.Fatal("Couldn't set period size")
	}

	rc = C.snd_pcm_hw_params(handle, params)
	if rc < 0 {
		log.Fatal("Couldn't set params")
	}

	rc = C.snd_pcm_hw_params_get_period_size(params, &frames, &dir)
	if rc < 0 {
		log.Fatal("Couldn't get period size")
	}

	var size C.snd_pcm_uframes_t
	log.Tracef("Frames: %v\n", frames)
	size = frames * 2
	buffer := make([]byte, size)
	log.Tracef("Buffer len: %v", len(buffer))

	rc = C.snd_pcm_hw_params_get_period_time(params, &val, &dir)
	if rc < 0 {
		log.Fatal("Couldn't get period time")
	}

	out, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Couldn't create output file: %v", err)
	}
	defer out.Close()
loop:
	for {
		select {
		case <-done:
			out.Close()
			break loop
		default:
		}

		var rcl C.long
		rcl = C.snd_pcm_readi(handle, unsafe.Pointer(&buffer[0]), frames)
		if rcl == -C.EPIPE {
			fmt.Printf("Overrun occurred\n")
			C.snd_pcm_prepare(handle)
		} else if rcl < 0 {
			fmt.Printf("Error from read: %v\n", C.snd_strerror(rc))
		} else if rcl != C.long(frames) {
			fmt.Printf("Short read, read %v frames\n", rc)
		}
		log.Tracef("rcl read: %v\n", rcl)
		n, err := out.Write(buffer)
		if err != nil && n < len(buffer) {
			fmt.Printf("Failed to write: %v\n", err)
		}
	}
}

func setupSignalHandling() chan bool {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	return done
}
