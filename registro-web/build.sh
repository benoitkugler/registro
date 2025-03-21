# Build script for web app

echo "Building app" &&
bun run build &&
echo "Copying files" &&
rm -rf ../server/static/* &&
cp -r dist/assets ../server/static &&
cp -r dist/acve ../server/static &&
cp -r dist/repere ../server/static &&
cd dist/src/clients/ && 
cp -r * ../../../../server/static &&
echo "Done."