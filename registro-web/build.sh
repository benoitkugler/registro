# Build script for web app

# We build one version of the app for every asso
for asso in acve repere
do
    echo "Building app for $asso..." &&
    bun run vite build --mode $asso &&
    echo "Copying files..." &&
    rm -rf ../server/static/$asso/* &&
    cp -r dist/assets ../server/static/$asso &&
    # only copy the asso specific folder
    cp -r dist/$asso ../server/static/$asso &&
    cd dist/src/clients/ && 
    cp -r * ../../../../server/static/$asso &&
    cd ../../.. &&
    echo "Done.";
done

