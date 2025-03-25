echo "Pulling git..." &&
git pull &&
echo "Entering server/" &&
cd server && 
echo "Building (downloading deps if needed)..." && 
go build *.go &&
echo "Done." && 
echo "Moving executable and leaving source..."
cd .. && 
mv server/main . && 
echo "Done"
