package duo

// Duo Methods.
const (
	// Push Method - The device is activated for Duo Push.
	Push = "push"
	// OTP Method - The device is capable of generating passcodes with the Duo Mobile app.
	OTP = "mobile_otp"
	// Phone Method - The device can receive phone calls.
	Phone = "phone"
	// SMS Method - The device can receive batches of SMS passcodes.
	SMS = "sms"
)

// PossibleMethods is the set of all possible 2FDuoA methods.
var PossibleMethods = []string{Push} // MOTP, Phone, SMS
