echo "Pulling git..." &&
git pull &&
echo "Entering server/" &&
cd server && 
echo "Building (downloading deps if needed)..." && 
go build *.go &&
echo "Done (build $(pwd)/main.go)" 
