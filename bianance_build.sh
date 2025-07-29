#!/bin/bash
cd contrib/binancefeeder_v2
go build -buildmode=plugin -o ../../bin/binancefeeder_v2.so .
cd ../../
./marketstore start --config contrib/binancefeeder_v2/run_config.yml
