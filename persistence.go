package cache

import (
	"encoding/json"
	"fmt"
	"os"
)

type PersistenceType struct {
	Key   string
	Value []byte
}

type GroupInfo struct {
	Name string
	Num  int
}

// ByteToPersistence 将ByteView类型转换为Persistence类型
func ByteToPersistence(key string, view *ByteView) PersistenceType {
	return PersistenceType{Key: key, Value: cloneBytes(view.b)}
}

// SavePersistence 将数据进行保存
func SavePersistence() {
	fmt.Println("saving the persistence file...")
	f, err := os.OpenFile("persistence.zsave", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, v := range GetGroupList() {
		g := GetGroup(v)
		if err := g.SaveGroup(f); err != nil {
			panic(err)
		}
	}
	fmt.Println("saving complete")
}

// LoadPersistence 加载持久化文件
func LoadPersistence() {
	fmt.Println("loading persistence file")
	f, err := os.OpenFile("persistence.zsave", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		d := json.NewDecoder(f)
		info := GroupInfo{}
		if err := d.Decode(&info); err != nil {
			fmt.Println(err)
			break
		}
		keys := make([]string, info.Num)
		values := make([]ByteView, info.Num)
		for i := range info.Num {
			e := PersistenceType{} // 每次创建新的实例
			if err := d.Decode(&e); err != nil {
				fmt.Println(err)
				break
			}
			keys[i] = e.Key
			values[i] = ByteView{b: e.Value}
		}
		g := NewGroup(info.Name, 2048, nil)
		g.SetList(keys, values)
	}

	fmt.Println("load persistence file complete")
}
