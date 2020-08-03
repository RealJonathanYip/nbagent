### NewManager()
```go
NewManager()
    ptrServer := network.NewServer()
        objListener := net.Listen()
    ptrServer.Start()
        objConnection := objListener.Accept()
        objServerConnection := ServerConnection{objConnection, ...}
        objServerConnection.readLoop()
```

### ServerConnection.readLoop()
```go
    atomic.StoreUint32(&this.nLastReceiveTime, uint32(time.Now().Unix()))

    ...

    ctx := context.WithValue(context.TODO(), REQUEST_CONTEXT,
        newContext(this, ""))
    this.objHandler.OnRead(this, ptrBuffer.Bytes(), ctx)
    // objHandler 为 NodeManager，在 NewManager 创建，
    // 通过 network.NewServer 传入（PS：目前handler、connectionHandler都是 NodeManager）
```

### NodeManager.OnRead()
```go
    ptrNodeRouter.Process() // ptrNodeRouter 在 NewManager 中创建：ptrRouter := router.NewRouter()
    ptrNode := getNode(ctx)
    ptrNode.refreshKeepLiveTime()
```

### Router.Process()
```go
    szUri, bSuccess := protocol.ReadString(&objReader)
    anyErr := binary.Read(&objReader, binary.LittleEndian, &nMsgType)
    szRequestID, bSuccess := protocol.ReadString(&objReader)
    network.SetRequestID(ctx, szRequestID)
        ptrMsgContext := GetRequestContext(ctx);
            ptrContext := ctx.Value(REQUEST_CONTEXT)
        ptrMsgContext.szRequestID = szRequestID
    
    switch nMsgType:
		case protocol.MSG_TYPE_CAST:
			fnRouter2Handler()
			break
		case protocol.MSG_TYPE_REQUEST:
			fnRouter2Handler()
			break
		case protocol.MSG_TYPE_RESPONE:
			ptrServerConnection := network.GetServerConnection(ctx)
			ptrRpcContext, bSuccess := ptrServerConnection.GetRpcContext(szRequestID) // 没有响应的情况，会造成泄露

			if bSuccess {
				ptrRpcContext.Msg <- objReader.Bytes()
			} else {
				log.Warningf("can not find response context szRequestID:%v", szRequestID)
			}
			break
```
