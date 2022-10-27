package stg

import "rsksync/utils/stg"

type ScanServer struct {
	client stg.StgWSClient
}

func (s *ScanServer) ScanHistory(interval string) {

}
