rm -r bin
GOOS=darwin GOARCH=arm64 go build -o bin/mezmer

plutil -convert xml1 ./resources/entitlements.plist
#codesign -vvv --deep --force --options=runtime --entitlements ./resources/entitlements.plist --sign "Developer ID Application: Ignat Drozdov (27B2YLEUVR)" --timestamp ./bin/mezmer
#codesign -v -vvv --strict --deep bin/mezmer

go run macappgo/macapp.go \
  -assets bin \
  -bin  mezmer \
  -icon resources/icon.png \
  -identifier com.beringresearch.Mezmer \
  -name "Mezmer" \
  -o app

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

APP_NAME="Mezmer"
DMG_FILE_NAME="${APP_NAME}.dmg"
VOLUME_NAME="${APP_NAME}"
SOURCE_FOLDER_PATH="dist/"


CREATE_DMG=create-dmg

# Create the DMG
$CREATE_DMG \
  --volname "${VOLUME_NAME}" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --icon "${APP_NAME}.app" 200 190 \
  --hide-extension "${APP_NAME}.app" \
  --app-drop-link 600 185 \
  "${DMG_FILE_NAME}" \
  "${SOURCE_FOLDER_PATH}"

  rm -r app
  rm *.zip