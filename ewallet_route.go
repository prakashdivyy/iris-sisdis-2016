package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	iris "gopkg.in/kataras/iris.v4"
)

type JSONRegister struct {
	UserID     string `json:"user_id"`
	Nama       string `json:"nama"`
	IPDomisili string `json:"ip_domisili"`
}

type JSONTransfer struct {
	UserID string `json:"user_id"`
	Nilai  int    `json:"nilai"`
}

type JSONPing struct {
	Pong int `json:"pong"`
}

type JSONSaldo struct {
	NilaiSaldo int `json:"nilai_saldo"`
}

func ewalletPing(c *iris.Context) {
	c.JSON(iris.StatusOK, iris.Map{"pong": 1})
}

func ewalletGetSaldo(c *iris.Context) {
	if checkQuorum() >= 5 {
		userID := c.URLParam("user_id")
		var nilaiSaldo int
		db, err := sql.Open("mysql", "root:password2016@/ewallet")
		stmtOut, err := db.Prepare("SELECT saldo FROM data_pengguna WHERE id = ?")
		err = stmtOut.QueryRow(userID).Scan(&nilaiSaldo)
		if err != nil {
			nilaiSaldo = -1
		}
		db.Close()
		c.JSON(iris.StatusOK, iris.Map{"nilai_saldo": nilaiSaldo})
	} else {
		c.JSON(iris.StatusOK, iris.Map{"error": "Quorum tidak tercapai"})
	}
}

func ewalletGetTotalSaldo(c *iris.Context) {
	if checkQuorum() >= 8 {
		userID := c.URLParam("user_id")
		var ip string
		db, err := sql.Open("mysql", "root:password2016@/ewallet")
		stmt, _ := db.Prepare("SELECT ip FROM data_pengguna WHERE id = ?")
		err = stmt.QueryRow(userID).Scan(&ip)
		db.Close()
		if err != nil {
			c.JSON(iris.StatusOK, iris.Map{"nilai_saldo": -1})
			return
		}
		if ip == "prakash.sisdis.ui.ac.id" || ip == "152.118.33.98" {
			total := 0
			listIP := [9]string{"prakash.sisdis.ui.ac.id", "aditya.sisdis.ui.ac.id", "ratna.sisdis.ui.ac.id", "azhari.sisdis.ui.ac.id", "kurniawan.sisdis.ui.ac.id", "alhafis.sisdis.ui.ac.id", "putra.sisdis.ui.ac.id", "radityo.sisdis.ui.ac.id", "ilham.sisdis.ui.ac.id"}
			for x := 0; x < 9; x++ {
				request, err := http.Get("https://" + listIP[x] + "/ewallet/getSaldo?user_id=" + userID)
				if err != nil {
				} else {
					decodeJSON := json.NewDecoder(request.Body)
					var data JSONSaldo
					err = decodeJSON.Decode(&data)
					if err != nil {
					} else {
						if data.NilaiSaldo != -1 {
							total += data.NilaiSaldo
						}
					}
				}
			}
			c.JSON(iris.StatusOK, iris.Map{"nilai_saldo": total})
		} else {
			request, err := http.Get("https://" + ip + "/ewallet/getTotalSaldo?user_id=" + userID)
			if err != nil {
			} else {
				decodeJSON := json.NewDecoder(request.Body)
				var data JSONSaldo
				err = decodeJSON.Decode(&data)
				if err != nil {
				} else {
					c.JSON(iris.StatusOK, iris.Map{"nilai_saldo": data.NilaiSaldo})
				}
			}
		}
	} else {
		c.JSON(iris.StatusOK, iris.Map{"error": "Quorum tidak tercapai"})
	}
}

func ewalletRegister(c *iris.Context) {
	if checkQuorum() >= 5 {
		req := new(JSONRegister)
		c.ReadJSON(req)
		db, err := sql.Open("mysql", "root:password2016@/ewallet")
		stmt, _ := db.Prepare("INSERT data_pengguna SET id=?,ip=?,nama=?,saldo=?")
		_, err = stmt.Exec(req.UserID, req.IPDomisili, req.Nama, 1000000)
		db.Close()
		if err != nil {
			c.JSON(iris.StatusOK, iris.Map{"error": "Already Registered"})
		}
	} else {
		c.JSON(iris.StatusOK, iris.Map{"error": "Quorum tidak tercapai"})
	}
}

func ewalletTransfer(c *iris.Context) {
	if checkQuorum() >= 5 {
		req := new(JSONTransfer)
		c.ReadJSON(req)
		request, err := http.Get("https://prakash.sisdis.ui.ac.id/ewallet/getSaldo?user_id=" + req.UserID)
		if err != nil {
			c.JSON(iris.StatusOK, iris.Map{"status_transfer": -1})
			return
		}
		decodeJSON := json.NewDecoder(request.Body)
		var data JSONSaldo
		_ = decodeJSON.Decode(&data)
		if data.NilaiSaldo == -1 {
			c.JSON(iris.StatusOK, iris.Map{"status_transfer": -1})
			return
		}
		db, err := sql.Open("mysql", "root:password2016@/ewallet")
		stmt, _ := db.Prepare("UPDATE data_pengguna SET saldo=? where id=?")
		_, err = stmt.Exec(req.Nilai+data.NilaiSaldo, req.UserID)
		db.Close()
		if err != nil {
			c.JSON(iris.StatusOK, iris.Map{"status_transfer": -1})
		} else {
			c.JSON(iris.StatusOK, iris.Map{"status_transfer": 0})
		}
	} else {
		c.JSON(iris.StatusOK, iris.Map{"error": "Quorum tidak tercapai"})
	}
}

func checkQuorum() (total int) {
	total = 0
	listIP := [9]string{"prakash.sisdis.ui.ac.id", "aditya.sisdis.ui.ac.id", "ratna.sisdis.ui.ac.id", "azhari.sisdis.ui.ac.id", "kurniawan.sisdis.ui.ac.id", "alhafis.sisdis.ui.ac.id", "putra.sisdis.ui.ac.id", "radityo.sisdis.ui.ac.id", "ilham.sisdis.ui.ac.id"}
	for x := 0; x < 9; x++ {
		request, err := http.Get("https://" + listIP[x] + "/ewallet/ping")
		if err != nil {
		} else {
			decodeJSON := json.NewDecoder(request.Body)
			var data JSONPing
			err = decodeJSON.Decode(&data)
			if err != nil {
			} else {
				total += data.Pong
			}
		}
	}
	return
}
