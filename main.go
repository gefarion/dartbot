package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/jsonq"
	"github.com/tucnak/telebot"
)

// UTILS FUNCS

func replyMessage(bot *telebot.Bot, message telebot.Message, response string) {
	if message.Chat.IsGroupChat() {
		bot.SendMessage(message.Chat, response, &telebot.SendOptions{ReplyTo: message})
	} else {
		bot.SendMessage(message.Chat, response, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
	}
}

func parseCommmand(bot *telebot.Bot, message telebot.Message) (command string, params []string) {
	r := regexp.MustCompile("'.+'|\".+\"|\\S+")
	fields := r.FindAllString(message.Text, -1)
	if len(fields) == 0 {
		return "", []string{}
	}

	if message.Chat.IsGroupChat() {
		if len(fields) > 1 && fields[0] == "@"+bot.Identity.Username {
			if len(fields) > 2 {
				return fields[1], fields[2:]
			} else {
				return fields[1], []string{}
			}
		}
	} else {
		if len(fields) > 1 {
			return fields[0], fields[1:]
		} else {
			return fields[0], []string{}
		}
	}

	return "", []string{}
}

func doJSONRequest(url string) (*jsonq.JsonQuery, error) {
	var jq *jsonq.JsonQuery

	resp, err := http.Get(url)
	if err != nil {
		return jq, errors.New(fmt.Sprint("Error on JSON request: ", err))
	}
	defer resp.Body.Close()

	data := map[string]interface{}{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&data)
	jq = jsonq.NewQuery(data)

	return jq, nil
}

// COMMAND HANDLERS

var handlers map[string]messageHandler

type messageHandler func(*telebot.Bot, telebot.Message, []string) error

func handlerHelp(bot *telebot.Bot, message telebot.Message, args []string) error {
	response := "Available commands:\n"
	for key, _ := range handlers {
		response = response + key + "\n"
	}
	replyMessage(bot, message, response)
	return nil
}

func handlerPing(bot *telebot.Bot, message telebot.Message, args []string) error {
	replyMessage(bot, message, "Pong!")
	return nil
}

func handlerDolar(bot *telebot.Bot, message telebot.Message, args []string) error {

	jq, err := doJSONRequest("http://api.bluelytics.com.ar/v2/latest")
	if err != nil {
		return err
	}

	sell_price, _ := jq.Float("oficial", "value_sell")
	buy_price, _ := jq.Float("oficial", "value_buy")

	replyMessage(bot, message, fmt.Sprintf("Sell: %v\nBuy: %v", sell_price, buy_price))
	return nil
}

func handlerWeather(bot *telebot.Bot, message telebot.Message, args []string) error {
	jq, err := doJSONRequest("http://api.openweathermap.org/data/2.5/weather?q=CiudadBuenosAires,ar&appid=a471bd5630cff0c5447128d3e3fd8ca3&units=metric&lang=sp")
	if err != nil {
		return err
	}

	weather, _ := jq.String("weather", "0", "main")
	temp, _ := jq.Float("main", "temp")
	temp_min, _ := jq.Float("main", "temp_min")
	temp_max, _ := jq.Float("main", "temp_max")
	humidity, _ := jq.Float("main", "humidity")

	replyMessage(bot, message, fmt.Sprintf("Weather: %v\nTemperature: %v\nMinimun: %v\nMaximun: %v\nHumidity: %v",
		weather, temp, temp_min, temp_max, humidity))

	return nil
}

func handlerMetro(bot *telebot.Bot, message telebot.Message, args []string) error {
	doc, err := goquery.NewDocument("http://www.metrovias.com.ar/")
	if err != nil {
		return err
	}

	status := ""
	lines := []string{"A", "B", "C", "D", "E", "H"}

	for _, line := range lines {
		span := doc.Find("span#status-line-" + line)
		status += fmt.Sprintf("Line %v: %v\n", line, span.Text())
	}

	replyMessage(bot, message, status)
	return nil
}

func handlerFutbol(bot *telebot.Bot, message telebot.Message, args []string) error {

	replyMessage(bot, message, "TODO")
	return nil
}

// RUN THE BOT!

func main() {
	startTime := time.Now()

	handlers = map[string]messageHandler{
		"ping":   handlerPing,
		"dolar":  handlerDolar,
		"ayuda":  handlerHelp,
		"clima":  handlerWeather,
		"subte":  handlerMetro,
		"futbol": handlerFutbol,
	}

	bot, err := telebot.NewBot("190438378:AAHVdKCSoTTDzp3_gBtYUG8r2iWxSczJJMU")
	if err != nil {
		panic(fmt.Sprint("Error on bot creation:", err))
	}

	messages := make(chan telebot.Message)
	bot.Listen(messages, 1*time.Second)

	for message := range messages {
		if time.Unix(int64(message.Unixtime), 0).Before(startTime) {
			continue
		}

		command, params := parseCommmand(bot, message)
		handler, ok := handlers[command]
		if ok {
			err := handler(bot, message, params)
			if err != nil {
				replyMessage(bot, message, "Error al ejecutar el comando: "+command)
				fmt.Println("[error]: ", err)
			}
		} else {
			replyMessage(bot, message, "Unknown command: "+command)
		}
	}
}
