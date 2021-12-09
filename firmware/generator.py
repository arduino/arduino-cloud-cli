#!/usr/bin/env python3

import os
import shutil
import json
import hashlib
import sys
from pathlib import Path
import argparse
import subprocess

DOWNLOAD_URL = "https://cloud-downloads.arduino.cc"

PROVISION_BINARY_PATHS = {
    "lora": "binaries/provision/lora",
    "crypto": "binaries/provision/crypto",
}

SKETCH_NAMES = {
    "lora": "LoraProvision",
    "crypto": "CryptoProvision"
}

INDEX_PATH = "binaries/index.json"

BOARDS = [
    {"type": "crypto", "ext": ".bin", "fqbn": "arduino:samd:nano_33_iot"},
    {"type": "crypto", "ext": ".bin", "fqbn": "arduino:samd:mkrwifi1010"},
    {"type": "crypto", "ext": ".elf", "fqbn": "arduino:mbed_nano:nanorp2040connect"},
    {"type": "crypto", "ext": ".bin", "fqbn": "arduino:mbed_portenta:envie_m7"},
    {"type": "crypto", "ext": ".bin", "fqbn": "arduino:samd:mkr1000"},
    {"type": "crypto", "ext": ".bin", "fqbn": "arduino:samd:mkrgsm1400"},
    {"type": "crypto", "ext": ".bin", "fqbn": "arduino:samd:mkrnb1500"},
    {"type": "lora", "ext": ".bin", "fqbn": "arduino:samd:mkrwan1300"},
    {"type": "lora", "ext": ".bin", "fqbn": "arduino:samd:mkrwan1310"},
]

# Generates file SHA256
def sha2(file_path):
    with open(file_path, "rb") as f:
        return hashlib.sha256(f.read()).hexdigest()

# Runs arduino-cli
def arduino_cli(cli_path, args=None):
    if args is None:
        args=[]
    res = subprocess.run([cli_path, *args], capture_output=True, text=True, check=True)
    return res.stdout, res.stderr

def provision_binary_details(board):
    bin_path = PROVISION_BINARY_PATHS[board["type"]] 
    simple_fqbn = board["fqbn"].replace(":", ".")
    sketch_dir = Path(__file__).parent / bin_path / simple_fqbn
    sketch_files = list(sketch_dir.iterdir())
    # there should be only one binary file
    if len(sketch_files) != 1:
        print(f"Invalid binaries found in {sketch_dir}")
        sys.exit(1)
    sketch_file = sketch_files[0]  

    sketch_dest = f"{bin_path}/{simple_fqbn}/{sketch_file.name}"
    file_hash = sha2(sketch_file)

    return {
        "url": f"{DOWNLOAD_URL}/{sketch_dest}",
        "checksum": f"SHA-256:{file_hash}",
        "size": f"{sketch_file.stat().st_size}",
    }

def generate_index(boards):
    index_json = {"boards": []}
    for board in boards:
        index_board = {"fqbn": board["fqbn"]}
        index_board["provision"] = provision_binary_details(board)
        index_json["boards"].append(index_board)

    p = Path(__file__).parent / INDEX_PATH
    with open(p, "w") as f:
        json.dump(index_json, f, indent=2)

def generate_binaries(arduino_cli_path, boards):
    for board in boards:
        sketch_path = Path(__file__).parent / "provision" / SKETCH_NAMES[board["type"]]
        print(f"Compiling for {board['fqbn']}")
        res, err = arduino_cli(arduino_cli_path, args=[
            "compile", 
            "-e", 
            "-b", board["fqbn"], 
            sketch_path,
        ])
        print(res, err)
        simple_fqbn = board["fqbn"].replace(":", ".")
        # Make output directory
        out = Path(__file__).parent / PROVISION_BINARY_PATHS[board["type"]] / simple_fqbn
        os.makedirs(out, exist_ok=True)
        # Copy the new binary file in the correct output directory
        compiled_bin = sketch_path / "build" / simple_fqbn / (SKETCH_NAMES[board["type"]] + ".ino" + board["ext"])
        shutil.copy2(compiled_bin, out / ("provision" + board["ext"]))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(prog="generator.py")
    parser.add_argument(
        "-a",
        "--arduino-cli",
        default="arduino-cli",
        help="Path to arduino-cli executable",
    )
    args = parser.parse_args(sys.argv[1:])
    generate_binaries(args.arduino_cli, BOARDS)
    generate_index(BOARDS)
