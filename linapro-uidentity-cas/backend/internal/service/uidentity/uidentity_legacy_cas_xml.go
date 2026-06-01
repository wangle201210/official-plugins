// This file renders legacy CAS serviceResponse XML documents while keeping
// ticket validation and account access checks in the existing runtime service.

package uidentity

import (
	"context"
	"encoding/xml"
)

const (
	legacyCASXMLNamespace       = "http://www.yale.edu/tp/cas"
	legacyCASFailureCodeInvalid = "INVALID_TICKET"
	legacyCASUserTypeStaff      = "01"
	legacyCASUserTypeExternal   = "02"
	legacyCASStaffContainerID   = int64(2)
)

// LegacyCASServiceXML validates a service ticket and renders old CAS XML.
func (s *serviceImpl) LegacyCASServiceXML(ctx context.Context, in LegacyCASServiceXMLInput) (*LegacyCASServiceXMLOutput, error) {
	out, err := s.ValidateServiceTicket(ctx, ServiceValidateInput{
		Ticket: in.Ticket,
		UserID: in.UserID,
	})
	if err != nil {
		return buildLegacyCASFailureXML(err.Error())
	}
	return buildLegacyCASSuccessXML(out.User)
}

func buildLegacyCASSuccessXML(account *RuntimeAccount) (*LegacyCASServiceXMLOutput, error) {
	if account == nil {
		return buildLegacyCASFailureXML("Runtime account is missing")
	}
	var (
		birthday string
		email    string
		sex      int
		idcard   string
	)
	if account.Detail != nil {
		birthday = account.Detail.Birthday
		email = account.Detail.Email
		sex = account.Detail.Gender
		idcard = account.Detail.Idcard
	}
	userType := legacyCASUserTypeExternal
	if account.ContainerID == legacyCASStaffContainerID {
		userType = legacyCASUserTypeStaff
	}
	return marshalLegacyCASXML(legacyCASServiceResponse{
		Xmlns: legacyCASXMLNamespace,
		AuthenticationSuccess: &legacyCASAuthenticationSuccess{
			User: account.Number,
			Attributes: legacyCASAttributes{
				Birthday:       birthday,
				LoginID:        account.Number,
				WorkCode:       account.Number,
				Sex:            int64(sex),
				DepartmentID:   int(account.UnitID),
				Mobile:         account.Phone,
				Telephone:      account.Phone,
				ID:             int(account.ID),
				CertificateNum: idcard,
				Email:          email,
				Status:         account.Status,
				Name:           account.Name,
				UserType:       userType,
			},
		},
	})
}

func buildLegacyCASFailureXML(message string) (*LegacyCASServiceXMLOutput, error) {
	return marshalLegacyCASXML(legacyCASServiceResponse{
		Xmlns: legacyCASXMLNamespace,
		AuthenticationFailure: &legacyCASAuthenticationFailure{
			Code:    legacyCASFailureCodeInvalid,
			Message: "CAS ticket validation failed: " + message,
		},
	})
}

func marshalLegacyCASXML(response legacyCASServiceResponse) (*LegacyCASServiceXMLOutput, error) {
	data, err := xml.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, err
	}
	return &LegacyCASServiceXMLOutput{XML: append([]byte(xml.Header), data...)}, nil
}

type legacyCASServiceResponse struct {
	XMLName               xml.Name                        `xml:"cas:serviceResponse"`
	Xmlns                 string                          `xml:"xmlns:cas,attr"`
	AuthenticationSuccess *legacyCASAuthenticationSuccess `xml:"cas:authenticationSuccess,omitempty"`
	AuthenticationFailure *legacyCASAuthenticationFailure `xml:"cas:authenticationFailure,omitempty"`
}

type legacyCASAuthenticationSuccess struct {
	User       string              `xml:"cas:user"`
	Attributes legacyCASAttributes `xml:"cas:attributes"`
}

type legacyCASAuthenticationFailure struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}

type legacyCASAttributes struct {
	Birthday       string `xml:"cas:birthday"`
	SubCompanyID   int    `xml:"cas:subcompanyid"`
	LoginID        string `xml:"cas:loginid"`
	WorkCode       string `xml:"cas:workcode"`
	Sex            int64  `xml:"cas:sex"`
	DepartmentID   int    `xml:"cas:departmentid"`
	Mobile         string `xml:"cas:mobile"`
	SystemLanguage int    `xml:"cas:systemlanguage"`
	Telephone      string `xml:"cas:telephone"`
	ManagerID      int    `xml:"cas:managerid"`
	CountryID      int    `xml:"cas:countryid"`
	AssistantID    int    `xml:"cas:assistantid"`
	ID             int    `xml:"cas:id"`
	CertificateNum string `xml:"cas:certificatenum"`
	Email          string `xml:"cas:email"`
	Status         int    `xml:"cas:status"`
	Name           string `xml:"cas:name"`
	UserType       string `xml:"cas:userType"`
}
