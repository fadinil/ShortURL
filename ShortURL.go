package main

/*
При тестировании сервиса отправлял запросы через Postman на localhost
*/
import (
	"database/sql" //предоставляет интерфейс для SQL
	"flag"         //для обработки флагов
	"fmt"          // пакет для форматированного ввода вывода
	"io/ioutil"
	"math/rand"
	"net/http" // пакет для поддержки HTTP протокола
	"time"

	_ "github.com/lib/pq" //драйвер для database/sql
)

const (
	// Инициализация констант
	HOST        = "localhost"
	PORT        = "5050"
	DATABASE    = "ShortURL"
	USER        = "fadinil"
	PASSWORD    = "123"
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Adress struct { //Структура для адреса
	Link     string
	ShortURL string
}

var useDB = flag.Bool("d", false, "add ShortURL to DB") //инициализация флага -d

func shorting() string { //ф-ия создает сокращение для ссылки
	b := make([]byte, 10)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func checkError(err error) { //проверка на ошибки
	if err != nil {
		panic(err)
	}
}

func ConnectToDB() *sql.DB { //подключение к postgresql
	var connectionString string = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", HOST, PORT, USER, PASSWORD, DATABASE)
	db, err := sql.Open("postgres", connectionString)
	checkError(err)
	return db
}

func Controller(w http.ResponseWriter, r *http.Request) { //обработчик запросов
	flag.Parse()
	adress := Adress{}     //создание объекта адреса
	if r.Method == "GET" { //обработка GET запроса
		adress.ShortURL = r.URL.Path
		db := ConnectToDB()
		er := db.QueryRow("SELECT link FROM links WHERE short = $1", adress.ShortURL[1:]).Scan(&adress.Link)
		checkError(er)
		fmt.Fprintf(w, adress.Link)
	}
	if r.Method == "POST" { //обработка POST запроса
		body, _ := ioutil.ReadAll(r.Body)
		adress.ShortURL = shorting()
		adress.Link = string(body)
		if *useDB { //если стоит флаг -d подключаемся к бд и сохраняем туда ссылку
			db := ConnectToDB()
			db.Exec("insert into links (link, short) values ($1, $2)", adress.Link, adress.ShortURL)
			fmt.Println("Ссылка сокращена и сохранена в базу данных успешно!")
		}
		fmt.Fprintf(w, "http://localhost:5000/"+adress.ShortURL)
	}

}

func main() {
	http.HandleFunc("/", Controller)
	// задаем слушать порт
	fmt.Println("Server is listening...")
	serv := http.ListenAndServe(":5000", nil)
	checkError(serv)
}
