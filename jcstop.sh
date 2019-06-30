#!/bin/bash
export JAVA_HOME=JAVA_HOME=/home/denis/app/jdk-11.0.2
export PATH=$JAVA_HOME/bin:$PATH
export JOB_CONTROL=/home/denis

cd /home/denis/mca/CI/jobcontrol
./jobcontrol stop  --server localhost --profile birdy-dev > /dev/null

