#!/bin/bash

aplay -t raw -f S16_LE -c1 -r44100 $1
