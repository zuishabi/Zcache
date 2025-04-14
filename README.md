# Zcache

## 简介

* Zcache是一个以groupCache的核心代码为基础，在此基础上以自己的拙见所魔改和拓展的缓存应用。
* 在启动服务后会默认创建一个名为default的组，大小为2048字节

## 基础功能

* 支持键值对的读写
* 缓存以group为核心，group可以创建，并配置其存储空间大小，存储的键值对相互隔离
* 可进行缓存的配置
* 可进行数据持久化

## 客户端命令使用

* set -groupName(默认:default) -key -value
  * 设置一个组中的键值对
* get -groupName(默认:default) -key
  * 获得一个组中的键对应的值
* getKeys -groupName(默认:default)
  * 获得一个组的键列表
* getGroups
  * 获得全局组列表
* exit
  * 退出客户端

## 更新日志

* 完成项目的配置
* 完成持久化文件的配置与读取
