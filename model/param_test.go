package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHanldeResponse(t *testing.T) {
	handler, err := NewHanldeBodyFromResponse("eyJib2R5Ijp7Im5vX3JlcGxheV90b2tlbiI6W1syNDcsMTI0LDE2MCwyNDUsNDAsMjcsMTQxLDE3OF0sMTI1NjAwXSwib3BlcmF0aW9ucyI6W3siVHJhbnNmZXJBc3NldCI6eyJib2R5Ijp7ImlucHV0cyI6W3siQWJzb2x1dGUiOjI5NTU4OX0seyJBYnNvbHV0ZSI6Mjk1NTkwfSx7IkFic29sdXRlIjoyOTcwNTh9LHsiQWJzb2x1dGUiOjI5NzA1OX0seyJBYnNvbHV0ZSI6Mjk3MDU3fV0sInBvbGljaWVzIjp7InZhbGlkIjp0cnVlLCJpbnB1dHNfdHJhY2luZ19wb2xpY2llcyI6W1tdLFtdLFtdLFtdLFtdXSwiaW5wdXRzX3NpZ19jb21taXRtZW50cyI6W251bGwsbnVsbCxudWxsLG51bGwsbnVsbF0sIm91dHB1dHNfdHJhY2luZ19wb2xpY2llcyI6W1tdLFtdLFtdXSwib3V0cHV0c19zaWdfY29tbWl0bWVudHMiOltudWxsLG51bGwsbnVsbF19LCJvdXRwdXRzIjpbeyJpZCI6bnVsbCwicmVjb3JkIjp7ImFtb3VudCI6eyJOb25Db25maWRlbnRpYWwiOiI0MDM4MDMyMjYwIn0sImFzc2V0X3R5cGUiOnsiTm9uQ29uZmlkZW50aWFsIjpbMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwXX0sInB1YmxpY19rZXkiOiJBUUVCQVFFQkFRRUJBUUVCQVFFQkFRRUJBUUVCQVFFQkFRRUJBUUVCQVFFPSJ9fSx7ImlkIjpudWxsLCJyZWNvcmQiOnsiYW1vdW50Ijp7Ik5vbkNvbmZpZGVudGlhbCI6IjMxMDk2NiJ9LCJhc3NldF90eXBlIjp7Ik5vbkNvbmZpZGVudGlhbCI6WzAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMF19LCJwdWJsaWNfa2V5IjoiY0dhajZoSDU5ZUxYVGdrZXBRYjBzVnhDRUdTaVBWeXVPUi16aVBiSlgyUT0ifX0seyJpZCI6bnVsbCwicmVjb3JkIjp7ImFtb3VudCI6eyJOb25Db25maWRlbnRpYWwiOiIxMDAwMCJ9LCJhc3NldF90eXBlIjp7Ik5vbkNvbmZpZGVudGlhbCI6WzAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMF19LCJwdWJsaWNfa2V5IjoiQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQT0ifX1dLCJ0cmFuc2ZlciI6eyJpbnB1dHMiOlt7ImFtb3VudCI6eyJOb25Db25maWRlbnRpYWwiOiI5OTIzMTIifSwiYXNzZXRfdHlwZSI6eyJOb25Db25maWRlbnRpYWwiOlswLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDBdfSwicHVibGljX2tleSI6ImNHYWo2aEg1OWVMWFRna2VwUWIwc1Z4Q0VHU2lQVnl1T1ItemlQYkpYMlE9In0seyJhbW91bnQiOnsiTm9uQ29uZmlkZW50aWFsIjoiMTg2ODc2MDA0MSJ9LCJhc3NldF90eXBlIjp7Ik5vbkNvbmZpZGVudGlhbCI6WzAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMF19LCJwdWJsaWNfa2V5IjoiY0dhajZoSDU5ZUxYVGdrZXBRYjBzVnhDRUdTaVBWeXVPUi16aVBiSlgyUT0ifSx7ImFtb3VudCI6eyJOb25Db25maWRlbnRpYWwiOiIxODEzNTk0MjI0In0sImFzc2V0X3R5cGUiOnsiTm9uQ29uZmlkZW50aWFsIjpbMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwXX0sInB1YmxpY19rZXkiOiJjR2FqNmhINTllTFhUZ2tlcFFiMHNWeENFR1NpUFZ5dU9SLXppUGJKWDJRPSJ9LHsiYW1vdW50Ijp7Ik5vbkNvbmZpZGVudGlhbCI6IjE2NzU5OTUzNiJ9LCJhc3NldF90eXBlIjp7Ik5vbkNvbmZpZGVudGlhbCI6WzAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMF19LCJwdWJsaWNfa2V5IjoiY0dhajZoSDU5ZUxYVGdrZXBRYjBzVnhDRUdTaVBWeXVPUi16aVBiSlgyUT0ifSx7ImFtb3VudCI6eyJOb25Db25maWRlbnRpYWwiOiIxODc0MDcxMTMifSwiYXNzZXRfdHlwZSI6eyJOb25Db25maWRlbnRpYWwiOlswLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDBdfSwicHVibGljX2tleSI6ImNHYWo2aEg1OWVMWFRna2VwUWIwc1Z4Q0VHU2lQVnl1T1ItemlQYkpYMlE9In1dLCJvdXRwdXRzIjpbeyJhbW91bnQiOnsiTm9uQ29uZmlkZW50aWFsIjoiNDAzODAzMjI2MCJ9LCJhc3NldF90eXBlIjp7Ik5vbkNvbmZpZGVudGlhbCI6WzAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMF19LCJwdWJsaWNfa2V5IjoiQVFFQkFRRUJBUUVCQVFFQkFRRUJBUUVCQVFFQkFRRUJBUUVCQVFFQkFRRT0ifSx7ImFtb3VudCI6eyJOb25Db25maWRlbnRpYWwiOiIzMTA5NjYifSwiYXNzZXRfdHlwZSI6eyJOb25Db25maWRlbnRpYWwiOlswLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDBdfSwicHVibGljX2tleSI6ImNHYWo2aEg1OWVMWFRna2VwUWIwc1Z4Q0VHU2lQVnl1T1ItemlQYkpYMlE9In0seyJhbW91bnQiOnsiTm9uQ29uZmlkZW50aWFsIjoiMTAwMDAifSwiYXNzZXRfdHlwZSI6eyJOb25Db25maWRlbnRpYWwiOlswLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDBdfSwicHVibGljX2tleSI6IkFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE9In1dLCJwcm9vZnMiOnsiYXNzZXRfdHlwZV9hbmRfYW1vdW50X3Byb29mIjoiTm9Qcm9vZiIsImFzc2V0X3RyYWNpbmdfcHJvb2YiOnsiYXNzZXRfdHlwZV9hbmRfYW1vdW50X3Byb29mcyI6W10sImlucHV0c19pZGVudGl0eV9wcm9vZnMiOltbXSxbXSxbXSxbXSxbXV0sIm91dHB1dHNfaWRlbnRpdHlfcHJvb2ZzIjpbW10sW10sW11dfX0sImFzc2V0X3RyYWNpbmdfbWVtb3MiOltbXSxbXSxbXSxbXSxbXSxbXSxbXSxbXV0sIm93bmVyc19tZW1vcyI6W251bGwsbnVsbCxudWxsXX0sInRyYW5zZmVyX3R5cGUiOiJTdGFuZGFyZCJ9LCJib2R5X3NpZ25hdHVyZXMiOlt7ImFkZHJlc3MiOnsia2V5IjoiY0dhajZoSDU5ZUxYVGdrZXBRYjBzVnhDRUdTaVBWeXVPUi16aVBiSlgyUT0ifSwic2lnbmF0dXJlIjoidDMwd1BTQUI3a1JBb29OaG1iY2J4LWhsci1BU2FUN2FiR0FpbWF6QXB2UWs0LXJTc3ZPY0Uwc1ZVRXVSTjczSlhzaU12UVpjRTBpMExCWFpQTHZLRGc9PSJ9XX19LHsiRGVsZWdhdGlvbiI6eyJib2R5Ijp7InZhbGlkYXRvciI6IjI2YWE3NTgxMjYzMzMyZjQ3ZTBjZTE3Y2Y0YjFmMzRkMjJjN2Y0Y2IiLCJuZXdfdmFsaWRhdG9yIjpudWxsLCJhbW91bnQiOjQwMzgwMzIyNjAsIm5vbmNlIjpbWzI0NywxMjQsMTYwLDI0NSw0MCwyNywxNDEsMTc4XSwxMjU2MDBdfSwicHVia2V5IjoiY0dhajZoSDU5ZUxYVGdrZXBRYjBzVnhDRUdTaVBWeXVPUi16aVBiSlgyUT0iLCJzaWduYXR1cmUiOiJRWjRwbnJoWTcyZjFMUVJPdnNNZUxzdE04T082TFRSMzZxa2NydWhFa3ozWUZ6TjZCX1l3RWQ4Z3d3NnV1OERwLXVNcFVBSUZYdnNmS1NUSnprS3JCQT09Iiwidl9zaWduYXR1cmUiOm51bGx9fV19LCJzaWduYXR1cmVzIjpbImNlT2o0ZkJqQXRXZ2QtYjZtckZobU5xN21sZHNrUkFqNzVYNXFSRWpvVVZqLXhEM1FpYmlDaEtZZkZlQ1FiRUxIcFFXNWZyV0FXcWR2YmVKZ3EtVEFBPT0iXX0=")
	assert.Nil(t, err)
	_ = handler
	assert.Equal(t, 5, len(handler.Body.Operations[0].TransferAsset.Body.Transfer.Inputs))
}

func TestTickerInfo(t *testing.T) {
	jsonStr := `{
		"code": 0,
		"msg": "ok",
		"data": {
			"ticker": "sats",
			"overallBalance": "10",
			"transferableBalance": "0",
			"availableBalance": "10",
			"availableBalanceSafe": "10",
			"availableBalanceUnSafe": "0",
			"transferableCount": 0,
			"transferableInscriptions": [],
			"historyCount": 1,
			"historyInscriptions": [
				{
					"data": {
						"op": "mint",
						"tick": "sats",
						"amt": "10",
						"minted": "10"
					},
					"inscriptionNumber": 1793178,
					"inscriptionId": "865fdf1878f19aebff422ca15b70dccc7e39fd79c59ba30edd2821d9de6e3b3fi0",
					"satoshi": 546,
					"confirmations": 0
				}
			]
		}
	}`

	var balanceResp UnisatResponse[UnisatTickerInfo]
	json.Unmarshal([]byte(jsonStr), &balanceResp)
	assert.Equal(t, "10", balanceResp.Data.OverallBalance)
}