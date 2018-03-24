export LOGXI=*
export LOGXI_FORMAT=pretty,happy

echo "Starting peer 1 (6001 -> 6002)"
echo "'m' - message, 'q' - quit"
go run peer.go message.go -name bobby -self localhost:6001 -next localhost:6002
