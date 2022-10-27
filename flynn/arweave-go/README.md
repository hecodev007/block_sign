# arweave-go sdk

## add transaction v2
### create v2 transfer transaction
    target:=""
    lastTx,_:=api.GetTransactionAnchor(context.TODO)
    reward,_:=api.GetRewardV2(context.TODO,nil,target)
    txV2:=NewTransactionV2(lastTx,wallet.PubKeyModulus(),"1000000000000",target,nil,reward)
    tx,errSig:=txV2.Sign(w)
    if errSig != nil {
        return nil,errSig
    }
    data,_:=json.Marshal(tx)
    resp,err :=api.Commit(context.TODO(),data)
    //resp="OK"