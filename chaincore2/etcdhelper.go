package main

//
//type EtcdClient struct {
//	client     *clientv3.Client
//	serverList map[string]string
//	lock       sync.Mutex
//}

var userKey string = "4V8sGLHlncT@tuTMRxsKvj6QdyVfnCyg"
var confKey string = "nhyL53JJgdqUNx0U*6%xW%K1n#k8FEAG"

//
//func NewEtcdClient(addr []string, user, pass string) (*EtcdClient, error) {
//	conf := clientv3.Config{
//		Endpoints:   addr,
//		DialTimeout: 5 * time.Second,
//	}
//	if user != "" && pass != "" {
//		conf.Username, _ = common.AesEncrypt(user, []byte(userKey))
//		conf.Password, _ = common.AesEncrypt(pass, []byte(userKey))
//	}
//	client, err := clientv3.New(conf)
//	if err == nil {
//		return &EtcdClient{
//			client:     client,
//			serverList: make(map[string]string),
//		}, nil
//	}
//	return nil, err
//}
//
//func (this *EtcdClient) GetConfig(key string) (string, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	resp, err := this.client.Get(ctx, key)
//	cancel()
//	if err != nil {
//		log.Error(err)
//		return "", err
//	}
//
//	if resp == nil || resp.Kvs == nil {
//		return "", nil
//	}
//	for i := range resp.Kvs {
//		if v := resp.Kvs[i].Value; v != nil {
//			encryptStr, err := common.AesDecrypt(string(v), []byte(confKey))
//			if err != nil {
//				log.Error(err)
//				return "", err
//			}
//			return encryptStr, nil
//		}
//	}
//	return "", nil
//}
//
//func (this *EtcdClient) SetConfig(key string, value string) error {
//	encryptStr, err := common.AesEncrypt(value, []byte(confKey))
//	if err != nil {
//		log.Error(err)
//		return err
//	}
//	_, err = this.client.Put(context.TODO(), key, encryptStr)
//	if err != nil {
//		log.Error(err)
//		return err
//	}
//	return nil
//}
//
//func (this *EtcdClient) GetService(prefix string) ([]string, error) {
//	resp, err := this.client.Get(context.Background(), prefix, clientv3.WithPrefix())
//	if err != nil {
//		return nil, err
//	}
//	addrs := this.extractAddrs(resp)
//
//	go this.watcher(prefix)
//	return addrs, nil
//}
//
//func (this *EtcdClient) watcher(prefix string) {
//	// nothing
//	//rch := this.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
//	//for wresp := range rch {
//	//	for _, ev := range wresp.Events {
//	//		switch ev.Type {
//	//		case mvccpb.PUT:
//	//			this.SetServiceList(string(ev.Kv.Key),string(ev.Kv.Value))
//	//		case mvccpb.DELETE:
//	//			this.DelServiceList(string(ev.Kv.Key))
//	//		}
//	//	}
//	//}
//}
//
//func (this *EtcdClient) extractAddrs(resp *clientv3.GetResponse) []string {
//	addrs := make([]string, 0)
//	if resp == nil || resp.Kvs == nil {
//		return addrs
//	}
//	for i := range resp.Kvs {
//		if v := resp.Kvs[i].Value; v != nil {
//			this.SetServiceList(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value))
//			addrs = append(addrs, string(v))
//		}
//	}
//	return addrs
//}
//
//func (this *EtcdClient) SetServiceList(key, val string) {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	this.serverList[key] = string(val)
//	log.Debug("set data key :", key, "val:", val)
//}
//
//func (this *EtcdClient) DelServiceList(key string) {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	delete(this.serverList, key)
//	log.Debug("del data key:", key)
//}
//
//func (this *EtcdClient) SerList2Array() []string {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	addrs := make([]string, 0)
//
//	for _, v := range this.serverList {
//		addrs = append(addrs, v)
//	}
//	return addrs
//}
