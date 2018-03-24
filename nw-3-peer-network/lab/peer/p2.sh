export LOGXI=*
export LOGXI_FORMAT=pretty,happy

echo "Starting peer 2 (6002 -> 6003)"
echo "'m' - message, 'q' - quit"
go run peer.go message.go -name donald -self localhost:6002 -next localhost:6003
