package main

import (
	"bunker-lite/define"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

const EnableDebug = true

func main() {
	// prepare 1
	authServerAddress := define.StdAuthServerAddress
	authServerToken := "YOUR FB TOKEN"

	// prepare 2
	engineVersion := "ENGINE VERSION (CAN BE EMPTY)"
	peAuthData := "PE AUTH DATA"
	saAuthData := "SA AUTH DATA"

	for {
		var sessionExpireTime int64
		var waitGroup sync.WaitGroup
		var ctx context.Context
		var cancel context.CancelFunc

		// Register new session
		registerActiveGuResp, err := SendAndGetHttpResponse[RegisterActiveGuResponse](
			fmt.Sprintf("%s/vitality_api/registry_active_gu", authServerAddress),
			RegisterActiveGuRequest{
				Token:              authServerToken,
				OverrideSession:    true,
				EngineVersion:      engineVersion,
				ProvidedPeAuthData: peAuthData,
				ProvidedSaAuthData: saAuthData,
			},
		)
		if err != nil {
			pterm.Warning.Printfln("[RegisterActive] %v", err)
			continue
		}

		if !registerActiveGuResp.Success {
			panic(fmt.Sprintf("RegisterActive: %v", registerActiveGuResp.ErrorInfo))
		}
		pterm.Info.Printfln("[RegisterActive] %#v", registerActiveGuResp)

		// Init variable
		sessionExpireTime = registerActiveGuResp.SessionExpireTime
		waitGroup.Add(2)
		ctx, cancel = context.WithCancel(context.Background())

		// Debug
		{
			resp, _ := SendAndGetHttpResponse[DailyGrowthResponse](
				fmt.Sprintf("%s/vitality_api/request_daily_growth", authServerAddress),
				DailyGrowthRequest{
					Token:     authServerToken,
					SessionID: registerActiveGuResp.SessionID,
				},
			)
			fmt.Printf("%#v\n", resp)
			// return
		}

		// Keep session alive
		go func() {
			defer waitGroup.Done()

			for {
				if time.Now().Unix() >= sessionExpireTime {
					cancel()
					return
				}

				nextTimeToRefreshSession := time.Unix(sessionExpireTime-600, 0)
				if nextTimeToRefreshSession.After(time.Now()) {
					timer := time.NewTimer(time.Until(nextTimeToRefreshSession))
					select {
					case <-timer.C:
					case <-ctx.Done():
						timer.Stop()
						return
					}
				}

				resp, err := SendAndGetHttpResponse[KeepGuAliveResponse](
					fmt.Sprintf("%s/vitality_api/keep_gu_alive", authServerAddress),
					KeepGuAliveRequest{
						Token:     authServerToken,
						SessionID: registerActiveGuResp.SessionID,
					},
				)
				if err != nil {
					pterm.Warning.Printfln("[KeepGuAlive] %v", err)
					continue
				}

				if !resp.Success {
					pterm.Error.Printfln("[KeepGuAlive] %v", resp.ErrorInfo)
					cancel()
					return
				}
				pterm.Success.Printfln("[KeepGuAlive] %#v", resp)

				sessionExpireTime = resp.SessionExpireTime
			}
		}()

		// Get currency online
		go func() {
			ticker := time.NewTicker(time.Minute * 5)
			defer func() {
				ticker.Stop()
				waitGroup.Done()
			}()

			for {
				resp, err := SendAndGetHttpResponse[CurrencyOnlineResponse](
					fmt.Sprintf("%s/vitality_api/get_currency_online", authServerAddress),
					CurrencyOnlineRequest{
						Token:     authServerToken,
						SessionID: registerActiveGuResp.SessionID,
					},
				)
				if err != nil {
					pterm.Warning.Printfln("[GetCurrencyOnline] %v", err)
					continue
				}

				if !resp.Success {
					pterm.Error.Printfln("[GetCurrencyOnline] %v", resp.ErrorInfo)
					cancel()
					return
				}
				pterm.Success.Printfln("[GetCurrencyOnline] %#v", resp)

				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
		}()

		// Wait all goroutine done
		waitGroup.Wait()
	}
}

// SendAndGetHttpResponse ..
func SendAndGetHttpResponse[T any](url string, request any) (result T, err error) {
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("SendAndGetHttpResponse: %v", err)
		return
	}

	buf := bytes.NewBuffer(jsonBytes)
	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		err = fmt.Errorf("SendAndGetHttpResponse: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("SendAndGetHttpResponse: Status code (%d) is not 200", resp.StatusCode)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("SendAndGetHttpResponse: %v", err)
		return
	}

	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		err = fmt.Errorf("SendAndGetHttpResponse: %v", err)
		return
	}

	return
}
