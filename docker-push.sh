#!/bin/bash
docker build . -t asia.gcr.io/accelia-dev/panoramic_image-2-thumbnail:0.1.0 &&
docker push asia.gcr.io/accelia-dev/panoramic_image-2-thumbnail:0.1.0
