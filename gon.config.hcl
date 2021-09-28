# Source: https://github.com/arduino/tooling-project-assets/blob/main/workflow-templates/assets/general/gon.config.hcl
# See: https://github.com/mitchellh/gon#configuration-file
source = ["dist/arduino-cloud-cli_osx_darwin_amd64/arduino-cloud-cli"]
bundle_id = "cc.arduino.arduino-cloud-cli"

sign {
  application_identity = "Developer ID Application: ARDUINO SA (7KT7ZWMCJT)"
}

# Ask Gon for zip output to force notarization process to take place.
# The CI will ignore the zip output, using the signed binary only.
zip {
  output_path = "unused.zip"
}