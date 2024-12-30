GOOS=darwin GOARCH=arm64 go build -o bin/mezmer

mkdir app
mkdir -p app/Mezmer.app/Contents/MacOS
mkdir -p app/Mezmer.app/Contents/Resources

codesign -vvv --deep --force --options=runtime --entitlements ./resources/entitlements.plist --sign "Developer ID Application: Ignat Drozdov (27B2YLEUVR)" --timestamp ./bin/mezmer

mv mezmer app/Mezmer.app/Contents/MacOS/mezmer
cp resources/LICENSE.txt app/Mezmer.app/Contents/Resources
cp resources/Mezmer.icns app/Mezmer.app/Contents/Resources/AppIcon.icns
cp resources/Info.plist app/Mezmer.app/Contents
cp resources/entitlements.plist app/Mezmer.app/Contents
rm *.dmg

APP_PATH="app/Mezmer.app"
ZIP_PATH="Mezmer.zip"

# Create a ZIP archive suitable for notarization.
/usr/bin/ditto -c -k --sequesterRsrc --keepParent "$APP_PATH" "$ZIP_PATH"

xcrun notarytool submit Mezmer.zip --keychain-profile "notarytool-password" --wait


APP_NAME="Mezmer"
DMG_FILE_NAME="${APP_NAME}.dmg"
VOLUME_NAME="${APP_NAME}"
SOURCE_FOLDER_PATH="app/"


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