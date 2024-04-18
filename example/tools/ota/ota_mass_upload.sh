#!/bin/bash

# This script is used to upload the firmware to the device using the OTA service.

export PATH=$PATH:.

checkExecutable () {
    if ! command -v $1 &> /dev/null
    then
        echo "$1 could not be found in PATH"
        exit 1
    fi
}

printHelp () {
    echo ""
    echo "Usage: $0 [-t <tag> | -d <device ids>] -f <firmwarefile> [-o <waittime in seconds - default 600>] [-v <new firmware version for tagging updated devices>]"
    echo ""
    echo "Examples -----------------"
    echo "  perform ota on devices with firmware=v1 tag"
    echo "    $0 -t firmware=v1 -f myfirmware.bin"
    echo "  perform ota on devices with firmware=v1 tag and apply new tag firmware=v2 to updated devices, waiting for 1200 seconds"
    echo "    $0 -t firmware=v1 -f myfirmware.bin -v firmware=v2 -o 1200"
    echo "  perform ota on two specified devices"
    echo "    $0 -d 261ec96a-38ba-4520-96e6-2447c4163e9b,8b10acdb-b722-4068-8e4d-d1c1b7302df4 -f myfirmware.bin"    
    echo ""
    exit 1
}

# Check dependencies...
checkExecutable "arduino-cloud-cli"
checkExecutable "jq"
checkExecutable "sort"
checkExecutable "uniq"
checkExecutable "paste"

# Default wait time for OTA process to complete
waittime=600
newtagversion=""

while getopts t:v:f:o:d: flag
do
    case "${flag}" in
        t) tag=${OPTARG};;
        v) newtagversion=${OPTARG};;
        f) firmwarefile=${OPTARG};;
        o) waittime=${OPTARG};;
        d) deviceids=${OPTARG};;
    esac
done

if [[ "$firmwarefile" == "" || "$waittime" == "" ]]; then
    printHelp    
fi
if [[ "$tag" == "" && "$deviceids" == "" ]]; then
    printHelp    
fi

if [[ "$deviceids" == "" ]]; then
    echo "Starting OTA process for devices with tag \"$tag\" using firmware \"$firmwarefile\""
    echo ""

    devicelistjson=$(arduino-cloud-cli device list --tags $tag --format json)
else
    echo "Starting OTA process for devices \"$deviceids\" using firmware \"$firmwarefile\""
    echo ""

    devicelistjson=$(arduino-cloud-cli device list -d $deviceids --format json)
fi

if [[ "$devicelistjson" == "" || "$devicelistjson" == "null" ]]; then
    echo "No device found"
    exit 1    
fi

devicecount=$(echo $devicelistjson | jq '.[] | .id' | wc -l)

if [ "$devicecount" -gt 0 ]; then
    echo "Found $devicecount devices"
    echo ""
    if [[ "$deviceids" == "" ]]; then
        arduino-cloud-cli device list --tags $tag
    else
        arduino-cloud-cli device list -d $deviceids
    fi
else 
    echo "No device found"
    exit 1    
fi

fqbncount=$(echo $devicelistjson | jq '.[] | .fqbn' | sort | uniq | wc -l)

if [ "$fqbncount" -gt 1 ]; then
    echo "Mixed FQBNs detected. Please ensure all devices have the same FQBN."
    fqbns=$(echo $devicelistjson | jq '.[] | .fqbn' | sort | uniq)
    echo "Detected FQBNs:"
    echo "$fqbns"
    exit 1    
fi

fqbn=$(echo $devicelistjson | jq -r '.[] | .fqbn' | sort | uniq | head -n 1)

echo "Sending OTA request to detected boards of type $fqbn..."
if [[ "$deviceids" == "" ]]; then
    otastartedout=$(arduino-cloud-cli ota mass-upload --device-tags $tag --file $firmwarefile -b $fqbn --format json)
else
    otastartedout=$(arduino-cloud-cli ota mass-upload -d $deviceids --file $firmwarefile -b $fqbn --format json)
fi
if [ $? -ne 0 ]; then
    echo "Detected error during OTA process. Exiting..."
    exit 1    
fi

otaids=$(echo $otastartedout | jq -r '.[] | .OtaStatus | .id' | uniq | paste -sd "," -)

if [ $otaids == "null" ]; then
    echo "No OTA processes to monitor. This could be due to an upgrade from previous ArduinoIotLibrary versions. Exiting..."
    exit 0
fi

correctlyfinished=0
while [ $waittime -gt 0 ]; do
    echo "Waiting for $waittime seconds for OTA process to complete..."
    sleep 15
    waittime=$((waittime-15))
    # Check status of running processess...
    otastatuslines=$(arduino-cloud-cli ota status --ota-ids $otaids --format json)
    otastatusinpcnt=$(echo $otastatuslines | grep in_progress | wc -l)
    otastatuspencnt=$(echo $otastatuslines | grep pending | wc -l)
    otasucceeded=$(echo $otastatuslines | jq -r '.[] | select (.status | contains("succeeded")) | .device_id' | uniq | paste -sd "," -)
    if [[ $otastatusinpcnt -eq 0 && $otastatuspencnt -eq 0 ]]; then
        correctlyfinished=1
        break   
    fi
done

echo ""
echo "Status report:"
arduino-cloud-cli ota status --ota-ids $otaids

if [ $correctlyfinished -eq 0 ]; then
    echo "OTA process did not complete within the specified time for some boards"
    exit 1
else 
    echo "OTA process completed successfully for all boards"
    if [ "$newtagversion" != "" ]; then
        echo "Tagging updated devices with tag: $newtagversion"
        arduino-cloud-cli device create-tags --ids $otasucceeded --tags $newtagversion
    fi
    exit 0
fi

exit 0
