#!/bin/bash
################################################################
#   prerequest
#   sudo yum install lsof

################################################################
GEROCFG=~/.gerocfg
DEFAULT_DATADIR=~/.datadir
PATTERN_ATTACH_PROCESS="gero.*attach"
PATTERN_MAIN_PROCESS=".*gero.*datadir.*port.*"
PATTERN_ACCOUNTS="\[*\]"
PATTERN_QUERY_ACCOUNTS="ser.*accounts"
PATTERN_NEW_ACCOUNTS="personal.*newAccounts()"
PATTERN_NEW_ACCOUNTS="personal.*newAccounts()"
PATTERN_NEW_PASSPHRASE="Passphrase:"
PATTERN_FATAL="Fatal.*"
PATTERN_NEW_REPEAT_PASSPHRASE=".*Repeat.*"
PATTERN_PASSPHRASE_WARN=".*Unsupported terminal.*"
PATTERN_PASSWORD="[/w]+"
PATTERN_ACCOUNT_CREATED="\[\".*\"\]"
PATTERN_ACCOUNT_JUSTCREATED="^\".*\""
declare -a GLOBAL_ARRAY
declare -a BEGIN_PASSWORD
declare -a REPEAT_PASSWORD
declare -a END_PASSWORD
declare -a NEWACCOUNT
export PASSWORD=""
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:`pwd`/lib:`pwd`/czero/lib
loadDefaultParams() {
   if [ ! -d "$DEFAULT_DATADIR" ]; then
       mkdir -p ~/.datadir
   fi
   echo "export DATADIR=~/.datadir"	> $GEROCFG
   echo "export SERVERPORT=60601" >> $GEROCFG 
   echo "export RPCPORT=8545" >> $GEROCFG 
}
killProcess() {
    if [[ -z $1  ]]; then
        echo "please input the process pattern to kill" 
    fi
    result=`ps -ef | grep "$1" | grep -v grep | awk '{print $2}'|wc -l`
    echo "now there is $result process like:$1 are running"
    if [ $result -gt 0  ]; then 
        echo "to kill process with pattern:$1"
        ps -ef | grep $1 | grep -v grep | awk '{print $2}' | xargs kill -9
    fi


}
readCfg()  { 
    . ~/.gerocfg
    . ~/.bash_profile
}
	
#check the parameters available in env
checkEnvParas() {
	${DATADIR?"Need to set DATADIR"}
	${SERVERPORT?"Need to set SERVERPORT"}
	${RPCPORT?"Need to set RPCPORT"}
}


#make account password
makePasswd(){
    read -p "please input the account password:" -e input
    passwd1=${input}
    echo "${passwd1}" >&2
    read -p "please repeat the account password:" -e input
    passwd2=${input}
    echo "${passwd2}" >&2
    if [ "x${passwd1}" != "x${passwd2}" ]; then
        echo "passwd inconsist" >&2
        echo "-1"
        return -1
    fi
    export PASSWORD="${passwd1}"
    echo "now password:${PASSWORD}" >&2
    echo "${PASSWORD}"
    return 0
}
#check datadir path is correct
checkDataDir() {
   if [ -f $1 ]; then
       echo "the datadir is actually a file:$1"
       return -1
   elif [ ! -d $1  ]; then
       echo "$1 for datadir is not exsist, now create it"
       mkdir -p "$1"
   fi
   return 0
}

#check port used
checkPortUsed(){
    portToCheck=$1
    if [ -z $portToCheck ]; then echo "please input port to check" >&2
    fi
    if lsof -Pi :$portToCheck -sTCP:LISTEN -t >/dev/null ; then
        echo "used" 
    else
        echo "notused" 
    fi
}

verifyServerPort(){
    portToCheck=$1
    portName=$2
    while [ "x$(checkPortUsed $portToCheck)" == "xused"  ]
    do 
        read -p "the $portName:${portToCheck} is used, we will choose random port not used as the SERVERPORT[no|*]" Syn
        case $Syn in
            [Nn]* ) 
                read -p "answer is no, input the $2 you want to use:" -e port
                portToCheck=$port        
                continue;;
             * ) echo "we will check next available port to use";let "portToCheck=${portToCheck}+1"; continue;;
        esac 
    done
    if [ $portToCheck != $1 ]; then
        echo "find port:$portToCheck for $2 can be used"
        set -o xtrace
        sed -i -e "s/${portName}=$1/${portName}=$portToCheck/g" ~/.gerocfg
        set +x
        export ${portName}=${portToCheck}
    fi
}
stringLength(){
    str=$1
    length=${#str}
#    echo ${length} >&2
    return ${length}
}
checkProcessWorking(){
    if [ $(ps -ef | grep $1 | grep -v grep | wc -l) -gt 0 ]; then
        echo "1"
    fi
    echo "-1"
}

checkProcess(){
    if [ $(ps -ef | grep $1 | grep -v grep | wc -l) -gt 0 ]; then
        read -p "there is already a process like:$1 running, do you want to stop it:[*|no]" -e ans
        case $ans in
            [Nn]* )
                echo "the previous process is still working" >&2
                echo "0"
                return 0;;
            * ) 
                echo "the old process like:$1 will be stopped" >&2
                killProcess "$1" >&2
                sleep 15
                echo "-1"
                return 0;;
        esac
    fi
    echo "-1"
}

makeAccount(){
    if [ -z "$1" ]; then
        echo "please provide password to create account"
        return -1
    fi
    echo "personal.newAccount()" > /tmp/gero-input
    echo "${PASSWORD}" > /tmp/gero-input
    echo "${PASSWORD}" > /tmp/gero-input
    while read lineA; do
        if [ -z ${lineA} ]; then
            echo "finished" >&2
            return 1
        fi
        echo "$lineAn"
    done </tmp/gero-output &
}  
 
readingFromOutput(){
    while true
    do 
        if read lineA; then
            echo "${lineA}"
            LINEA_NO_WHITESPACE="$(echo -e "${lineA}" | tr -d '[:space:]')"
            stringLength ${LINEA_NO_WHITESPACE}
            length=$?
            # echo "trimed:$LINEA_NO_WHITESPACE==${length}"
            if [ $length == 1 -a "${LINEA_NO_WHITESPACE}" == ">" ]  ; then
             #    echo "it is a hint" >&2
                continue 
            elif [ "$LINEA_NO_WHITESPACE" == '[]' ]; then
                echo "NEWACCOUNT:" >./localio
            elif [[ "$LINEA_NO_WHITESPACE" =~ ${PATTERN_FATAL} ]]; then
                echo "meet fatal error, please check the log in ~/working/gero.log" >&2
                exit -1
            fi
            if [[ "$lineA" == $PATTERN_NEW_PASSPHRASE ]]; then 
              #  echo "wait for password:">&2
                BEGIN_PASSWORD="true"
                echo "BEGIN_PASSWORD:true" >./localio
                #echo "now input ${PASSWORD} directlly"
                #printf '%s\n' "${PASSWORD}" >/tmp/gero-input &
                sleep 5
            elif [[ "$lineA" =~ ${PATTERN_NEW_REPEAT_PASSPHRASE} ]]; then  
              #  echo "repeat password:">&2
                REPEAT_PASSWORD="true" 
                #echo "now repeat input ${PASSWORD} directlly"
                echo "REPEAT_PASSWORD:${REPEAT_PASSWORD}" >./localio
                #printf '%s\n' "${PASSWORD}" >/tmp/gero-input &
                sleep 5
            elif [[ "$lineA" =~ ${PATTERN_ACCOUNT_CREATED} ]]; then  
                localLine=${lineA}
                echo "account created $localLine" >&2
                #NEWACCOUNT=$(echo $lineA |  grep -o $PATTERN_ACCOUNT_CREATED | sed -e 's/"//g')
                NEWACCOUNT=$(echo $localLine| sed -e 's/"//g')
                echo "NEWACCOUNT:$NEWACCOUNT" >./localio
            elif [[ "$lineA" =~ ${PATTERN_ACCOUNT_JUSTCREATED} ]]; then
                echo "Account just created $lineA" >&2
                NEWACCOUNT=$(echo $lineA| sed -e 's/"//g')
                echo "NEWACCOUNT:$NEWACCOUNT" >./localio
            elif [[ "$lineA" =~ ${PATTERN_PASSPHRASE_WARN} ]]; then
                echo "warning"
                printf '%s\n' "${PASSWORD}" >/tmp/gero-input &
            fi
            
        fi
    done </tmp/gero-output
}
checkOutputAndWorkNex() {
    while true
    do
        if read lineB; then
            echo "checkoutput $lineB"
            if [ "$lineB" == "BEGIN_PASSWORD:true" ]; then
                sleep 3
                printf '%s\n' "${PASSWORD}" >/tmp/gero-input & 
              #  echo "${PASSWORD}"
                echo "waiting for repeat password"
            elif [ "$lineB" == "REPEAT_PASSWORD:true" ]; then
                sleep 3
                printf '%s\n' "${PASSWORD}" >/tmp/gero-input & 
              #  echo "${PASSWORD}"
                echo "waiting for account created"
            elif [[ "$lineB" == "NEWACCOUNT:"* ]]; then
                sleep 3
                echo "lineB:$lineB"
                NEWACCOUNT=$(echo $lineB |  grep -o "NEWACCOUNT:.*" | sed -e 's/NEWACCOUNT\://g' | sed -e 's/\[//g'|sed -e 's/\]//g')
                ACCOUNT=$(getAccount $NEWACCOUNT)
                echo "out [${ACCOUNT}]"
                # if account length is 0 then make new account 
                # else output the new accout to localio
                stringLength ${NEWACCOUNT}
                length=$?

                if [ $length -le 10 ]; then
                    echo "empty accounts, need make new account"
                    printf '%s\n' "personal.newAccount()" >/tmp/gero-input
                    sleep 5
                    continue 
                fi 
                result=$(checkStringHaveSpecial ${ACCOUNT})
                if [ "${result}" == "true" ]; then
                    echo "it is not correct account"
                    continue
                fi
                echo "account created:[$ACCOUNT] begin to mine"
                printf 'miner.setSerobase("%s")\n' "${ACCOUNT}" 
                printf 'miner.setSerobase("%s")\n' "${ACCOUNT}" >/tmp/gero-input &
                echo 'miner.start()' >/tmp/gero-input &
                exit 
            elif [[ "$lineB" == "warning" ]]; then
                printf '%s\n' "${PASSWORD}" >/tmp/gero-input &
            fi  
        fi
    done <./localio
}
checkStringHaveSpecial(){
    if [[ "$1" == *['!'@#\$%^\&*()_+\ ]*  ]]; then
       echo "true"
    else 
        echo "false"
    fi
} 
getAccount() {
    echo "begin getAccount $1"  >&2
    accounts=$(echo "$1"| cut -d "[" -f2 | cut -d "]" -f1)
    GLOBAL_ARRAY=($(echo $1|tr "," "\n"))
    echo "${GLOBAL_ARRAY[0]}"
    #GLOBAL_ARRAY=( "${array[@]}" )
#    for element in "${global_array[@]}"
#    do
#        echo "inside global_array:$element" >&2
#    done
}
# begin


pkill cat &
killProcess ${PATTERN_ATTACH_PROCESS}
if [ -f ${GEROCFG} ]; then
    readCfg ${GEROCFG}
else
	touch $GEROCFG
	chmod 755 ${GEROCFG}
	loadDefaultParams
    readCfg ${GEROCFG}
    echo "default parameters loaded"
fi
OLDDATADIR=${DATADIR}
read -p "If you don't accept config datadir as ${DATADIR} , please input the datadir here:" -e input
DATADIR=${input}
if [ -z "$DATADIR" ]; then
    echo "We will use default ${OLDDATADIR} for your datadir"
    DATADIR=${OLDDATADIR}
else
    while [ "x$(checkDataDir ${DATADIR})" == "x0" ]; do
        read -p "the ${DATADIR} you input is not valid, please input correct path as datadir here:" -e input
        DATADIR=${input}
    done
fi
PASSWORD=$(makePasswd)
#echo "result:${result}"
while ! [[ "$?" == "0" ]] 
do 
    read -p "do you still want to create miner[挖矿账号] account now?[yes|no]:" Syn
    case $Syn in
        [Yy]* ) echo "answer is yes, so you need set password to create account";result=$(makePasswd); continue;;
        [Nn]* ) exit;;
        * ) echo "Please answer yes or no.";;
    esac 
done
result=$(checkProcess "$PATTERN_MAIN_PROCESS")
echo ${result}
if [ "${result}" == "-1" ]; then
    echo    "now begin to start main proces"
    verifyServerPort $SERVERPORT SERVERPORT
    verifyServerPort $RPCPORT RPCPORT
    set -o xtrace
    nohup ./bin/gero   --datadir=${DATADIR} --rpc --rpcport ${RPCPORT}  --port ${SERVERPORT} --rpccorsdomain "*" > gero.log 2>&1 &
    set +x
    sleep 30
fi
result=$(checkProcessWorking "$PATTERN_MAIN_PROCESS")
if [ "${result}" == "-1" ]; then
    echo "Please check why the main process of gero not running with log in ~/working/"
    exit
fi 
if [ -p "/tmp/gero-input" ] || [ -f "/tmp/gero-input" ]; then
    rm /tmp/gero-input
fi
if [ -p "/tmp/gero-output" ] || [ -f "/tmp/gero-output" ]; then
    rm /tmp/gero-output
fi
if [ -p "./localio" ] || [ -f "./localio" ]; then
    rm ./localio
fi
mkfifo -Z --mode='a=rwx' /tmp/gero-input
mkfifo -Z --mode='a=rwx' /tmp/gero-output
mkfifo -Z --mode='a=rwx' ./localio
cat /tmp/gero-input|./bin/gero --datadir=~/.datadir attach > /tmp/gero-output &
exec 9<> /tmp/gero-input
readingFromOutput &
backgroundPid=$!
printf '%s\n' 'sero.accounts' >/tmp/gero-input
BEGIN_PASSWORD=false
REPEAT_PASSWORD=false
NEWACCOUNT=""

#echo 'personal.newAccount()' >/tmp/gero-input&

checkOutputAndWorkNex &
anotherBackgroundPid=$!
wait ${anotherBackgroundPid}
