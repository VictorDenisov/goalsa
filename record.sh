#!/bin/bash

arecord -f S32_LE --device="plughw:0,0" test-mic.wav
