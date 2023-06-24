package app

import (
	"os"
	"path"
	"sort"
	"strings"

	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	bolt "go.etcd.io/bbolt"
)

const (
	TransLateSettings = "settings"
	ZhCN              = "zh_CN"
	EnUS              = "en_US"
)

func (a *Application) TransDB() (*bolt.DB, error) {
	if a.transDB == nil {
		var err error
		a.transDB, err = bolt.Open(path.Join(a.appConfig.System.Workdir, "data/trans.db"), 0666, nil)
		if err != nil {
			log.Errorf("open trans db erro %s", err)
			return nil, err
		}

		_ = a.transDB.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(TransLateSettings)) // ③
			if err != nil {
				log.Errorf("create settings bucket: %s", err)
			}
			_, err = tx.CreateBucketIfNotExists([]byte(ZhCN)) // ③
			if err != nil {
				log.Errorf("create zh_CN bucket: %s", err)
			}
			_, err = tx.CreateBucketIfNotExists([]byte(EnUS)) // ③
			if err != nil {
				log.Errorf("create en_US bucket: %s", err)
			}
			return nil
		})

	}
	return a.transDB, nil
}

func (a *Application) GetTranslateLang() string {
	var lang = EnUS
	transdb, err := a.TransDB()
	if err != nil {
		return EnUS
	}
	_ = transdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TransLateSettings))
		if b != nil {
			_lang := string(b.Get([]byte("CurrentLang")))
			if _lang != "" {
				lang = _lang
			}
		}
		return nil
	})
	return lang
}

func (a *Application) SetTranslateLang(lang string) {
	if !common.InSlice(lang, []string{ZhCN, EnUS}) {
		return
	}
	transdb, err := a.TransDB()
	if err != nil {
		return
	}
	_ = transdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TransLateSettings))
		if b != nil {
			b.Put([]byte("CurrentLang"), []byte(lang))
		}
		return nil
	})
}

func (a *Application) LoadTranslateDict(lang string) map[string]map[string]string {
	transdb, err := a.TransDB()
	if err != nil {
		return nil
	}
	var result = make(map[string]map[string]string)
	_ = transdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(lang))
		if b == nil {
			return nil
		}

		_ = b.ForEach(func(k, v []byte) error {
			if v != nil {
				return nil
			}
			result[string(k)] = make(map[string]string)
			sub := b.Bucket(k)
			_ = sub.ForEach(func(kk, vv []byte) error {
				result[string(k)][string(kk)] = string(vv)
				return nil
			})
			return nil
		})

		return nil
	})

	return result
}

type TransTable struct {
	Lang   string `json:"lang"`
	Module string `json:"module"`
	Source string `json:"source"`
	Result string `json:"result"`
}

func (a *Application) QueryTranslateTable(lang string, module, keyword string) []TransTable {
	transdb, err := a.TransDB()
	if err != nil {
		return nil
	}
	var result = make([]TransTable, 0)
	_ = transdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(lang))
		if b == nil {
			return nil
		}

		_ = b.ForEach(func(k, v []byte) error {
			if v != nil {
				return nil
			}
			sub := b.Bucket(k)
			_ = sub.ForEach(func(kk, vv []byte) error {
				if module != "" && module != string(k) {
					return nil
				}
				if keyword != "" && !strings.Contains(string(kk), keyword) {
					return nil
				}
				result = append(result, TransTable{
					Lang:   lang,
					Module: string(k),
					Source: string(kk),
					Result: string(vv),
				})
				return nil
			})
			return nil
		})

		return nil
	})
	sort.Slice(result, func(i, j int) bool {
		return result[i].Module < result[j].Module
	})

	return result
}

func (a *Application) Translate(lang, module, src, defValue string) string {
	transdb, err := a.TransDB()
	if err != nil {
		return src
	}
	tx, err := transdb.Begin(true)
	if err != nil {
		return src
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte(lang))
	if b == nil {
		return src
	}

	isUpdate := false

	sub := b.Bucket([]byte(module))
	if sub == nil {
		sub, err = b.CreateBucketIfNotExists([]byte(module))
		if err != nil {
			return src
		}
		isUpdate = true
	}

	ret := sub.Get([]byte(src))
	if ret == nil {
		ret = []byte(defValue)
		_ = sub.Put([]byte(src), ret)
		isUpdate = true
	}

	if isUpdate {
		_ = tx.Commit()
	}

	return string(ret)
}

func (a *Application) TranslateUpdate(lang, module, src, value string) string {
	transdb, err := a.TransDB()
	if err != nil {
		return src
	}
	tx, err := transdb.Begin(true)
	if err != nil {
		return src
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte(lang))
	if err != nil {
		return src
	}

	sub, err := b.CreateBucketIfNotExists([]byte(module))
	if err != nil {
		return src
	}

	_ = sub.Put([]byte(src), []byte(value))

	_ = tx.Commit()

	return value
}

func (a *Application) RenderTranslateFiles() {
	langs := []string{ZhCN, EnUS}
	for _, lang := range langs {
		r := a.LoadTranslateDict(lang)
		os.WriteFile(path.Join(a.appConfig.System.Workdir, "data", "trans_"+lang+".js"), []byte("window.GlobalTrans="+common.ToJson(r)), 0666)
	}
}

func (a *Application) RemoveTranslateItems(items []TransTable) {
	transdb, err := a.TransDB()
	if err != nil {
		return
	}
	tx, err := transdb.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()
	langs := []string{ZhCN, EnUS}
	for _, lang := range langs {
		b := tx.Bucket([]byte(lang))
		if b == nil {
			continue
		}
		for _, item := range items {
			sub := b.Bucket([]byte(item.Module))
			if sub == nil {
				continue
			}
			_ = sub.Delete([]byte(item.Source))
		}
	}
	_ = tx.Commit()

}

func Trans(module, src string) string {
	return app.Translate(app.GetTranslateLang(), module, src, src)
}
