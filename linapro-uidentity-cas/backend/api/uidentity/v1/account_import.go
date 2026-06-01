// This file declares legacy-compatible account import endpoint DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// AccountImportCheckReq validates an account import workbook before execution.
type AccountImportCheckReq struct {
	g.Meta   `path:"/uidentity/accounts/import-checks" method:"post" tags:"UIdentity Account Import" summary:"Validate account import workbook" dc:"Open an uploaded account import workbook, validate the Sheet1 header and row bounds, and return the number of importable rows before any data is written." permission:"uidentity:cas:write"`
	Filepath string `json:"filepath" v:"required" dc:"Server-side workbook path produced by the upload flow" eg:"/tmp/account_import.xlsx"`
	Limit    int    `json:"limit" dc:"Optional maximum row count; defaults to the plugin import limit when omitted" eg:"1000"`
}

// AccountImportCheckRes carries workbook validation metadata.
type AccountImportCheckRes struct {
	Rows int `json:"rows" dc:"Number of non-empty import rows detected in Sheet1" eg:"10"`
}

// AccountImportReq imports account rows from one workbook.
type AccountImportReq struct {
	g.Meta   `path:"/uidentity/accounts/imports" method:"post" tags:"UIdentity Account Import" summary:"Import accounts from workbook" dc:"Import or update plugin accounts and account details from the Sheet1 rows of a server-side workbook. Existing accounts are matched by account number and updated with non-empty workbook fields." permission:"uidentity:cas:write"`
	Filepath string `json:"filepath" v:"required" dc:"Server-side workbook path produced by the upload flow" eg:"/tmp/account_import.xlsx"`
	Limit    int    `json:"limit" dc:"Optional maximum row count; defaults to the plugin import limit when omitted" eg:"1000"`
}

// AccountImportRes carries account import results.
type AccountImportRes struct {
	Success      int      `json:"success" dc:"Number of rows successfully imported or updated" eg:"8"`
	FailedNumber []string `json:"failedNumber" dc:"Account numbers that failed to import or update" eg:"[\"A001\"]"`
}
