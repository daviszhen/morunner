#!/bin/bash


commit_id=""
last_commit_id=""
mo_runner_path=""

check_os(){
    os=`uname -s`
    echo $os
    if [[ "$os" == "Darwin" ]]
    then
        echo "Mac OS"
        mo_runner_path="/Users/pengzhen/Documents/GitHub/morunner"
    elif [[ "$os" == "Linux" ]]
    then
        echo "Linux"
        mo_runner_path="/root/morunner"
    else
        echo "Other OS: $os"
        echo "does not support"
        exit 1
    fi
}

update_code(){
    cd ${mo_runner_path}
    git pull
    commit_id=`git rev-parse HEAD`
    git log --pretty=format:"%an %cd %s" ${commit_id} | head -n 1
}

run(){
    cd ${mo_runner_path}
    go build -o morunner main.go types.go
    nohup ./morunner --loop --url freetier-01.cn-hangzhou.cluster.aliyun-dev.matrixone.tech --user dump --password "M-62kCmR0jwP" &
}

run_morunner(){
    ps -fe | grep "morunner" |grep -v grep
    if [ $? -ne 0 ]
    then
        echo "morunner is not runing....."
        update_code
        last_commit_id=$commit_id

        echo "start morunner....."
        run
    else
        echo "morunner is runing....."
        update_code

        if [[ "${commit_id}" != "${last_commit_id}" ]]
        then
            last_commit_id=$commit_id

            #kill morunner 
            #get process id of morunner
            echo "morunner code is updated, restart morunner....."
            echo "kill morunner first"
            morunner_proc_id=`ps -ef | grep morunner | grep -v grep |  awk '{print $2}'`
            kill -9 ${morunner_proc_id}

            echo "start morunner again....."
            run
            
        fi
    fi
    cd ~
}

main(){
    check_os
    a=1
    while [ $a -ne 0 ]
    do
        date
        echo "daemon running..."
        cd ~
        run_morunner
        sleep 60
    done
}

main