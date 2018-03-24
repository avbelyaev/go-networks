export LOGXI=*
export LOGXI_FORMAT=pretty,happy

echo "Starting peer 3 (6003 -> 6001)"
echo "'m' - message, 'q' - quit"
go run peer.go message.go -name eddy -self localhost:6003 -next localhost:6001
