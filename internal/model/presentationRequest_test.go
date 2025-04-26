package model

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/test"
	"testing"
)

func TestCreatePresentationRequestDBEntry(t *testing.T) {
	user := "test"
	requestId := "test"
	ttl := 1
	proofRequestId := "test"
	db, mock := test.GetDBMock()
	mock.ExpectExec("SET search_path TO accounts").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectBegin()
	// Expectation for the insert command resulting in a duplicate key error
	mock.ExpectQuery("INSERT INTO \"presentation_requests\" (\"created_at\",\"updated_at\",\"deleted_at\",\"user_id\",\"request_id\",\"proof_request_id\",\"ttl\") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING \"id\"").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), user, requestId, proofRequestId, ttl).
		WillReturnError(fmt.Errorf("ERROR: duplicate key value violates unique constraint \"presentation_requests_pkey\" (SQLSTATE 23505)"))

	// Expectation for rolling back the transaction
	mock.ExpectRollback()

	err := CreatePresentationRequestDBEntry(db, user, requestId, ttl, proofRequestId)

	// Ensure the error returned is the expected PresentationAlreadyExistsError
	require.ErrorIs(t, err, PresentationAlreadyExistsError)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
