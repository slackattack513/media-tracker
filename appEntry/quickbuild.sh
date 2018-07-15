#! /bin/bash
cd ../youtubecontent
go install
cd ../appEntry/
go build
./appEntry