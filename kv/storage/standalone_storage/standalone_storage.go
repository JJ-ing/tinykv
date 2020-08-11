package standalone_storage

import (
	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
	"os"
	"path/filepath"
)

// StandAloneStorage is an implementation of `Storage` for a single-node TinyKV instance. It does not
// communicate with other nodes and all data is stored locally.
type StandAloneStorage struct {
	// Your Data Here (1).
	engines *engine_util.Engines
	config  *config.Config
}

func NewStandAloneStorage(conf *config.Config) *StandAloneStorage {
	// Your Code Here (1).
	dbPath := conf.DBPath
	kvPath := filepath.Join(dbPath, "kv")
	snapPath := filepath.Join(dbPath, "snap")

	os.MkdirAll(kvPath, os.ModePerm)
	os.MkdirAll(snapPath, os.ModePerm)

	kvDB := engine_util.CreateDB(kvPath, false)
	engines := engine_util.NewEngines(kvDB, nil, kvPath, "")
	return &StandAloneStorage{engines: engines, config: conf}

}

func (s *StandAloneStorage) Start() error {
	// Your Code Here (1).

	return nil
}

func (s *StandAloneStorage) Stop() error {
	// Your Code Here (1).
	if err := s.engines.Kv.Close(); err !=nil {
		return err
	}
	return nil
}

func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	// Your Code Here (1).
	//var standaloneReader *StandaloneReader
	//err := s.engines.Kv.View(func(txn *badger.Txn) error {
	//	standaloneReader = NewStandaloneReader(txn)
	//	return nil
	//})
	txn := s.engines.Kv.NewTransaction(false)
	standaloneReader := NewStandaloneReader(txn)
	return standaloneReader, nil
}

func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	// Your Code Here (1).
	for _, m := range batch{
		switch m.Data.(type) {
		case storage.Put:
			put := m.Data.(storage.Put)
			wb := engine_util.WriteBatch{}
			wb.SetCF(put.Cf, put.Key, put.Value)
			if err := s.engines.WriteKV(&wb); err != nil{
				return err
			}
		case storage.Delete:
			delete := m.Data.(storage.Delete)
			wb := engine_util.WriteBatch{}
			wb.SetCF(delete.Cf, delete.Key, nil)		// #? tCF(), len(key)!=len(val)
			if err := s.engines.WriteKV(&wb); err != nil{
				return err
			}
		}
	}
	return nil
}
