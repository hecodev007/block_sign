# helium go sdk
### ***Notice***
    Only ed25519 curve is currently spported
    
### Payment_v1 
    #  init a transaction v1
    v1 := transactions.NewPaymentV1Tx(payer, payee, 10, 0, 1, make([]byte,64))
    
    #  set fee 
    c := http.NewHeliumRpc("https://api.helium.io")
    vars,_ := c.GetVars()
    payload, err := v1.Serialize()
    fee := transactions.CalculateFee(int64(len(payload)), vars.DcPayloadSize, vars.TxnFeeMultiplier)
    v1.SetFee(fee)
    
    #  get unsign transaction data
    v1Tx, err := v1.BuildTransaction(true)
    if err != nil {
    	panic(err)
    }
   
    # sign transaction
    kp := keypair.NewKeypairFromHex(1, "secret key")   # type=1 is mean use ed25519
    sig, err := kp.Sign(v1Tx)
    if err != nil {
    	panic(err)
    }
    
    #  set sign data
    v1.SetSignature(sig)
    
    # get finally data
    ser, err := v1.Serialize()
    if err != nil {
    	panic(err)
    }
    txn :=base64.StdEncoding.EncodeToString(ser)
    
    # sumbbit transaction
    txid,err := c.BroadcastTransaction(txn)
    
 ### Payment_v2
    it is like Payment_v1