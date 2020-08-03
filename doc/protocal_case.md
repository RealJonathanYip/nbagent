### NODE_SAY_HELLO
```
	NodeManager) checkNeighbour		Info) reconnect		SayHelloReq
```
### NODE_SAY_GOOBEY
```
	NodeManager) Stop	SayGoodbyeNotify
```
### NODE_GET_ENTRY
```
	NodeManager) trySyncAll		GetEntryReq
```
### NODE_ENTRY_NOTIFY
```
	NodeManager) RemoveEntry	NewEntryNotify
	NodeManager) AddEntry		NewEntryNotify
```
### NODE_KEEP_ALIVE
```
	NodeManager) keepalive			KeepAliveNotify
	NodeManager) handleSayHello		KeepAliveNotify
```