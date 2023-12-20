package tools

import (
	"fmt"
	"strconv"
	"strings"
)

//字符串切片比较
func ArrStrCmp(src []string, dest []string) ([]string, []string) {
	msrc := make(map[string]byte) //按源数组建索引
	mall := make(map[string]byte) //源+目所有元素建索引
	var set []string              //交集
	//1.源数组建立map
	for _, v := range src {
		msrc[v] = 0
		mall[v] = 0
	}
	//2.目数组中，存不进去，即重复元素，所有存不进去的集合就是并集
	for _, v := range dest {
		l := len(mall)
		mall[v] = 1
		if l != len(mall) {
			continue
		} else {
			set = append(set, v)
		}
	}
	//3.遍历交集，在并集中找，找到就从并集中删，删完后就是补集（即并-交=所有变化的元素）
	for _, v := range set {
		delete(mall, v)
	}
	//4.此时，mall是补集，所有元素去源中找，找到就是删除的，找不到的必定能在目数组中找到，即新加的
	var added, deleted []string
	for v := range mall {
		_, exist := msrc[v]
		if exist {
			deleted = append(deleted, v)
		} else {
			added = append(added, v)
		}
	}
	return added, deleted
}

//uint切片比较
func ArrUintCmp(src []uint, dest []uint) ([]uint, []uint) {
	msrc := make(map[uint]byte) //按源数组建索引
	mall := make(map[uint]byte) //源+目所有元素建索引
	var set []uint              //交集
	//1.源数组建立map
	for _, v := range src {
		msrc[v] = 0
		mall[v] = 0
	}
	//2.目数组中，存不进去，即重复元素，所有存不进去的集合就是并集
	for _, v := range dest {
		l := len(mall)
		mall[v] = 1
		if l != len(mall) {
			continue
		} else {
			set = append(set, v)
		}
	}
	//3.遍历交集，在并集中找，找到就从并集中删，删完后就是补集（即并-交=所有变化的元素）
	for _, v := range set {
		delete(mall, v)
	}
	//4.此时，mall是补集，所有元素去源中找，找到就是删除
	var added, deleted []uint
	for v := range mall {
		_, exist := msrc[v]
		if exist {
			deleted = append(deleted, v)
		} else {
			added = append(added, v)
		}
	}
	return added, deleted
}

//将字符串切片转换为uint切片
func SliceToString(src []uint, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(src), " ", delim, -1), "[]")
}

//将字符串切片转换为uint切片
func StringToSlice(src string, delim string) []uint {
	var dest []uint
	if src == "" {
		return dest
	}
	strs := strings.Split(src, delim)
	for _, v := range strs {
		t, _ := strconv.Atoi(v)
		dest = append(dest, uint(t))
	}
	return dest
}
