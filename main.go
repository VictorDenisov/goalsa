package main

import (
	"fmt"
	"math/cmplx"
	"os"
	"unsafe"

	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
	log "github.com/sirupsen/logrus"
)

/*
#cgo LDFLAGS: -lasound
#include <alsa/asoundlib.h>
#include <stdint.h>
*/
import "C"

var filter []complex128

func init() {
	filter = dsputils.ZeroPad([]complex128{5}, 5)
}

// This code captures standard input
// Redirect this files standard output into a file.
// Play the file using the following command:
// aplay -t raw -f S16_LE -c1 -r44100 $1
// 1 channel
// little endian
// 16 bit
// 44100 sampling rate
//
// Import raw data using audacity with the specified parameters
func main() {
	testFft()

	return

	var handle *C.snd_pcm_t
	var rc C.int

	deviceCString := C.CString("default")
	defer C.free(unsafe.Pointer(deviceCString))

	rc = C.snd_pcm_open(&handle, deviceCString, C.SND_PCM_STREAM_CAPTURE, 0)
	if rc < 0 {
		log.Fatal("Unable to open pcm device: %v", C.snd_strerror(rc))
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
	var loops C.long = 10_000_000 / C.long(val)

	out, err := os.Create("output.raw")
	if err != nil {
		log.Fatalf("Couldn't create output file: %v", err)
	}
	defer out.Close()
	for loops > 0 {
		loops--
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

func testFft() {
	data := []complex128{1, 2, 3, 4, 5}
	fmt.Printf("data: %v\n", data)

	spectrum := fft.FFT(data)

	fmt.Printf("spectrum: \n")
	for i := 0; i < len(spectrum); i++ {
		fmt.Printf("%d %0.5f\n", i, cmplx.Abs(spectrum[i]))
	}

	result := fft.Convolve(filter, data)
	fmt.Printf("result: \n")
	for i := 0; i < len(result); i++ {
		fmt.Printf("%0.5f ", cmplx.Abs(result[i]))
	}
	fmt.Printf("\n")

	for i := 0; i < len(result); i++ {
		/*
			if imag(result[i]) < 0.00001 {
				fmt.Printf("%0.5d ", real(result[i]))
			} else {
				fmt.Printf("%0.5d ", real(result[i]))
			}
		*/
		fmt.Printf("(%0.5f, %0.5f) ", real(result[i]), imag(result[i]))
	}

}
