#!/usr/bin/env bash

# Build script for web app

# We build one version of the app for every asso
for asso in acve repere; do
    echo "Building app for $asso..."
    bun run vite build --mode "$asso"

    static_dir="../server/static/$asso"
    echo "Copying files to $static_dir..."
    rm -rf "$static_dir"
    mkdir -p "$static_dir"

    # Shared frontend bundles are served from /static/assets.
    cp -R dist/assets "$static_dir/"
    # Keep assets such as the association-specific favicon and logos.
    cp -R "dist/$asso" "$static_dir/"
    # The server serves each app's HTML from these flattened client directories.
    cp -R dist/src/clients/. "$static_dir/"
    # Workbox precache URLs keep this path; the flattened copies above serve app routes.
    mkdir -p "$static_dir/src"
    cp -R dist/src/clients "$static_dir/src/"
    # The manifest and worker are served from /static; the icon is relative to the manifest.
    cp dist/manifest.webmanifest dist/sw.js "dist/$asso/logo-small.png" "$static_dir/"

    echo "Done building $asso."
done

