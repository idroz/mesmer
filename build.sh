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
codesign --timestamp --options=runtime -s "$APP_CERTIFICATE" -v --entitlements ./resources/entitlements.plist ./app/Mezmer.app

ZIP_PATH="Mezmer.zip"

# Create a ZIP archive suitable for notarization.
/usr/bin/ditto -c -k --keepParent "$APP_PATH" "$ZIP_PATH"


xcrun notarytool submit Mezmer.zip --keychain-profile "notarytool-password" --wait

unzip Mezmer.zip
rm Mezmer.zip
xcrun stapler staple "Mezmer.app"

mkdir dist
mv Mezmer.app dist/
/usr/bin/ditto -c -k --keepParent "dist/Mezmer.app" "dist/Mezmer.zip"

rm -r app bin
sudo rm -r dist/Mezmer.app