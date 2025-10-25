package main

import (
	"bunker-lite/service/define"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

func main() {
	// prepare 1
	authServerAddress := define.AddressVitalityAPI
	authServerToken := "YOUR FB TOKEN"

	// prepare 2
	peAuthData := "PE AUTH DATA"
	saAuthData := "SA AUTH DATA"

	for {
		var sessionExpireTime int64
		var waitGroup sync.WaitGroup
		var ctx context.Context
		var cancel context.CancelFunc

		// Register new session
		registerActiveGuResp, err := SendAndGetHttpResponse[RegisterActiveGuResponse](
			fmt.Sprintf("%s/registry_active_gu", authServerAddress),
			RegisterActiveGuRequest{
				Token:              authServerToken,
				OverrideSession:    true,
				ProvidedPeAuthData: peAuthData,
				ProvidedSaAuthData: saAuthData,
			},
		)
		if err != nil {
			pterm.Warning.Printfln("[RegisterActiveGu] %v", err)
			continue
		}
		if !registerActiveGuResp.Success {
			panic(fmt.Sprintf("main/RegisterActiveGu: %v", registerActiveGuResp.ErrorInfo))
		}
		pterm.Info.Printfln("[RegisterActiveGu] %#v", registerActiveGuResp)

		// Init variable
		sessionExpireTime = registerActiveGuResp.SessionExpireTime
		waitGroup.Add(2)
		ctx, cancel = context.WithCancel(context.Background())

		// Debug
		if true {
			resp, _ := SendAndGetHttpResponse[SessionInfoResponse](
				fmt.Sprintf("%s/request_session_info", authServerAddress),
				SessionInfoRequest{
					Token: authServerToken,
				},
			)
			fmt.Printf("%#v\n", resp)
			// return
		}
		if true {
			resp, _ := SendAndGetHttpResponse[DailyGrowthResponse](
				fmt.Sprintf("%s/request_daily_growth", authServerAddress),
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
					fmt.Sprintf("%s/keep_gu_alive", authServerAddress),
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
					switch resp.ErrorType {
					case KeepGuAliveErrorMeetError:
						pterm.Error.Printfln("[KeepGuAlive] %v", resp.ErrorInfo)
					case KeepGuAliveErrorLifeLimit:
						pterm.Info.Printfln("[KeepGuAlive] This session reach its max life time limit. Do refresh.")
						cleanUpSession(authServerAddress, authServerToken, registerActiveGuResp.SessionID)
					}
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
					fmt.Sprintf("%s/get_currency_online", authServerAddress),
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

func cleanUpSession(address string, token string, sessionID string) {
	for {
		resp, err := SendAndGetHttpResponse[CleanUpSessionResponse](
			fmt.Sprintf("%s/clean_up_session", address),
			CleanUpSessionRequest{
				Token:     token,
				SessionID: sessionID,
			},
		)
		if err != nil {
			pterm.Warning.Printfln("[CleanUpSession] %v", err)
			continue
		}

		if resp.Success {
			pterm.Success.Printfln("[CleanUpSession] %#v", resp)
		} else {
			pterm.Error.Printfln("[CleanUpSession] %v", err)
		}

		break
	}
}
