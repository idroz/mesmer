GOOS=darwin GOARCH=arm64 go build -o mezmer

mkdir bin
mkdir -p bin/Mezmer.app/Contents/MacOS
mkdir -p bin/Mezmer.app/Contents/Resources

mv mezmer bin/Mezmer.app/Contents/MacOS/Mezmer
cp resources/LICENSE.txt bin/Mezmer.app/Contents/Resources
cp resources/Mezmer.icns bin/Mezmer.app/Contents/Resources/AppIcon.icns
cp resources/Info.plist bin/Mezmer.app/Contents
rm *.dmg

APP_NAME="Mezmer"
DMG_FILE_NAME="${APP_NAME}.dmg"
VOLUME_NAME="${APP_NAME}"
SOURCE_FOLDER_PATH="bin/"

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

  rm -r bin