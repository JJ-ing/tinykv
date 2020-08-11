package standalone_storage

import (
	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
)



type StandaloneReader struct{
	txn    *badger.Txn
}


func NewStandaloneReader(txn *badger.Txn) *StandaloneReader {
	return &StandaloneReader{
		txn:    txn,
	}
}


func (r *StandaloneReader) 	GetCF (cf string, key []byte) ([]byte, error){
	//if err := util.CheckKeyInRegion(key, r.region); err != nil {
	//	return nil, err
	//}
	val, err := engine_util.GetCFFromTxn(r.txn, cf, key)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return val, err
}


func (r *StandaloneReader) IterCF(cf string) engine_util.DBIterator{
	return NewStandaloneIterator(engine_util.NewCFIterator(cf, r.txn))
}


func (r *StandaloneReader) Close(){
	r.txn.Discard()
}

type StandaloneIterator struct {
	iter   *engine_util.BadgerIterator
	//region *metapb.Region
}

func NewStandaloneIterator(iter *engine_util.BadgerIterator) *StandaloneIterator {
	return &StandaloneIterator{
		iter:   iter,
		//region: region,
	}
}

func (it *StandaloneIterator) Item() engine_util.DBItem {
	return it.iter.Item()
}

func (it *StandaloneIterator) Valid() bool {
	if !it.iter.Valid() {
		return false
	}
	return true
}

func (it *StandaloneIterator) ValidForPrefix(prefix []byte) bool {
	if !it.iter.ValidForPrefix(prefix) {
		return false
	}
	return true
}

func (it *StandaloneIterator) Close() {
	it.iter.Close()
}

func (it *StandaloneIterator) Next() {
	it.iter.Next()
}

func (it *StandaloneIterator) Seek(key []byte) {
	//if err := util.CheckKeyInRegion(key, it.region); err != nil {
	//	panic(err)
	//}
	it.iter.Seek(key)
}

func (it *StandaloneIterator) Rewind() {
	it.iter.Rewind()
}