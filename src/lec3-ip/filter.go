package main

type Filter interface {
	Run(src interface{}) interface{}
}
