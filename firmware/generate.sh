# Exit if any of the following commands fails
set -e

# Prerequisite: arduino-cli should be already installed

# Install needed arduino cores
arduino-cli core update-index -v
arduino-cli version
arduino-cli core install arduino:samd
arduino-cli core install arduino:mbed_nano
arduino-cli core install arduino:mbed_portenta

# Install crypto dependencies
arduino-cli lib install ArduinoIotCloud
arduino-cli lib install ArduinoECCX08
arduino-cli lib install ArduinoSTL
arduino-cli lib install uCRC16Lib

# Install lora dependencies
arduino-cli lib install MKRWAN

# Compile in binaries folder

CRYPTO_FQBNS=" 
	arduino:samd:nano_33_iot 
	arduino:samd:mkrwifi1010
	arduino:mbed_nano:nanorp2040connect 
	arduino:mbed_portenta:envie_m7
	arduino:samd:mkr1000 
	arduino:samd:mkrgsm1400 
	arduino:samd:mkrnb1500
"	

LORA_FQBNS=" 
	arduino:samd:mkrwan1300
	arduino:samd:mkrwan1310
"

# Generate crypto provisioning binaries 
SKETCH_FOLDER="provisioning/CryptoProvision"
SKETCH_NAME="CryptoProvision"
OUTPUT_FOlDER="binaries/crypto"
mkdir -p $OUTPUT_FOlDER

for BOARD in $CRYPTO_FQBNS
do
	echo "compiling for $BOARD"
	arduino-cli compile -e -b $BOARD $SKETCH_FOLDER
	FORMATTED_BOARD=${BOARD//:/.}
	EXT=".bin"

	if [ $BOARD == "arduino:mbed_nano:nanorp2040connect" ]
	then
		EXT=".elf"
	fi

	cp $SKETCH_FOLDER/build/$FORMATTED_BOARD/$SKETCH_NAME.ino$EXT $OUTPUT_FOlDER/$FORMATTED_BOARD$EXT
done

# Generate lora provisioning binaries 
SKETCH_FOLDER="provisioning/LoraProvision"
SKETCH_NAME="LoraProvision"
OUTPUT_FOlDER="binaries/lora"
mkdir -p $OUTPUT_FOlDER

for BOARD in $LORA_FQBNS
do
	echo "compiling for $BOARD"
	arduino-cli compile -e -b $BOARD $SKETCH_FOLDER
	FORMATTED_BOARD=${BOARD//:/.}
	EXT=".bin"

	cp $SKETCH_FOLDER/build/$FORMATTED_BOARD/$SKETCH_NAME.ino$EXT $OUTPUT_FOlDER/$FORMATTED_BOARD$EXT
done

