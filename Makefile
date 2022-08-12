mocks:
	mockgen -source=mail.go -destination=tests/mock_mailmerger/mail_mock.go

test:
	go test -race ./...