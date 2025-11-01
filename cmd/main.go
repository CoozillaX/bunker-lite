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
		registerActiveSessionResp, err := SendAndGetHttpResponse[RegisterActiveSessionResponse](
			fmt.Sprintf("%s/register_active_session", authServerAddress),
			RegisterActiveSessionRequest{
				Token:              authServerToken,
				OverrideSession:    true,
				ProvidedPeAuthData: peAuthData,
				ProvidedSaAuthData: saAuthData,
			},
		)
		if err != nil {
			pterm.Warning.Printfln("[RegisterActiveSession] %v", err)
			continue
		}
		if !registerActiveSessionResp.Success {
			panic(fmt.Sprintf("main/RegisterActiveSession: %v", registerActiveSessionResp.ErrorInfo))
		}
		pterm.Info.Printfln("[RegisterActiveSession] %#v", registerActiveSessionResp)

		// Init variable
		sessionExpireTime = registerActiveSessionResp.SessionExpireTime
		waitGroup.Add(2)
		ctx, cancel = context.WithCancel(context.Background())

		// Debug
		if true {
			resp, _ := SendAndGetHttpResponse[VitalityDebugResponse](
				fmt.Sprintf("%s/request_vitality_debug", authServerAddress),
				VitalityDebugRequest{
					Token:       authServerToken,
					SessionID:   registerActiveSessionResp.SessionID,
					RequestType: RequestTypeGetCurrencyOnline,
				},
			)
			fmt.Printf("%#v\n", resp)
			// return
		}
		if true {
			resp, _ := SendAndGetHttpResponse[VitalityDebugResponse](
				fmt.Sprintf("%s/request_vitality_debug", authServerAddress),
				VitalityDebugRequest{
					Token:       authServerToken,
					SessionID:   registerActiveSessionResp.SessionID,
					RequestType: RequestTypeGetDailyGrowth,
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

				nextTimeToRefreshSession := time.Unix(sessionExpireTime-30*60, 0)
				if nextTimeToRefreshSession.After(time.Now()) {
					timer := time.NewTimer(time.Until(nextTimeToRefreshSession))
					select {
					case <-timer.C:
					case <-ctx.Done():
						timer.Stop()
						return
					}
				}

				resp, err := SendAndGetHttpResponse[KeepSessionAliveResponse](
					fmt.Sprintf("%s/keep_session_alive", authServerAddress),
					KeepSessionAliveRequest{
						Token:     authServerToken,
						SessionID: registerActiveSessionResp.SessionID,
					},
				)
				if err != nil {
					pterm.Warning.Printfln("[KeepSessionAlive] %v", err)
					continue
				}

				if !resp.Success {
					switch resp.ErrorType {
					case KeepSessionAliveErrorMeetError:
						pterm.Error.Printfln("[KeepSessionAlive] %v", resp.ErrorInfo)
					case KeepSessionAliveErrorLifeLimit:
						pterm.Info.Printfln("[KeepSessionAlive] This session reach its max life time limit. Do refresh.")
						cleanUpSession(authServerAddress, authServerToken, registerActiveSessionResp.SessionID)
					}
					cancel()
					return
				}
				pterm.Success.Printfln("[KeepSessionAlive] %#v", resp)

				sessionExpireTime = resp.SessionExpireTime
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
