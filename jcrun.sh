#!/bin/bash
export JAVA_HOME=JAVA_HOME=/home/denis/app/jdk-11.0.2
export PATH=$JAVA_HOME/bin:$PATH
export JOB_CONTROL=/home/denis

cd /home/denis/mca/CI/jobcontrol
./jobcontrol run --server localhost --port 9999  --profile birdy-dev  --project /home/denis/birdy-server/ --file birdy.toml > /dev/null

