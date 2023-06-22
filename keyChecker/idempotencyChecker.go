package keyChecker

import (
	"encoding/json"
	"errors"
	"github.com/Ignaciocl/tp1SisdisCommons/fileManager"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	"io"
	"strings"
)

const Sep = "hereEnds the key"

type key struct {
	IdempotencyKey string `json:"idempotency_key"`
	Id             int64  `json:"id"`
}

func (k key) GetId() int64 {
	return k.Id
}

func (k key) SetId(id int64) {
	k.Id = id
}

type trans struct{}

func (t trans) ToWritable(data key) []byte {
	k, _ := json.Marshal(data)
	return k
}

func (t trans) FromWritable(d []byte) key {
	data := strings.Split(string(d), Sep)[0]
	var r key
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		utils.LogError(err, "could not unmarshal from db")
	}
	return r
}

type idempotencyChecker struct {
	keys              map[string]key
	chronologicalKeys []string
	db                fileManager.Manager[key]
	limitKeys         int
	currentPosKey     int
}

func (i *idempotencyChecker) IsKey(key string) bool {
	_, ok := i.keys[key]
	return ok
}

func (i *idempotencyChecker) AddKey(keyToAdd string) error {
	if i.IsKey(keyToAdd) {
		return nil
	}
	if i.limitKeys <= len(i.chronologicalKeys) {
		d := i.keys[i.chronologicalKeys[i.currentPosKey]]
		delete(i.keys, d.IdempotencyKey)
		newKey := key{keyToAdd, d.Id}
		if err := i.db.Write(newKey); err != nil {
			return err
		}
		i.chronologicalKeys[i.currentPosKey] = newKey.IdempotencyKey
	} else {
		k := key{
			IdempotencyKey: keyToAdd,
			Id:             0,
		}
		if err := i.db.Write(k); err != nil {
			return err
		}
		i.keys[k.IdempotencyKey] = k
		i.chronologicalKeys = append(i.chronologicalKeys, k.IdempotencyKey)
	}
	i.currentPosKey = (i.currentPosKey + 1) % i.limitKeys
	return nil
}

func (i *idempotencyChecker) fillData() error {
	if data, err := i.db.Read(); err != nil && !errors.Is(err, io.EOF) {
		return err
	} else {
		for _, d := range data {
			i.keys[d.IdempotencyKey] = d
			i.chronologicalKeys = append(i.chronologicalKeys, d.IdempotencyKey)
			i.currentPosKey = (i.currentPosKey + 1) % i.limitKeys
		}
	}
	return nil
}

func CreateIdempotencyChecker(maxAmountKeys int) (Checker, error) {
	db, err := fileManager.CreateDB[key](trans{}, "idempotencyDB", 200, Sep)
	if err != nil {
		return nil, err
	}
	ik := idempotencyChecker{
		keys:              make(map[string]key, 0),
		chronologicalKeys: make([]string, 0, maxAmountKeys),
		db:                db,
		limitKeys:         maxAmountKeys,
		currentPosKey:     0,
	}
	return &ik, ik.fillData()
}
