// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package lib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransactionDetail(t *testing.T) {

	record := transactionDetail{
		TypeCode: "890",
	}
	require.NoError(t, record.validate())

	record.TypeCode = "AAA"
	require.Error(t, record.validate())
	require.Equal(t, "TransactionDetail: invalid TypeCode", record.validate().Error())

}

func TestTransactionDetailWithSample(t *testing.T) {

	sample := "16,409,000000000002500,V,060316,,,,RETURNED CHEQUE     /"
	record := transactionDetail{
		TypeCode: "890",
	}

	size, err := record.parse(sample)
	require.NoError(t, err)
	require.Equal(t, 56, size)

	require.Equal(t, "409", record.TypeCode)
	require.Equal(t, "000000000002500", record.Amount)
	require.Equal(t, "V", string(record.FundsType.TypeCode))
	require.Equal(t, "060316", record.FundsType.Date)
	require.Equal(t, "", record.FundsType.Time)
	require.Equal(t, "", record.BankReferenceNumber)
	require.Equal(t, "", record.CustomerReferenceNumber)
	require.Equal(t, "RETURNED CHEQUE     ", record.Text)

	require.Equal(t, sample, record.string())
}

func TestTransactionDetailOutputWithContinuationRecords(t *testing.T) {

	record := transactionDetail{
		TypeCode:                "409",
		Amount:                  "111111111111111",
		BankReferenceNumber:     "222222222222222",
		CustomerReferenceNumber: "333333333333333",
		Text:                    "RETURNED CHEQUE     444444444444444",
		FundsType: FundsType{
			TypeCode:           FundsTypeD,
			DistributionNumber: 5,
			Distributions: []Distribution{
				{
					Day:    1,
					Amount: 1000000000,
				},
				{
					Day:    2,
					Amount: 2000000000,
				},
				{
					Day:    3,
					Amount: 3000000000,
				},
				{
					Day:    4,
					Amount: 4000000000,
				},
				{
					Day:    5,
					Amount: 5000000000,
				},
				{
					Day:    6,
					Amount: 6000000000,
				},
				{
					Day:    7,
					Amount: 7000000000,
				},
			},
		},
	}

	result := record.string()
	expectResult := `16,409,111111111111111,D,5,1,1000000000,2,2000000000,3,3000000000,4,4000000000,5,5000000000,6,6000000000,7,7000000000,222222222222222,333333333333333,RETURNED CHEQUE     444444444444444/`
	require.Equal(t, expectResult, result)
	require.Equal(t, len(expectResult), len(result))

	result = record.string(80)
	expectResult = `16,409,111111111111111,D,5,1,1000000000,2,2000000000,3,3000000000,4,4000000000/
88,5,5000000000,6,6000000000,7,7000000000,222222222222222,333333333333333/
88,RETURNED CHEQUE     444444444444444/`
	require.Equal(t, expectResult, result)
	require.Equal(t, len(expectResult), len(result))

	result = record.string(50)
	expectResult = `16,409,111111111111111,D,5,1,1000000000,2/
88,2000000000,3,3000000000,4,4000000000,5/
88,5000000000,6,6000000000,7,7000000000/
88,222222222222222,333333333333333/
88,RETURNED CHEQUE     444444444444444/`
	require.Equal(t, expectResult, result)
	require.Equal(t, len(expectResult), len(result))

}

/**
 * This test outlines the behavior of a Detail record when the Detail includes Continuation data where the
 * Continuation(s) don't match a defined type. The Detail record is parsed but the Continuation data is not
 * currently retained.
 */
func TestTransactionDetailOutputWithContinuationRecords_AdHocContinuationFormat(t *testing.T) {
	data := `16,266,1912,,GI2118700002010,20210706MMQFMPU8000001,Outgoing Wire Return,-/
88,CREF: 20210706MMQFMPU8000001/
88,EREF: 20210706MMQFMPU8000001/
88,DBIC: GSCRUS33/
88,CRNM: ABC Company/
88,DBNM: SAMPLE INC./`

	record := transactionDetail{}

	size, err := record.parse(data)
	require.NoError(t, err)

	require.Equal(t, "266", record.TypeCode)
	require.Equal(t, "1912", record.Amount)
	require.Equal(t, "", string(record.FundsType.TypeCode))
	require.Equal(t, "", record.FundsType.Date)
	require.Equal(t, "", record.FundsType.Time)
	require.Equal(t, "GI2118700002010", record.BankReferenceNumber)
	require.Equal(t, "20210706MMQFMPU8000001", record.CustomerReferenceNumber)
	require.Equal(t, "Outgoing Wire Return", record.Text)
	require.Equal(t, 73, size)

	result := record.string()
	// NB. This is the current output of a Detail formatted like the above.
	expectResult := `16,266,1912,,GI2118700002010,20210706MMQFMPU8000001,Outgoing Wire Return/`
	require.Equal(t, expectResult, result)
	require.Equal(t, len(expectResult), len(result))
}

/**
 * This test outlines the behavior of a Detail record when the Detail includes Identifier and Text records
 * that include a Terminal character ('/'). The BAI2 spec indicates that most fields "must not contain a comma or a slash"
 * and places strict rules on when/where a slash can occur. Nevertheless, real world data is observed where a slash
 * is used in records that disallow it.
 *
 * This library currently raises an error when parsing a record that includes slash characters.
 */
func TestTransactionDetailOutput_FailsIfRecordIncludesIllegalCharacters(t *testing.T) {
	data := `16,447,928650,,SPB2322684598521,AB/GS/RPFILERP0001/RPBA0001,ACH Credit Payment,Entry Description: TRADE; -, SEC: CTX, Client Ref ID: AB/GS/TEST0001/RPBA0001, GS ID: SPB2322684598521/
88,EREF: AB/GS/RPFILERP0001/RPBA0001/
88,DBNM: SAMPLE INC/
88,CACT: ACHCONTROLOUTUSD01/`

	record := transactionDetail{}

	size, err := record.parse(data)
	require.Equal(t, fmt.Errorf("TransactionDetail: unable to parse Text"), err)
	require.Equal(t, size, 0)

	// require.Equal(t, "447", record.TypeCode)
	// require.Equal(t, "928650", record.Amount)
	// require.Equal(t, "", string(record.FundsType.TypeCode))
	// require.Equal(t, "", record.FundsType.Date)
	// require.Equal(t, "", record.FundsType.Time)
	// require.Equal(t, "SPB2322684598521", record.BankReferenceNumber)
	// require.Equal(t, "AB/GS/RPFILERP0001/RPBA0001", record.CustomerReferenceNumber)
	// require.Equal(t, "ACH Credit Payment", record.Text)
	// require.Equal(t, 79, size)

	// result := record.string()
	// // NB. This is the current output of a Detail formatted like the above.
	// expectResult := `16,447,928650,,SPB2322684598521,AB/GS/RPFILERP0001/RPBA0001,ACH Credit Payment/`
	// require.Equal(t, expectResult, result)
}

/**
 * This test outlines the behavior of a Detail record when the Detail and Continuations for the detail are terminated
 * by a newline character ("\n") rather than a slash ("/").
 *
 * This library on `master` raises an error when parsing a detail that is newline terminated. This branch includes a
 * proposed change to `util.getIndex` and `util.GetSize` that may support newline termination without regression.
 */
func TestTransactionDetailOutput_FailsIfLinesAreNotTerminated(t *testing.T) {
	data := `16,266,1912,,GI2118700002010,20210706MMQFMPU8000001,Outgoing Wire Return,-
88,CREF: 20210706MMQFMPU8000001
88,EREF: 20210706MMQFMPU8000001
88,DBIC: GSCRUS33
88,CRNM: ABC Company
88,DBNM: SAMPLE INC.`

	record := transactionDetail{}

	// size, err := record.parse(data)

	// require.Equal(t, fmt.Errorf("TransactionDetail: unable to parse record"), err)
	// require.Equal(t, size, 0)

	size, err := record.parse(data)
	require.NoError(t, err)

	require.Equal(t, "266", record.TypeCode)
	require.Equal(t, "1912", record.Amount)
	require.Equal(t, "", string(record.FundsType.TypeCode))
	require.Equal(t, "", record.FundsType.Date)
	require.Equal(t, "", record.FundsType.Time)
	require.Equal(t, "GI2118700002010", record.BankReferenceNumber)
	require.Equal(t, "20210706MMQFMPU8000001", record.CustomerReferenceNumber)
	require.Equal(t, "Outgoing Wire Return", record.Text)
	require.Equal(t, 73, size)

	result := record.string()
	// NB. This is the current output of a Detail formatted like the above.
	expectResult := `16,266,1912,,GI2118700002010,20210706MMQFMPU8000001,Outgoing Wire Return/`
	require.Equal(t, expectResult, result)
}

