source = ["./mariquita"]
bundle_id = "dev.zmoog.mariquita"

apple_id {
  username = "maurizio.branca@gmail.com"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "02B1797580ADB94948688199684FE9C75284D6D3"
}

zip {
  output_path = "./mariquita.zip"
}
