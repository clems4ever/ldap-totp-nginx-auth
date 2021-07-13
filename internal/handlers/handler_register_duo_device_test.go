package handlers

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/mocks"
)

type RegisterDuoDeviceSuite struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *RegisterDuoDeviceSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	s.mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}

func (s *RegisterDuoDeviceSuite) TearDownTest() {
	s.mock.Close()
}

func (s *RegisterDuoDeviceSuite) TestShouldCallDuoAPIAndFail() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	values := url.Values{}
	values.Set("username", "john")

	duoMock.EXPECT().PreauthCall(gomock.Eq(values), s.mock.Ctx).Return(nil, fmt.Errorf("Connnection error"))

	SecondFactorDuoDevicesGet(duoMock)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Duo PreAuth API errored: Connnection error", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *RegisterDuoDeviceSuite) TestShouldRespondWithDummyOnBypass() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	values := url.Values{}
	values.Set("username", "john")

	duoresponse := duo.PreauthResponse{}
	duoresponse.Result = testResultAllow

	duoMock.EXPECT().PreauthCall(gomock.Eq(values), s.mock.Ctx).Return(&duoresponse, nil)

	SecondFactorDuoDevicesGet(duoMock)(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), DuoDevicesResponse{Result: testResultAllow, Devices: nil})
}

func (s *RegisterDuoDeviceSuite) TestShouldRespondOK() {
	s.mock.Ctx.Request.SetBodyString("{\"device\":\"1234567890123456\", \"method\":\"push\"}")
	s.mock.StorageProviderMock.EXPECT().
		SavePreferredDuoDevice(gomock.Eq("john"), gomock.Eq("1234567890123456"), gomock.Eq("push")).
		Return(nil)

	SecondFactorDuoDevicePost(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
}

func (s *RegisterDuoDeviceSuite) TestShouldRespondKOOnInvalidMethod() {
	s.mock.Ctx.Request.SetBodyString("{\"device\":\"1234567890123456\", \"method\":\"testfailure\"}")

	SecondFactorDuoDevicePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *RegisterDuoDeviceSuite) TestShouldRespondKOOnEmptyMethod() {
	s.mock.Ctx.Request.SetBodyString("{\"device\":\"1234567890123456\", \"method\":\"\"}")

	SecondFactorDuoDevicePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to validate body: method: non zero value required", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *RegisterDuoDeviceSuite) TestShouldRespondKOOnEmptyDevice() {
	s.mock.Ctx.Request.SetBodyString("{\"device\":\"\", \"method\":\"push\"}")

	SecondFactorDuoDevicePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to validate body: device: non zero value required", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func TestRunRegisterDuoDeviceSuite(t *testing.T) {
	s := new(RegisterDuoDeviceSuite)
	suite.Run(t, s)
}
