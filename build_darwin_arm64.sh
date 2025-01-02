rm -r bin
GOOS=darwin GOARCH=arm64 go build -o bin/mezmer

plutil -convert xml1 ./resources/entitlements.plist

go run macappgo/macapp.go \
  -assets bin \
  -bin  mezmer \
  -icon resources/icon.png \
  -identifier com.beringresearch.Mezmer \
  -name "Mezmer" \
  -o app

cp ./resources/Info.plist ./app/Mezmer.app/Contents

APP_PATH="app/Mezmer.app"
APP_CERTIFICATE="Developer ID Application: Ignat Drozdov (27B2YLEUVR)"
codesign --timestamp --options=runtime -s "$APP_CERTIFICATE" -v --entitlements ./resources/entitlements.plist "$APP_PATH"

ZIP_PATH="Mezmer_Darwin_arm64.zip"

# Create a ZIP archive suitable for notarization.
/usr/bin/ditto -c -k --keepParent "$APP_PATH" "$ZIP_PATH"

xcrun notarytool submit "$ZIP_PATH" --keychain-profile "notarytool-password" --wait

unzip $ZIP_PATH
rm $ZIP_PATH
xcrun stapler staple "Mezmer.app"

mkdir dist
mv Mezmer.app dist/
/usr/bin/ditto -c -k --keepParent "dist/Mezmer.app" "dist/$ZIP_PATH"

rm -r app bin
rm Mezmer.zip
sudo rm -r dist/Mezmer.app