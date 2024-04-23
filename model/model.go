package model

const DatePattern string = "20060102"

type NextDate struct {
	Now    string `form:"now"`
	Date   string `form:"date"`
	Repeat string `form:"repeat"`
}
type ListTasks struct {
	Tasks []Task `json:"tasks"`
}

type Task struct {
	ID      string `json:"-"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
