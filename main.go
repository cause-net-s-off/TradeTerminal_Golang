package main

import (
	"encoding/json"
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/gosuri/uilive"
	"github.com/guptarohit/asciigraph"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Cur map[string]CurValue

type CurValue struct {
	BuyPrice  string `json:"buy_price"`
	SellPrice string `json:"sell_price"`
	LastTrade string `json:"last_trade"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Avg       string `json:"avg"`
	Vol       string `json:"vol"`
	VolCurr   string `json:"vol_curr"`
	Updated   int64  `json:"updated"`
}

func getJSON() ([]byte, error) {
	url := "https://api.exmo.com/v1.1/ticker"
	method := "POST"

	payload := strings.NewReader("")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func UnmarshalCur(data []byte) (Cur, error) {
	var r Cur
	err := json.Unmarshal(data, &r)
	if err != nil {
		return Cur{}, err
	}

	return r, err
}

func (r *Cur) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func getDateTime() string {
	res := ""
	times := time.Now()
	res += times.Format("Текущая дата: 2006-01-02") + "\n"
	res += times.Format("Текущее время: 15:04:05")
	return res
}

func getCurrentPrice(pair string) (float64, error) {
	jsonData, err := getJSON()
	if err != nil {
		return 0.0, err
	}
	obj, err := UnmarshalCur(jsonData)
	_ = obj
	if err != nil {
		return 0.0, err
	}
	toRound, err := strconv.ParseFloat(obj[pair].BuyPrice, 64)
	if err != nil {
		return 0.0, err
	}
	rounded := fmt.Sprintf("%0.2f", toRound)
	res, err := strconv.ParseFloat(rounded, 64)
	if err != nil {
		return 0.0, err
	}

	//Имитация волатильности, для отладки, добавление амплитуды
	fastChanger := float64(rand.Intn(2)) + float64(rand.Intn(1000))/1000

	fastChanger = 0.0
	res += fastChanger

	return res, err
}

func renderCoin(ch <-chan keyboard.KeyEvent, writer *uilive.Writer, name string) {

	data := make([]float64, 100)
	initialPrice, _ := getCurrentPrice(name)
	for i := range data {
		data[i] = initialPrice
	}
	var graph string
	toExit := false
	for !toExit {
		select {
		case event := <-ch:
			if event.Key == keyboard.KeyBackspace2 {
				toExit = true
			}
		default:
		}

		num, err := getCurrentPrice(name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		data = data[1:]
		data = append(data, num)

		graph = asciigraph.Plot(data,
			asciigraph.Height(8),
			asciigraph.Precision(5),
			asciigraph.Width(100),
			asciigraph.SeriesColors(asciigraph.Red))
		//	str := name + "\n" + graph
		//fmt.Fprintf(writer.Newline(), "Downloading %s.. (%d/%d) GB\n", f[1], i, 50)
		//_, _ = fmt.Fprintf(writer.Bypass(), "%s: %f\n", name, num)
		//_, _ = fmt.Fprintf(writer.Bypass(), "%s\n", graph)
		_, _ = fmt.Fprintln(writer, name, num)
		//_, _ = fmt.Fprintf(writer.Newline(), "%s\n", "")
		_, _ = fmt.Fprintln(writer, graph)
		_, _ = fmt.Fprintln(writer, getDateTime())
		//fmt.Fprintln(writer, len(data), cap(data))

		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {
	writer := uilive.New()
	// start listening for updates and render
	writer.Start()
	defer writer.Stop() // flush and stop rendering

	getMenu := func() {
		//	fmt.Print("\033[H\033[2J") // очистка терминала
		fmt.Fprintf(writer, "1. BTC_USD\n")
		fmt.Fprintf(writer, "2. LTC_USD\n")
		fmt.Fprintf(writer, "3. ETH_USD\n")
		fmt.Fprintf(writer, "\n")
		fmt.Fprintf(writer, "Press 1-3 to change symbol, press q to exit\n")
	}

	keysEvents, err := keyboard.GetKeys(2)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	getMenu()
	toExit := false
	for !toExit {
		select {
		case event := <-keysEvents:
			switch event.Rune {
			case '1':
				//	fmt.Print("\033[H\033[2J") // очистка терминала
				renderCoin(keysEvents, writer, "BTC_USD")
				getMenu()
			case '2':
				//	fmt.Print("\033[H\033[2J") // очистка терминала
				renderCoin(keysEvents, writer, "LTC_USD")
				getMenu()
			case '3':
				//	fmt.Print("\033[H\033[2J") // очистка терминала
				renderCoin(keysEvents, writer, "ETH_USD")
				getMenu()
			case 'q':
				//	fmt.Print("\033[H\033[2J") // очистка терминала
				toExit = true
				//	fmt.Fprintf(writer, "EXIT\n")
			}
		default:

		}
	}

}
