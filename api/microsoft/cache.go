package microsoft

import (
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"io/ioutil"
	"log"
	"os"
)

type TokenCache struct {
	file string
}

func (t *TokenCache) Replace(cache cache.Unmarshaler, key string) {
	jsonFile, err := os.Open(t.file)
	if err != nil {
		log.Println(err)
	}
	defer jsonFile.Close()
	data, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Println(err)
	}
	err = cache.Unmarshal(data)
	if err != nil {
		log.Println(err)
	}
}

func (t *TokenCache) Export(cache cache.Marshaler, key string) {
	data, err := cache.Marshal()
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(t.file, data, 0600)
	if err != nil {
		log.Println(err)
	}
}
