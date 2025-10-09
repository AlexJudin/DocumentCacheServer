package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetWalletBalanceByUUIDWhenOk(t *testing.T) {
	truncateTable(walletTest.db)

	err := walletTest.db.Exec(`INSERT INTO payment_operations
	(wallet_id,  operation_type, amount) VALUES
	( 'ec82ea03-2b53-4258-ba87-a7efae979c43', 'deposit', 5000),
	( 'ec82ea03-2b53-4258-ba87-a7efae979c43', 'withdraw', -1000),
	( '7a6f774f-2f85-4e61-8830-833aaec60f14', 'deposit', 3500);`).Error
	if err != nil {
		t.Error(err)
		return
	}

	req := httptest.NewRequest("GET", "/api/v1/wallets/?WALLET_UUID=ec82ea03-2b53-4258-ba87-a7efae979c43", nil)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.GetWalletBalanceByUUID)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusOK {
		t.Errorf("expected status code: %d, got %d", http.StatusOK, status)
	}

	expected := `{"balance":4000}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestGetWalletBalanceByUUIDWhenWalletUUIDIsEmpty(t *testing.T) {
	truncateTable(walletTest.db)

	req := httptest.NewRequest("GET", "/api/v1/wallets/?WALLET_UUID=", nil)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.GetWalletBalanceByUUID)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("expected status code: %d, got %d", http.StatusBadRequest, status)
	}

	expected := `{"error":"Не передан идентификатор кошелька, получение баланса невозможно."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestGetWalletBalanceByUUIDWhenMissingWalletUUID(t *testing.T) {
	truncateTable(walletTest.db)

	req := httptest.NewRequest("GET", "/api/v1/wallets/", nil)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.GetWalletBalanceByUUID)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("expected status code: %d, got %d", http.StatusBadRequest, status)
	}

	expected := `{"error":"Не передан идентификатор кошелька, получение баланса невозможно."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestCreateOperationWhenOk(t *testing.T) {
	truncateTable(walletTest.db)

	bodyJSON := `{"walletId":"ec82ea03-2b53-4258-ba87-a7efae979c43", "operationType":"deposit", "amount": 4000}`
	req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(bodyJSON))

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.CreateOperation)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusCreated {
		t.Errorf("expected status code: %d, got %d", http.StatusCreated, status)
	}
}

func TestCreateOperationWhenBodyIsEmpty(t *testing.T) {
	truncateTable(walletTest.db)

	req := httptest.NewRequest("POST", "/api/v1/wallet", nil)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.CreateOperation)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("expected status code: %d, got %d", http.StatusBadRequest, status)
	}

	expected := `{"error":"Не удалось прочитать данные о платежной операции."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestCreateOperationWhenUncorrectBody(t *testing.T) {
	truncateTable(walletTest.db)

	bodyJSON := `{"walletId":"ec82ea03-2b53-4258-ba87-a7efae979c43", "operationType":1, "amount": 4000}`
	req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(bodyJSON))

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.CreateOperation)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("expected status code: %d, got %d", http.StatusBadRequest, status)
	}

	expected := `{"error":"Не удалось прочитать данные о платежной операции."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestCreateOperationWhenOperationTypeIsEmpty(t *testing.T) {
	truncateTable(walletTest.db)

	bodyJSON := `{"walletId":"ec82ea03-2b53-4258-ba87-a7efae979c43", "amount": 4000}`
	req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(bodyJSON))

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.CreateOperation)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("expected status code: %d, got %d", http.StatusBadRequest, status)
	}

	expected := `{"error":"В данных о платежной операции переданы некорректные поля [operationType]."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestCreateOperationWhenAmountIsNegative(t *testing.T) {
	truncateTable(walletTest.db)

	bodyJSON := `{"walletId":"ec82ea03-2b53-4258-ba87-a7efae979c43", "operationType":"deposit", "amount": -4000}`
	req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(bodyJSON))

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.CreateOperation)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("expected status code: %d, got %d", http.StatusBadRequest, status)
	}

	expected := `{"error":"В данных о платежной операции переданы некорректные поля [amount]."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestCreateOperationWhenOperationTypeIsEmptyAndAmountIsNegative(t *testing.T) {
	truncateTable(walletTest.db)

	bodyJSON := `{"walletId":"ec82ea03-2b53-4258-ba87-a7efae979c43", "amount": -4000}`
	req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(bodyJSON))

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.CreateOperation)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusBadRequest {
		t.Errorf("expected status code: %d, got %d", http.StatusBadRequest, status)
	}

	expected := `{"error":"В данных о платежной операции переданы некорректные поля [operationType, amount]."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestCreateOperationWhenInsufficientFunds(t *testing.T) {
	truncateTable(walletTest.db)

	bodyJSON := `{"walletId":"ec82ea03-2b53-4258-ba87-a7efae979c43", "operationType":"withdraw", "amount": 4000}`
	req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(bodyJSON))

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.CreateOperation)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusOK {
		t.Errorf("expected status code: %d, got %d", http.StatusOK, status)
	}

	expected := `{"error":"Недостаточно средств."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

// When missing connect to database
func TestGetWalletBalanceByUUIDWhenMissingConnectToDB(t *testing.T) {
	err := closeDB()
	if err != nil {
		t.Error(err)
		return
	}

	req := httptest.NewRequest("GET", "/api/v1/wallets/?WALLET_UUID=ec82ea03-2b53-4258-ba87-a7efae979c43", nil)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.GetWalletBalanceByUUID)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusInternalServerError {
		t.Errorf("expected status code: %d, got %d", http.StatusInternalServerError, status)
	}

	expected := `{"error":"Ошибка сервера, не удалось получить баланс. Попробуйте позже или обратитесь в тех. поддержку."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}

func TestCreateOperationWhenMissingConnectToDB(t *testing.T) {
	err := closeDB()
	if err != nil {
		t.Error(err)
		return
	}

	bodyJSON := `{"walletId":"ec82ea03-2b53-4258-ba87-a7efae979c43", "operationType":"deposit", "amount": 4000}`
	req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(bodyJSON))

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(walletTest.handler.CreateOperation)
	handler.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != http.StatusInternalServerError {
		t.Errorf("expected status code: %d, got %d", http.StatusInternalServerError, status)
	}

	expected := `{"error":"Ошибка сервера, не удалось сохранить данные о платежной операции. Попробуйте позже или обратитесь в тех. поддержку."}`
	if responseRecorder.Body.String() != expected {
		t.Errorf("expected body: %s, got %s", expected, responseRecorder.Body.String())
	}
}
