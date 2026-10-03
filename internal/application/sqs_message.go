package application

type WagerTransactionMessage struct {
	MessageID     string                      `json:"messageId"`
	CorrelationID string                      `json:"correlationId"`
	Data          WagerTransactionMessageData `json:"data"`
}

type WagerTransactionMessageData struct {
	IdempotencyKey                 string `json:"idempotencyKey"`
	TransactionID                  string `json:"transactionId"`
	PlayerID                       string `json:"playerId"`
	ProviderID                     string `json:"providerId"`
	ExternalTransactionID          string `json:"externalTransactionId"`
	WalletID                       string `json:"walletId"`
	RoundID                        string `json:"roundId"`
	GameID                         string `json:"gameId"`
	Type                           string `json:"type"`
	Amount                         int64  `json:"amount"`
	Currency                       string `json:"currency"`
	ReferenceExternalTransactionID string `json:"referenceExternalTransactionId,omitempty"`
}
