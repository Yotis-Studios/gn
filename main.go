package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Connected")
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// handle connection error
			fmt.Fprintln(os.Stderr, err)
		}
		go func() {
			defer conn.Close()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				fmt.Println(op)
				if err != nil {
					// handle read error
					fmt.Fprintln(os.Stderr, err)
					break
				}
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					// handle write error
					fmt.Fprintln(os.Stderr, err)
					break
				}
			}
		}()
	}))
}
